package main

import (
	"log"
	"net/http"

	"github.com/abenz1267/cachy"
)

func main() {
	p := &cachy.Params{URL: "/hotreload", Ext: "html", Duplicates: false, Recursive: true}
	c, _ := cachy.New(p, nil, "../")
	go c.Watch(true)

	http.Handle("/", index(c))
	http.Handle(c.URL(), http.HandlerFunc(c.HotReload))

	log.Println("Server running on http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func index(c *cachy.Cachy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Execute(w, nil, "index", "child")
	})
}
