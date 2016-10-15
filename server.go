package main

import (
	"bytes"
	"net"
	"net/http"
	"strings"
)

// represents a chat participant
type client struct {
	username string
	host     string
	port     string
	blocked  bool
}

// handles incoming requests, and spawns print and send services
func serve() {
	go sendService()
	http.HandleFunc("/", handler)
	l4, _ := net.Listen("tcp4", ":"+ARGS.port)
	l6, _ := net.Listen("tcp6", ":"+ARGS.port)
	go http.Serve(l6, http.DefaultServeMux) // Serve ipv6 requests
	http.Serve(l4, http.DefaultServeMux)    // and ipv4 requests
}

// a handler for each incoming request/message
func handler(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 1024)     // make 1kb buffer for message body
	r.Body.Read(buf)              // read into buffer
	r.Body.Close()                // close request
	buf = bytes.Trim(buf, "\x00") // trim null bytes

	uMatch := CONF.uPattern.FindAllStringSubmatch(string(buf), -1)     // sender's username added to msg body
	ipMatch := CONF.ipPattern.FindAllStringSubmatch(string(buf), -1)   // IP manually added to msg body (127.0.0.1)
	pMatch := CONF.portPattern.FindAllStringSubmatch(string(buf), -1)  // Port number specified in msg body
	ripMatch := CONF.ipPattern.FindAllStringSubmatch(r.RemoteAddr, -1) // request IP: external ip of sender

	if len(uMatch) >= 1 && len(ipMatch) >= 1 && len(pMatch) >= 1 && len(ripMatch) >= 1 {
		uname := uMatch[0][1]   // sender's username
		uport := pMatch[0][1]   // port the sender is listening on
		mip := ipMatch[0][1]    // sender's local IP appended to message body
		uip := ripMatch[0][1]   // user's external IP
		fuip := formatAddr(uip) // formated user ip

		<-partSemaphore // enter critical section
		participant, exists := participants[uname]
		if exists == true && participant.blocked == true {
			ok(w, "")
			partSemaphore <- 1
			return
		} else if exists == true && participant.host != fuip {
			badResponse(w, "Another user with the same username: "+uname+"@"+uip+":"+uport)
			partSemaphore <- 1 // exit critical section before return
			return
		} else if exists == false && uname == ARGS.username {
			badResponse(w, "Same as host username.")
			partSemaphore <- 1 // exit critical section before return
			return
		} else if exists == false && uname != ARGS.username { // if username does not exist, add to participants
			participants[uname] = &client{
				username: uname,
				host:     fuip,
				port:     uport,
			}
		}
		partSemaphore <- 1 // exit critical section

		ok(w, uname+"@"+uip+":"+uport)
		inbox <- message{ // finally,
			user:    uname + "@" + uip + ":" + uport,
			content: string(buf)[len(uname+"@"+mip+":"+uport+CONF.userSuffix):],
		} // push to channel for printService
	} else {
		badResponse(w, "Bad format.")
	}

}

// uses a POST request to send message to all or selected participants
func sendService() {
	var err error
	var resp *http.Response
	for {
		msg := <-outbox
		unames := CONF.addrPattern.FindAllStringSubmatch(msg.content, -1)
		var clients []client

		<-partSemaphore      // enter critical section to access participants
		if len(unames) > 0 { // create a slice of clients to send to
			for _, u := range unames {
				if c, exists := participants[u[1]]; exists == true {
					clients = append(clients, *c)
				}
			}
		} else {
			for _, c := range participants {
				if c.blocked == false {
					clients = append(clients, *c)
				}
			}
		}
		partSemaphore <- 1 // exit critical section to access participants

		for _, c := range clients { // send clients
			resp, err = http.Post("http://"+c.host+":"+c.port, "text/html",
				bytes.NewReader([]byte(ARGS.username+"@"+ARGS.host+":"+ARGS.port+CONF.userSuffix+msg.content)))

			if err != nil {
				inbox <- message{user: "LOCAL ERROR", content: err.Error()}
			}
			if resp != nil && resp.StatusCode != http.StatusOK {
				buf := make([]byte, 1024)     // make 1kb buffer for message body
				resp.Body.Read(buf)           // read into buffer
				resp.Body.Close()             // close request
				buf = bytes.Trim(buf, "\x00") // trim null bytes
				inbox <- message{user: "REMOTE ERROR", content: string(buf)}
			}
		}
	}
}

// returns an error response
func badResponse(w http.ResponseWriter, s string) {
	w.WriteHeader(http.StatusConflict)
	w.Write([]byte(s))
}

// writes all clear to sender
func ok(w http.ResponseWriter, s string) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s))
}

// wraps ipv6 addresses in [ ]
func formatAddr(addr string) string {
	if strings.Count(addr, ":") > 1 {
		return "[" + addr + "]"
	}
	return addr
}
