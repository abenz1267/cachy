package cachy

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Watch is used to monitor file changes and update the template cache.
// You can set debug = true if you want Cachy to ouput log entries on an event.
func (c *Cachy) Watch(debug bool) error {
	c.debug = debug

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if strings.Contains(event.Name, c.ext) {
					clearPath := strings.TrimPrefix(strings.TrimSuffix(event.Name, c.ext), c.wDir+"/")

					if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
						if err := c.updateTmpl(clearPath); err != nil && debug {
							c.log(fmt.Sprintf("couldn't cache template %s", err))
						} else if debug {
							c.log(fmt.Sprintf("update template %s", clearPath))
						}
					} else if event.Op == fsnotify.Remove || event.Op == fsnotify.Rename {
						deleteTmpl(clearPath, c)
					}
				}

			case err := <-watcher.Errors:
				c.log(err.Error())
			}
		}
	}()

	counter := 0
	for _, v := range c.folders {
		v = filepath.Join(c.wDir, v)
		if err := watcher.Add(v); err != nil {
			c.log(fmt.Sprintf("Cachy: %s:%s", err, v))
			counter++
		}
	}

	if counter == len(c.folders) {
		c.log("nothing to watch, closing watcher")
		done <- true
	}
	c.log("watching templates for changes...")

	<-done
	return nil
}

func (c *Cachy) log(msg string) {
	if c.debug {
		log.Printf("Cachy: %s", msg)
	}
}

func (c *Cachy) updateTmpl(path string) (err error) {
	pathParts := strings.Split(path, "/")
	length, err := c.cache(filepath.Join(pathParts[:len(pathParts)-1]...), pathParts[len(pathParts)-1]+c.ext)
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

	if length > 0 && c.reloadChan != nil {
		c.reloadChan <- true
	}
	return
}

func deleteTmpl(clearPath string, c *Cachy) {
	if _, exists := c.stringTemplates[clearPath]; exists {
		c.log(fmt.Sprintf("deleting template from cache: %s\n", clearPath))
		delete(c.stringTemplates, clearPath)
		delete(c.templates, clearPath)
	}

	for k := range c.multiTmpls {
		if strings.Contains(k, clearPath) {
			delete(c.multiTmpls, k)
		}
	}
}
