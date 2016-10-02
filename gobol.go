package main

import (
    "fmt"
    "flag"
    "bufio"
    "os"
)

var ARGS args

func init() {
    flag.StringVar(&ARGS.host, "h", "0.0.0.0", "Specify host for gobol.")
    flag.StringVar(&ARGS.port, "p", "8000", "Specify port for gobol.")
}

func main() {
    flag.Parse()
    fmt.Println("GOBOL\n\nArguments:")
    fmt.Println(ARGS)
    go serve(&ARGS)

    scanner := bufio.NewReader(os.Stdin)
    for {
        fmt.Print(":: ")
        in,_:=scanner.ReadString('\n');
        if (in[:5]=="\\quit") {
            os.Exit(0)
        }
    }
}
