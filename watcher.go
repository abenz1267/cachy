package cachy

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func watch(folders []string, ext string, c *Cachy, isPackr bool) {
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
				if strings.Contains(event.Name, ext) {
					clearPath := strings.TrimPrefix(strings.TrimSuffix(event.Name, ext), wDir+"/")

					if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
						if err := updateTmpl(clearPath, ext, c); err != nil {
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
	for _, v := range folders {
		v = filepath.Join(wDir, v)
		if err := watcher.Add(v); err != nil {
			if !isPackr {
				log.Fatalf("Cachy: %s:%s", err, v)
			} else {
				log.Printf("Cachy: %s:%s", err, v)
				counter++
			}
		}
	}

	if counter == len(folders) {
		log.Println("Cachy: nothing to watch, closing watcher")
		done <- true
	}

	log.Println("Cachy: Watching templates for changes...")

	<-done
}

func updateTmpl(path, ext string, c *Cachy) (err error) {
	pathParts := strings.Split(path, "/")
	if err = cache(c, filepath.Join(pathParts[:len(pathParts)-1]...), pathParts[len(pathParts)-1]+ext, ext, nil); err != nil {
		return
	}

	for k := range c.multiTmpls {
		if strings.Contains(k, path) {
			files := strings.Split(k, ",")
			c.multiTmpls[k], err = parseMultiple(c, files)
			if err != nil {
				return
			}
		}
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
