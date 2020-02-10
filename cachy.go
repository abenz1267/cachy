// Package cachy is a package that can be used to cache and execute templates.
package cachy

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Cachy represents the template cache
type Cachy struct {
	funcs           template.FuncMap
	templates       map[string]*template.Template
	multiTmpls      map[string]*template.Template
	stringTemplates map[string]string
	checksums       map[string][]byte
	folders         []string
	ext             string
	debug           bool
	allowDuplicates bool
	recursive       bool
	wDir            string
	reloadURL       string
	SSE             *broker
}

// Params can be used to modify Cachy's default behaviour.
type Params struct {
	URL        string
	Ext        string
	Duplicates bool
	Recursive  bool
}

const errAlreadyUpdated = "already updated"

// New processes all templates and returns a populated Cachy struct.
// You can provide template folders, otherwise it will scan the whole working dir for templates.
func New(p *Params, funcs template.FuncMap, folders ...string) (c *Cachy, err error) {
	c = &Cachy{}
	c.bindBroker()

	if p != nil {
		c.allowDuplicates = p.Duplicates
		c.recursive = p.Recursive
		c.ext = "." + p.Ext
		c.reloadURL = p.URL

		if strings.HasPrefix(p.Ext, ".") {
			return nil, errors.New("extension can't start with a '.'")
		}
	} else {
		c.ext = ".html"
		c.reloadURL = "/hotreload"
	}

	c.templates = make(map[string]*template.Template)
	c.multiTmpls = make(map[string]*template.Template)
	c.stringTemplates = make(map[string]string)
	c.checksums = make(map[string][]byte)
	c.funcs = template.FuncMap{}

	if c.reloadURL != "" {
		c.funcs["reloadScript"] = func() template.HTML {
			src := `<script>
  var client = new EventSource("` + c.reloadURL + `")
  client.onmessage = function (msg) {
    location.reload();
  }
</script>
`
			return template.HTML(src)
		}
	}

	for k, v := range funcs {
		if _, exists := c.funcs[k]; exists {
			return nil, fmt.Errorf("cachy: function '%s' already exists", k)
		}
		c.funcs[k] = v
	}

	c.wDir, err = os.Getwd()
	if err != nil {
		return
	}

	var noExplFolders bool
	if len(folders) == 0 {
		folders, err = walkDir(c.wDir)
		if err != nil {
			return
		}

		noExplFolders = true
	}

	if c.recursive && !noExplFolders {
		var toAdd []string
		for _, v := range folders {
			toAdd, err = walkDir(v)
			if err != nil {
				return
			}

			for k, s := range toAdd {
				toAdd[k] = filepath.Join(v, s)
			}

			c.folders = append(c.folders, toAdd...)
		}

		c.folders = uniquePaths(c.folders)
	} else {
		c.folders = folders
	}

	return c, c.load()
}

// Execute executes the given template(s).
func (c *Cachy) Execute(w io.Writer, data interface{}, files ...string) (err error) {
	if v, ok := w.(http.ResponseWriter); ok {
		v.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	switch l := len(files); {
	case l == 0:
		return errors.New("Cachy: there are no templates to execute")
	case l == 1:
		return c.templates[files[0]].Execute(w, data)
	default:
		return c.executeMultiple(w, data, files)
	}
}

// GetString returns the string representation of the given template.
func (c *Cachy) GetString(file string) string {
	return c.stringTemplates[file]
}

func (c *Cachy) executeMultiple(w io.Writer, data interface{}, files []string) (err error) {
	templates := strings.Join(files, ",")

	if val, exists := c.multiTmpls[templates]; exists {
		return val.Execute(w, data)
	}

	c.multiTmpls[templates], err = c.parseMultiple(files)
	if err != nil {
		return
	}

	return c.multiTmpls[templates].Execute(w, data)
}

func (c *Cachy) parseMultiple(files []string) (tmpl *template.Template, err error) {
	tmpl = template.New("tmpl").Funcs(c.funcs)

	for _, v := range files {
		if val, exists := c.stringTemplates[v]; exists {
			_, err = tmpl.Parse(val)
		} else {
			return nil, fmt.Errorf("Cachy: there is no template '%s'", v)
		}
	}

	return
}

func (c *Cachy) load() (err error) {
	dirs := make(map[string][]os.FileInfo)
	for _, v := range c.folders {
		files, err := ioutil.ReadDir(filepath.Join(c.wDir, v))
		if err != nil {
			return err
		}

		dirs[v] = files
	}

	for k, v := range dirs {
		for _, file := range v {
			if !file.IsDir() && strings.HasSuffix(file.Name(), c.ext) {
				if _, err := c.cache(k, file.Name(), false); err != nil {
					return err
				}
			}
		}
	}

	return
}

func (c *Cachy) cache(path, file string, update bool) (length int, err error) {
	var tmpl *template.Template
	var clearPath string
	var tmplBytes []byte

	h := md5.New()

	if c.allowDuplicates {
		clearPath = filepath.Join(strings.TrimPrefix(path, "/"), strings.TrimSuffix(file, c.ext))
	} else {
		clearPath = strings.TrimSuffix(file, c.ext)
	}

	if !update {
		if _, exists := c.templates[clearPath]; exists {
			return len(tmplBytes), fmt.Errorf("template '%s' already exists", clearPath)
		}
	}

	tmplBytes, err = ioutil.ReadFile(filepath.Join(c.wDir, path, file))
	if err != nil {
		return len(tmplBytes), err
	}

	checksum := h.Sum(tmplBytes)
	if bytes.Equal(c.checksums[clearPath], checksum) {
		return 0, errors.New(errAlreadyUpdated)
	}

	c.checksums[clearPath] = checksum

	c.stringTemplates[clearPath] = string(tmplBytes)

	tmpl, err = template.New(file).Funcs(c.funcs).Parse(c.stringTemplates[clearPath])
	if err != nil {
		return len(tmplBytes), err
	}

	c.templates[clearPath] = tmpl

	for k := range c.multiTmpls {
		templates := strings.Split(k, ",")

		for _, v := range templates {
			if v == clearPath {
				delete(c.multiTmpls, k)
			}
		}
	}

	return len(tmplBytes), err
}

func walkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !strings.Contains(path, "/.") {
			if !strings.Contains(info.Name(), ".") {
				files = append(files, strings.TrimPrefix(path, root))
			}
		}
		return nil
	})

	return files, err
}

// URL returns the hot-reload URL
func (c Cachy) URL() string {
	return c.reloadURL
}

func uniquePaths(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}
