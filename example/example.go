package main

import (
	"log"
	"net/http"
	"os"

	"github.com/abenz1267/cachy"
)

func main() {
	c, err := cachy.New(".html", nil, nil, "templates")
	if err != nil {
		log.Fatal(err)
	}

	go c.Watch("wss://localhost:3000/ws")

	http.Handle("/", index(c))
	http.HandleFunc("/ws", c.HotReload)

	log.Println("Running server on https://localhost:3000 ...")
	log.Fatal(http.ListenAndServeTLS(":3000", os.Getenv("CERT"), os.Getenv("KEY"), nil))
}

func index(c *cachy.Cachy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := c.Execute(w, nil, "templates/base", "templates/index")
		if err != nil {
			log.Println(err)
		}
	})
}
