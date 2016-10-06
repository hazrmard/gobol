package main

import (
    "github.com/nsf/termbox-go"
    "bytes"
)

type args struct {
    host string
    port string
}

type cursor struct {
    x int                       // position from left edge of console
    y int                       // position from top edge of console
}

type config struct {
    h int                       // height of console
    w int                       // width of console
    bg termbox.Attribute        // background attributes (color)
    fg termbox.Attribute        // foreground attributes (color)
    bgx termbox.Attribute       // active background attributes (color)
    fgx termbox.Attribute       // active foreground attributes (color)
    prompt string               // prompt at beginning of new line
}

// interactive shell loop
func shell() {
    termbox.Init()                      // start termbox
    nullChar := termbox.Cell{Buffer()[0]}.Ch
    CONF.w, CONF.h = termbox.Size()     // get initial dimensions
    defer termbox.Close()               // push Close to bottom of stack
    var mbuff bytes.Buffer              // set a buffer for current message
    prints(CONF.prompt)

    loop:
    for {
        // infinite loop of io
        termbox.SetCursor(C.x, C.y)
        termbox.Flush()
        CONF.w, CONF.h = termbox.Size()
        ev:=termbox.PollEvent()

        if ev.Type==termbox.EventKey {
            if ev.Ch != 0 {
                print(ev.Ch)
                mbuff.WriteRune(ev.Ch)
            } else {
                switch ev.Key {

                case termbox.KeyCtrlQ:
                    break loop

                case termbox.KeyEnter:
                    nextline(true)
                    mbuff.Reset()

                case termbox.KeySpace:
                    print(' ')
                    mbuff.WriteRune(' ')

                case termbox.KeyBackspace:
                    C.x--
                    if C.x<0 {
                        C.x = 0
                        C.y--
                        if C.y<0 {
                            C.y = 0
                        }
                    }
                    termbox.SetCell(C.x, C.y, nullChar, CONF.fg, CONF.bg)
                    // TODO: truncate buff to reflect deletion
                }
            }
        }
    }
}

// print a single unicode character
func print(this rune) {
    termbox.SetCell(C.x, C.y, this, CONF.fg, CONF.bg)
    if C.x++; C.x>=CONF.w {
        nextline(false)
    }
}

// print a string
func prints(this string) {
    runes := []rune(this)
    for _,c:=range(runes) {
        print(c)
    }
}

// print a new line with/without a newline prompt
func nextline(prompt bool) {
    C.y++
    C.x = 0
    if prompt==true {
        prints(CONF.prompt)
    }
}
