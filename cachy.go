package cachy

import (
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
	folders         []string
	ext             string
	reloadChan      chan bool
	reload          bool
	debug           bool
	wDir            string
}

// New processes all templates and returns a populated Cachy struct.
// You can provide template folders, otherwise it will scan the whole working dir for templates.
func New(tmplExt string, funcs template.FuncMap, folders ...string) (c *Cachy, err error) {
	c = &Cachy{}
	c.templates = make(map[string]*template.Template)
	c.multiTmpls = make(map[string]*template.Template)
	c.stringTemplates = make(map[string]string)
	c.ext = tmplExt
	c.reloadChan = make(chan bool)
	c.funcs = template.FuncMap{}
	c.funcs["reloadURL"] = func() string {
		return "no hot-reloading URL defined"
	}

	for k, v := range funcs {
		if _, exists := c.funcs[k]; exists {
			return nil, fmt.Errorf("Cachy: function '%s' already exists!", k)
		} else {
			c.funcs[k] = v
		}
	}

	c.wDir, err = os.Getwd()
	if err != nil {
		return
	}

	if len(folders) == 0 {
		folders, err = walkDir(c.wDir)
		if err != nil {
			return
		}
	}

	c.folders = folders

	return c, c.load()
}

// Execute executes the given template(s).
func (c *Cachy) Execute(w io.Writer, data interface{}, files ...string) (err error) {
	if v, ok := w.(http.ResponseWriter); ok {
		v.Header().Set("Content-Type", "text/html")
	}

	switch len := len(files); {
	case len == 0:
		return errors.New("Cachy: there are no templates to execute")
	case len == 1:
		return c.templates[files[0]].Execute(w, data)
	case len > 1:
		return c.executeMultiple(w, data, files)
	}

	return
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
				if _, err := c.cache(k, file.Name()); err != nil {
					return err
				}
			}
		}
	}

	return
}

func (c *Cachy) cache(path, file string) (length int, err error) {
	var tmpl *template.Template
	var clearPath string
	var tmplBytes []byte

	clearPath = filepath.Join(strings.TrimPrefix(path, "/"), strings.TrimSuffix(file, c.ext))
	tmplBytes, err = ioutil.ReadFile(filepath.Join(c.wDir, path, file))
	if err != nil {
		return len(tmplBytes), err
	}

	c.stringTemplates[clearPath] = string(tmplBytes)

	tmpl, err = template.New(file).Funcs(c.funcs).Parse(c.stringTemplates[clearPath])
	if err != nil {
		return len(tmplBytes), err
	}

	c.templates[clearPath] = tmpl

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

func (c *Cachy) HotReload(w http.ResponseWriter, r *http.Request) {
	<-c.reloadChan
	w.WriteHeader(http.StatusOK)
}
