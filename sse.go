package cachy

import (
	"fmt"
	"net/http"
)

// via https://thoughtbot.com/blog/writing-a-server-sent-events-server-in-go
type broker struct {
	Notifier chan []byte
	nClients chan chan []byte
	cClients chan chan []byte
	clients  map[chan []byte]bool
}

func (b *broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	messageChan := make(chan []byte)
	b.nClients <- messageChan

	defer func() {
		b.cClients <- messageChan
	}()

	notify := w.(http.CloseNotifier).CloseNotify()

	go func() {
		<-notify
		b.cClients <- messageChan
	}()

	for {
		fmt.Fprintf(w, "data: %s\n\n", <-messageChan)
		flusher.Flush()
	}
}

func (b *broker) listen() {
	for {
		select {
		case s := <-b.nClients:
			b.clients[s] = true
		case s := <-b.cClients:
			delete(b.clients, s)
		case event := <-b.Notifier:
			for clientMessageChan := range b.clients {
				clientMessageChan <- event
			}
		}
	}
}

func (c *Cachy) bindBroker() {
	c.SSE = &broker{
		Notifier: make(chan []byte, 1),
		nClients: make(chan chan []byte),
		cClients: make(chan chan []byte),
		clients:  make(map[chan []byte]bool),
	}

	go c.SSE.listen()
}
