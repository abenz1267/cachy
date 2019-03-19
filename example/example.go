package main

import (
	"log"
	"net/http"

	"github.com/abenz1267/cachy"
)

func main() {
	c, _ := cachy.New("/reload", ".html", nil)
	go c.Watch(true)

	http.Handle("/", Index(c))
	http.Handle("/reload", http.HandlerFunc(c.HotReload))

	log.Println("Server running on http://localhost:3000/")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func Index(c *cachy.Cachy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Execute(w, nil, "index")
	})
}
