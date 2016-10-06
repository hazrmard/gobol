package main

import (
    //"fmt"
    "flag"
    //"bufio"
    //"os"
)

var ARGS args
var C cursor = cursor{0,0}
var CONF config = config{prompt: ":: "}

func init() {
    flag.StringVar(&ARGS.host, "h", "0.0.0.0", "Specify host for gobol.")
    flag.StringVar(&ARGS.port, "p", "8000", "Specify port for gobol.")
}


func main() {
    flag.Parse()            // parse commandline arguments
    go serve(&ARGS)         // start server routine to send/receive messages
    shell()                 // start interactive shell
}
