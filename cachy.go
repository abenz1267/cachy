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
func Init(tmplExt string, enableWatcher bool, funcs template.FuncMap, folders ...string) (c Cachy, err error) {
	c.templates = make(map[string]*template.Template)
	c.multiTmpls = make(map[string]*template.Template)
	c.stringTemplates = make(map[string]string)
	c.funcs = funcs

	wDir, err = os.Getwd()
	if err != nil {
		return
	}

	if len(folders) == 0 {
		log.Println("Cachy: no folders specified, walking whole directory...")
		folders, err = walkDir(wDir)
		if err != nil {
			return
		}
	}

	err = load(tmplExt, &c, folders)

	if enableWatcher {
		go watch(folders, tmplExt, &c)
	}

	return
}

// Execute executes the given template(s).
func (c *Cachy) Execute(w io.Writer, data interface{}, files ...string) (err error) {
	if len(files) == 0 {
		return errors.New("Cachy: there are no templates to execute")
	}

	if len(files) == 1 {
		return c.templates[files[0]].Execute(w, data)
	}

	templates := strings.Join(files, ",")

	if val, exists := c.multiTmpls[templates]; exists {
		return val.Execute(w, data)
	}

	c.multiTmpls[templates], err = parseMultiple(c, files)
	if err != nil {
		return
	}

	return c.multiTmpls[templates].Execute(w, data)
}

func parseMultiple(c *Cachy, files []string) (tmpl *template.Template, err error) {
	tmpl = template.New("tmpl").Funcs(c.funcs)

	for _, v := range files {
		if val, exists := c.stringTemplates[v]; exists {
			_, err = tmpl.Parse(val)
		} else {
			log.Fatalf("Cachy: there is no template '%s'", v)
		}
	}

	return
}

func load(tmplExt string, c *Cachy, folders []string) (err error) {
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
				if err := cache(c, k, file.Name(), tmplExt); err != nil {
					return err
				}
			}
		}
	}

	return
}

func cache(c *Cachy, path, file, tmplExt string) (err error) {
	// parse template and cache it
	tmpl, err := template.New(file).Funcs(c.funcs).ParseFiles(filepath.Join(wDir, path, file))
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
		if info.IsDir() && !strings.Contains(path, "/.") {
			if !strings.Contains(info.Name(), ".") {
				files = append(files, strings.TrimPrefix(path, root))
			}
		}
		return nil
	})

	return files, err
}
