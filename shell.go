package main

import (
	"strings"

	termbox "github.com/nsf/termbox-go"
)

type args struct {
	host     string
	port     string
	username string
	buffer   int
}

type cursor struct {
	x  int // position from left edge of console
	y  int // position from top edge of console
	rx int // relative pos. from beginning of curr message
	ry int // relative pos. from top of current message
}

type config struct {
	h        int               // height of console
	w        int               // width of console
	mw       int               // message width (w-len(prompt))
	bg       termbox.Attribute // background attributes (color)
	fg       termbox.Attribute // foreground attributes (color)
	bgx      termbox.Attribute // active background attributes (color)
	fgx      termbox.Attribute // active foreground attributes (color)
	prompt   string            // prompt at beginning of new line
	nullChar rune              // representation of empty char on console
}

type message struct {
	content   string
	recipient client
}

// interactive shell loop
func shell() {
	termbox.Init()                             // start termbox
	defer termbox.Close()                      // push Close to bottom of stack
	CONF.nullChar = termbox.CellBuffer()[0].Ch // get "empty" char for future
	CONF.w, CONF.h = termbox.Size()            // get initial dimensions
	CONF.mw = CONF.w - len(CONF.prompt)        // calculate message width per line
	prints(CONF.prompt, false)

loop:
	for {
		// infinite loop of io
		termbox.SetCursor(C.x, C.y)
		termbox.Flush()
		CONF.w, CONF.h = termbox.Size()
		CONF.mw = CONF.w - len(CONF.prompt)
		ev := termbox.PollEvent()

		if ev.Type == termbox.EventKey {
			if ev.Ch != 0 {
				insert(ev.Ch)
			} else {
				switch ev.Key {

				case termbox.KeySpace:
					insert(rune(' '))

				case termbox.KeyCtrlQ:
					break loop

				case termbox.KeyEnter:
					C.x = len(CONF.prompt) + len(mbuff)%CONF.mw // move cursor to end of message
					C.y = (C.y - C.ry) + len(mbuff)/CONF.mw
					nextline(true, false)
					// TODO: send msg buffer to server
					mbuff = []rune{} // reset buffer
					C.rx = 0         // reset relative position w.r.t msg
					C.ry = 0

				case termbox.KeyArrowLeft:
					moveCursorLeft()

				case termbox.KeyArrowRight:
					moveCursorRight()

				case termbox.KeyBackspace:
					if C.rx > 0 {
						moveCursorLeft()
						delete()
					}
				}
			}
		} else if ev.Type == termbox.EventInterrupt {
			<-sync // wait for all-clear signal from printService
		}
	}
}

// print a single unicode character
func print(this rune, trackRelative bool) int {
	newline := 0
	termbox.SetCell(C.x, C.y, this, CONF.fg, CONF.bg)
	if C.x++; C.x >= CONF.w {
		newline = nextline(false, trackRelative)
	}
	return newline
}

// print a string of unicode characters
func prints(this string, trackRelative bool) int {
	newline := 0
	runes := []rune(this)
	for _, c := range runes {
		newline += print(c, trackRelative)
	}
	return newline
}

// print a new line with/without a newline prompt
// trackRelative=true  -> return 1 for each new line,
//              =false -> only return 1 when actual Y coord changes, like
//                        scrolling when cursor at bottom does not change C.y
func nextline(printPrompt bool, trackRelative bool) int {
	newline := 1
	if C.y++; C.y >= CONF.h { // scroll up if at bottom
		scroll(1)                   // scrolling also takes up C.y
		C.y = CONF.h - 1            // restoring C.y to end (since nextline)
		if trackRelative == false { // i.e. only return change in abs. Y coord
			newline = 0
		}
	}
	C.x = 0
	if printPrompt == true {
		prints(CONF.prompt, false)
	} else {
		prints(strings.Repeat(" ", len(CONF.prompt)), false)
	}
	return newline
}

// scrolls screen up (out of screen stuff lost)
func scroll(lines int) {
	if lines < CONF.h {
		C.y -= lines
		if C.y < 0 {
			C.y = 0
		}
		cells := termbox.CellBuffer()
		for i := 0; i < CONF.w*(CONF.h-lines); i++ {
			cells[i].Ch = cells[i+CONF.w*lines].Ch
		}
		for i := CONF.w * (CONF.h - lines); i < CONF.w*CONF.h; i++ {
			cells[i].Ch = CONF.nullChar
		}
	}
}

// move cursor left, account for newlines. Does not move beyond current message
func moveCursorLeft() {
	if C.rx > 0 {
		C.rx--
		C.x--
		if C.x < len(CONF.prompt) && C.ry > 0 { // C at beginning of multiline msg
			C.x = CONF.w - 1 // move C to end of previous line
			C.y--
			C.ry--
		}
	}
}

// move cursor right, account for newlines. Does not move beyond current message
func moveCursorRight() {
	if C.rx < len(mbuff) {
		C.rx++
		C.x++
		if C.x == CONF.w {
			C.x = len(CONF.prompt)
			C.y++
		}
	}
}

// remove whatever is under cursor
func delete() {
	mbuff = append(mbuff[:C.rx], mbuff[C.rx+1:]...)
	i := C.rx % CONF.mw
	j, k := C.ry, 0
	m := C.rx
	for ; j <= len(mbuff)/CONF.mw; j, k = j+1, k+1 {
		for ; i < CONF.mw && m < len(mbuff); i++ {
			termbox.SetCell(i+len(CONF.prompt), C.y+k, mbuff[m],
				CONF.fg, CONF.bg)
			m++
		}
		i = 0
	}
	termbox.SetCell(len(CONF.prompt)+len(mbuff)%CONF.mw,
		C.y-C.ry+len(mbuff)/CONF.mw, CONF.nullChar, CONF.fg, CONF.bg)
}

// insert new rune at cursor's position
func insert(this rune) {
	mbuff = append(mbuff[:C.rx], // append char to msg buffer
		append([]rune{this}, mbuff[C.rx:]...)...)
	newline := print(this, true)
	currx, curry := C.x, C.y // store current cursor position
	C.rx++                   // change relative positions
	C.ry += newline
	if C.rx < len(mbuff) {
		prints(string(mbuff[C.rx:]), true) // print remaining
		C.x, C.y = currx, curry            // restore cursor
	}
}

// prints a new incoming message above the current message.
func printNewMessage(m message) {
	oldCx := C.x
	oldCy := C.y
	C.y -= C.ry
	C.x = 0
	for j := 0; j <= len(mbuff)/CONF.mw; j++ { // clear current message
		for i := 0; i < CONF.w; i++ {
			termbox.SetCell(i, C.y+j, CONF.nullChar, CONF.fg, CONF.bg)
		}
	}
	oldCy += prints(CONF.prompt, false)   // print new prompt
	oldCy += prints(m.content, false)     // print incoming message
	oldCy += nextline(true, false)        // go to next line
	oldCy += prints(string(mbuff), false) // print current message
	C.x = oldCx                           // restore x/y cursors b/c printing moves it to end
	C.y = (C.y - len(mbuff)/CONF.mw) + C.ry
}
