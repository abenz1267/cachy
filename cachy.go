package cachy

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/packr/v2"
)

var wDir string

// Cachy represents the template cache
type Cachy struct {
	funcs           template.FuncMap
	templates       map[string]*template.Template
	multiTmpls      map[string]*template.Template
	stringTemplates map[string]string
	folders         []string
	ext             string
}

// New processes all templates and returns a populated Cachy struct.
// You can provide template folders, otherwise it will scan the whole working dir for templates.
func New(tmplExt string, enableWatcher bool, funcs template.FuncMap, boxes map[string]*packr.Box, folders ...string) (c Cachy, err error) {
	c.templates = make(map[string]*template.Template)
	c.multiTmpls = make(map[string]*template.Template)
	c.stringTemplates = make(map[string]string)
	c.funcs = funcs
	c.ext = tmplExt

	wDir, err = os.Getwd()
	if err != nil {
		return
	}

	// set folders
	switch {
	case len(folders) == 0 && boxes == nil:
		folders, err = walkDir(wDir)
		if err != nil {
			return
		}
	case boxes != nil:
		for k := range boxes {
			folders = append(folders, k)
		}
	}

	c.folders = folders

	// cache templates
	switch boxes {
	case nil:
		return c, c.load()
	default:
		return c, c.loadBoxes(boxes)
	}
}

// Execute executes the given template(s).
func (c *Cachy) Execute(w io.Writer, data interface{}, files ...string) (err error) {
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
			return nil, errors.New(fmt.Sprintf("Cachy: there is no template '%s'", v))
		}
	}

	return
}

func (c *Cachy) load() (err error) {
	dirs := make(map[string][]os.FileInfo)
	for _, v := range c.folders {
		files, err := ioutil.ReadDir(filepath.Join(wDir, v))
		if err != nil {
			return err
		}

		dirs[v] = files
	}

	for k, v := range dirs {
		for _, file := range v {
			if !file.IsDir() && strings.HasSuffix(file.Name(), c.ext) {
				if err := c.cache(k, file.Name(), nil); err != nil {
					return err
				}
			}
		}
	}

	return
}

func (c *Cachy) loadBoxes(boxes map[string]*packr.Box) (err error) {
	for _, v := range boxes {
		for _, f := range v.List() {
			err = c.cache("", f, v)
			if err != nil {
				return err
			}
		}
	}

	return
}

func (c *Cachy) cache(path, file string, box *packr.Box) (err error) {
	var tmpl *template.Template
	var clearPath string
	var tmplBytes []byte

	if box == nil {
		clearPath = filepath.Join(strings.TrimPrefix(path, "/"), strings.TrimSuffix(file, c.ext))
		tmplBytes, err = ioutil.ReadFile(filepath.Join(wDir, path, file))
		if err != nil {
			return err
		}
	} else {
		clearPath = filepath.Join(box.Name, strings.TrimSuffix(file, c.ext))
		tmplBytes, err = box.Find(file)
		if err != nil {
			return err
		}
	}

	c.stringTemplates[clearPath] = string(tmplBytes)

	tmpl, err = template.New(file).Funcs(c.funcs).Parse(c.stringTemplates[clearPath])
	if err != nil {
		return err
	}

	c.templates[clearPath] = tmpl

	return
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
