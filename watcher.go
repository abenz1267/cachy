package cachy

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Watch is used to monitor file changes and update the template cache.
// You can set debug = true if you want Cachy to output log entries on an event.
func (c *Cachy) Watch(debug bool) error {
	c.debug = debug

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	c.addFolders(watcher)

	done := make(chan bool)

	go c.watch(watcher)

	if len(c.folders) == 0 {
		c.log("nothing to watch, closing watcher")
		done <- true
	}
	c.log("watching templates for changes...")

	<-done
	return nil
}

func (c *Cachy) watch(w *fsnotify.Watcher) {
	for {
		select {
		case event := <-w.Events:
			if strings.HasSuffix(event.Name, c.ext) {
				clearPath := strings.TrimPrefix(strings.TrimSuffix(event.Name, c.ext), c.wDir+"/")

				if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
					if err := c.updateTmpl(clearPath); err != nil && c.debug {
						if err.Error() != errAlreadyUpdated {
							c.log(fmt.Sprintf("couldn't cache template %s. %s", clearPath, err))
						}
					} else if c.debug {
						c.log(fmt.Sprintf("update template %s", clearPath))
					}
				} else if event.Op == fsnotify.Remove || event.Op == fsnotify.Rename {
					deleteTmpl(clearPath, c)
				}
			}

		case err := <-w.Errors:
			c.log(err.Error())
		}
	}
}

func (c *Cachy) addFolders(w *fsnotify.Watcher) {
	counter := 0
	for _, v := range c.folders {
		v = filepath.Join(c.wDir, v)
		if err := w.Add(v); err != nil {
			c.log(fmt.Sprintf("Cachy: %s:%s", err, v))
			counter++
		}
	}
}

func (c *Cachy) log(msg string) {
	if c.debug {
		log.Printf("Cachy: %s", msg)
	}
}

func (c *Cachy) updateTmpl(path string) (err error) {
	pathParts := strings.Split(path, "/")
	file := pathParts[len(pathParts)-1] + c.ext

	var templatepath string
	if c.allowDuplicates {
		templatepath = filepath.Join(pathParts[:len(pathParts)-1]...)
	} else {
		templatepath = findFile(c, file)
	}

	length, err := c.cache(templatepath, file, true)
	if err != nil {
		return err
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
		c.SSE.Notifier <- []byte("updated")
	}
	return
}

func findFile(c *Cachy, file string) string {
	var realpath string

	for _, v := range c.folders {
		if v == "" {
			return realpath
		}

		err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == file {
				realpath = v

				return nil
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	return realpath
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
