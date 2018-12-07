package cachy

import (
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var wDir string

// Cachy represents the template cache
type Cachy struct {
	funcs           template.FuncMap
	templates       map[string]*template.Template
	multiTmpls      map[string]*template.Template
	stringTemplates map[string]string
}

// Init processes all templates and returns a populated Cachy struct.
// You can provide template folders, otherwise it will scan the whole working dir for templates.
func Init(tmplExt string, watch bool, funcs template.FuncMap, folders ...string) (c Cachy, err error) {
	c.templates = make(map[string]*template.Template)
	c.multiTmpls = make(map[string]*template.Template)
	c.stringTemplates = make(map[string]string)
	c.funcs = funcs

	wDir, err = os.Getwd()
	if err != nil {
		return
	}

	if len(folders) == 0 {
		folders, err = walkDir(wDir)
		if err != nil {
			return
		}
	}

	err = load(tmplExt, watch, &c, funcs, folders)

	return
}

// Execute executes the given template(s).
func (c Cachy) Execute(w io.Writer, data interface{}, files ...string) (err error) {
	if len(files) == 0 {
		return errors.New("there are no templates to execute")
	}

	if len(files) == 1 {
		c.templates[files[0]].Execute(w, data)
		return
	}

	templates := strings.Join(files, ",")

	if val, exists := c.multiTmpls[templates]; exists {
		return val.Execute(w, data)
	}

	tmpl := template.New("tmpl").Funcs(c.funcs)

	for _, v := range files {
		if val, exists := c.stringTemplates[v]; exists {
			tmpl.Parse(val)
		} else {
			log.Fatalf("There is no template '%s'", v)
		}
	}

	c.multiTmpls[templates] = tmpl

	tmpl.Execute(w, data)

	return
}

func load(tmplExt string, watch bool, c *Cachy, funcs template.FuncMap, folders []string) (err error) {
	dirs := make(map[string][]os.FileInfo)
	for _, v := range folders {
		files, err := ioutil.ReadDir(filepath.Join(wDir, v))
		if err != nil {
			return err
		}

		dirs[v] = files
	}

	for k, v := range dirs {
		for _, file := range v {
			if !file.IsDir() && strings.HasSuffix(file.Name(), tmplExt) {
				if err := cache(c, funcs, k, file.Name(), tmplExt); err != nil {
					return err
				}
			}
		}
	}

	return
}

func cache(c *Cachy, funcs template.FuncMap, path, file, tmplExt string) (err error) {
	// parse template and cache it
	tmpl, err := template.New(file).Funcs(funcs).ParseFiles(filepath.Join(wDir, path, file))
	if err != nil {
		return err
	}

	clearPath := filepath.Join(strings.TrimPrefix(path, "/"), strings.TrimSuffix(file, tmplExt))

	c.templates[clearPath] = tmpl

	// parse string representation of template and cache it
	b, err := ioutil.ReadFile(filepath.Join(wDir, path, file))
	if err != nil {
		return err
	}

	c.stringTemplates[clearPath] = string(b)

	return
}

func walkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if !strings.Contains(info.Name(), ".") {
				files = append(files, strings.TrimPrefix(path, root))
			}
		}
		return nil
	})
	return files, err
}
