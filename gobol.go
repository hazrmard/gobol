package main

import (
	"flag"
	termbox "github.com/nsf/termbox-go"
	"regexp"
)

type config struct {
	h            int               // height of console
	w            int               // width of console
	mw           int               // message width (w-len(prompt))
	bg           termbox.Attribute // background attributes (color)
	fg           termbox.Attribute // foreground attributes (color)
	prompt       string            // prompt at beginning of new line
	cprompt      string            // string to indicate commant to gobol
	nullChar     rune              // representation of empty char on console
	addrCh       string            // char representing user address. Default '@'
	userSuffix   string            // string in message after username
	addrPattern  *regexp.Regexp    // pattern to match @username addressing
	unamePattern *regexp.Regexp    // pattern to match incoming message username: msg
	ipPattern    *regexp.Regexp    // pattern to match ip in IP:PORT
	portPattern  *regexp.Regexp    // pattern to match port in IP:PORT
	cPattern     *regexp.Regexp    // pattern to match commands
	uPattern     *regexp.Regexp    // pattern to match username@
}

var ARGS args                      // contains commandline arguments
var C cursor = cursor{}            // current position of cursor
var mbuff []rune                   // a slice of runes for current message
var outbox chan message            // a channel for outgoing messages
var inbox chan message             // a channel for incoming messages
var sync chan int                  // syncs b/w new messages and shell (barrier)
var partSemaphore chan int         // semaphore for participants map
var participants map[string]client // a username->client map
var CONF config = config{
	prompt:     ":: ",
	addrCh:     "@",
	userSuffix: ": ",
	cprompt:    "\\\\",
} // some configuration settings

func init() {
	flag.StringVar(&ARGS.host, "h", "127.0.0.1", "Specify host for gobol.")
	flag.StringVar(&ARGS.port, "p", "8000", "Specify port for gobol.")
	flag.StringVar(&ARGS.username, "u", "DEFAULT", "Username for chat.")
	flag.IntVar(&ARGS.buffer, "b", 10, "Buffer size for messages.")
}

func main() {
	flag.Parse()                             // parse commandline arguments
	outbox = make(chan message, ARGS.buffer) // declare outgoing channel w/ buff
	inbox = make(chan message, ARGS.buffer)  // declare outgoing channel w/ buff
	sync = make(chan int)                    // make a syncrinization channel
	partSemaphore = make(chan int, 1)        // semaphore b/w commandHandler & handler
	partSemaphore <- 1                       // initialize semaphore with 1 token
	participants = make(map[string]client)   // initialize map for participants

	CONF.addrPattern, _ = regexp.Compile(CONF.addrCh + "(\\w+)")       // regex pattern for @username
	CONF.unamePattern, _ = regexp.Compile("^(\\w+)" + CONF.userSuffix) // regex pattern for username: msg
	CONF.ipPattern, _ = regexp.Compile("^\\w*?@?([\\w\\.]+):")
	CONF.portPattern, _ = regexp.Compile("^\\w*@?[0-9\\.]+:([0-9]+)")
	CONF.cPattern, _ = regexp.Compile("^\\\\\\\\(.*)") // \\ followed by command
	CONF.uPattern, _ = regexp.Compile("^(\\w+)@")

	go serve() // start server routine to send/receive messages
	shell()    // start interactive shell
}
