package main

import (
	"bytes"
	termbox "github.com/nsf/termbox-go"
	"net/http"
	"regexp"
)

// represents a chat participant
type client struct {
	username     string
	host         string
	port         string
	participants map[string]client
}

// handles incoming requests, and spawns print and send services
func serve(a *args) {
	go printService()
	go sendService()
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

// uses a POST request to send message to all or selected participants
func sendService() {
	r, _ := regexp.Compile("@(\\w+)") // alphanumeric usernames only
	var err error
	for {
		msg := <-outbox
		unames := r.FindAllStringSubmatch(msg.content, -1)
		if len(unames) > 0 {

		} else {
			_, err = http.Post("http://"+ARGS.host+":"+ARGS.port, "text/html", bytes.NewReader([]byte(msg.content)))
		}
		if err != nil {
			inbox <- message{content: "ERROR:" + err.Error()}
		}
	}
}
