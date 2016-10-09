package main

import (
	"bytes"
	"net/http"

	termbox "github.com/nsf/termbox-go"
)

// represents a chat participant
type client struct {
	username     string
	host         string
	port         string
	participants []client
}

// an infinite loop that handles incoming requests
func serve(a *args) {
	go printService()
	http.HandleFunc("/", handler)
	http.ListenAndServe(a.host+":"+a.port, nil)
}

// a handler for each incoming request
func handler(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 1024)              // make 1kb buffer for message body
	r.Body.Read(buf)                       // read into buffer
	buf = bytes.Trim(buf, "\x00")          // trim null bytes
	r.Body.Close()                         // close request
	inbox <- message{content: string(buf)} // push to channel for printService
}

// a service that prints whatever message comes into 'inbox' channel
func printService() {
	for {
		msg := <-inbox
		termbox.Interrupt()
		printNewMessage(msg)
		sync <- 1
	}
}
