package main

import (
	"flag"
)

var ARGS args                                        // contains commandline arguments
var C cursor = cursor{}                              // current position of cursor
var mbuff []rune                                     // a slice of runes for current message
var outbox chan message                              // a channel for outgoing messages
var inbox chan message                               // a channel for incoming messages
var sync chan int                                    // syncs b/w new messages and shell
var CONF config = config{prompt: ":: ", addrCh: '@'} // some configuration settings

func init() {
	flag.StringVar(&ARGS.host, "h", "127.0.0.1", "Specify host for gobol.")
	flag.StringVar(&ARGS.port, "p", "8000", "Specify port for gobol.")
	flag.StringVar(&ARGS.username, "u", "", "Username for chat.")
	flag.IntVar(&ARGS.buffer, "b", 10, "Buffer size for messages.")
}

func main() {
	flag.Parse()                             // parse commandline arguments
	outbox = make(chan message, ARGS.buffer) // declare outgoing channel w/ buff
	inbox = make(chan message, ARGS.buffer)  // declare outgoing channel w/ buff
	sync = make(chan int)                    // make a syncrinization channel
	go serve(&ARGS)                          // start server routine to send/receive messages
	shell()                                  // start interactive shell
}
