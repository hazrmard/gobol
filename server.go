package main

import (
	"bytes"
	"net/http"
)

// represents a chat participant
type client struct {
	username string
	host     string
	port     string
}

// handles incoming requests, and spawns print and send services
func serve() {
	go sendService()
	http.HandleFunc("/", handler)
	http.ListenAndServe(ARGS.host+":"+ARGS.port, nil)
}

// a handler for each incoming request/message
func handler(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 1024)     // make 1kb buffer for message body
	r.Body.Read(buf)              // read into buffer
	r.Body.Close()                // close request
	buf = bytes.Trim(buf, "\x00") // trim null bytes
	// unameMatch := CONF.unamePattern.FindAllStringSubmatch(string(buf), -1)
	uMatch := CONF.uPattern.FindAllStringSubmatch(string(buf), -1)
	ipMatch := CONF.ipPattern.FindAllStringSubmatch(string(buf), -1)
	pMatch := CONF.portPattern.FindAllStringSubmatch(string(buf), -1)

	if len(uMatch) == 1 && len(ipMatch) == 1 && len(pMatch) == 1 {

		uname := uMatch[0][1]
		uport := pMatch[0][1]
		uip := CONF.ipPattern.FindAllStringSubmatch(r.RemoteAddr, -1)[0][1]
		// Port that the POST request originates from is not the same as the
		// port # that the other client is listening on.
		//uport := CONF.portPattern.FindAllStringSubmatch(r.RemoteAddr, -1)[0][1]

		<-partSemaphore // enter critical section
		participant, exists := participants[uname]

		if exists == true && participant.host != uip {
			badResponse(w, "Another user with the same username: "+uname+"@"+uip+":"+uport)
			partSemaphore <- 1
			return
		} else if exists == false && uname == ARGS.username {
			badResponse(w, "Same as host username.")
			partSemaphore <- 1
			return
		} else if exists == false && uname != ARGS.username { // if username does not exist, add to participants
			participants[uname] = client{
				username: uname,
				host:     uip,
				port:     uport,
			}
		}
		ok(w, uname+"@"+uip+":"+uport)
		partSemaphore <- 1
		inbox <- message{ // finally,
			user:    uname + "@" + uip + ":" + uport,
			content: string(buf)[len(uname+"@"+uip+":"+uport+CONF.userSuffix):],
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
					clients = append(clients, c)
				}
			}
		} else {
			for _, c := range participants {
				clients = append(clients, c)
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
