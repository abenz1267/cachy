package cachy

import (
	"html/template"
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func (c *Cachy) Watch(wsURL string) {
	if wsURL != "" {
		c.funcs["ws"] = func() template.HTML {
			src := `<script>
		var ws = new WebSocket('` + wsURL + `');
		ws.onclose = () => {
		  location.reload(true);
		};
	  </script>`
			return template.HTML(src)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if strings.Contains(event.Name, c.ext) {
					clearPath := strings.TrimPrefix(strings.TrimSuffix(event.Name, c.ext), wDir+"/")

					if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
						if err := c.updateTmpl(clearPath); err != nil {
							log.Printf("Cachy: couldn't cache template: %s", err)
						} else {
							log.Printf("Cachy: updated template file: %s\n", clearPath)
						}
					} else if event.Op == fsnotify.Remove || event.Op == fsnotify.Rename {
						deleteTmpl(clearPath, c)
					}
				}

			case err := <-watcher.Errors:
				log.Println(err)
			}
		}
	}()

	counter := 0
	for _, v := range c.folders {
		v = filepath.Join(wDir, v)
		if err := watcher.Add(v); err != nil {
			log.Printf("Cachy: %s:%s", err, v)
			counter++
		}
	}

	if counter == len(c.folders) {
		log.Println("Cachy: nothing to watch, closing watcher")
		done <- true
	}

	log.Println("Cachy: Watching templates for changes...")

	<-done
}

func (c *Cachy) updateTmpl(path string) (err error) {
	pathParts := strings.Split(path, "/")
	length, err := c.cache(filepath.Join(pathParts[:len(pathParts)-1]...), pathParts[len(pathParts)-1]+c.ext, nil)
	if err != nil {
		return
	}

	for k := range c.multiTmpls {
		if strings.Contains(k, path) {
			files := strings.Split(k, ",")
			c.multiTmpls[k], err = c.parseMultiple(files)
			if err != nil {
				return
			}
		}
	}

	if length > 0 {
		c.wsChan <- true
	}
	return
}

func deleteTmpl(clearPath string, c *Cachy) {
	if _, exists := c.stringTemplates[clearPath]; exists {
		log.Printf("Cachy: deleting template from cache: %s\n", clearPath)
		delete(c.stringTemplates, clearPath)
		delete(c.templates, clearPath)
	}

	for k := range c.multiTmpls {
		if strings.Contains(k, clearPath) {
			delete(c.multiTmpls, k)
		}
	}
}
