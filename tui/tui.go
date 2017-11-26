package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/itchyny/bed/core"
	"github.com/itchyny/bed/util"
	termbox "github.com/nsf/termbox-go"
)

// Tui implements UI
type Tui struct {
	width  int
	height int
	ch     chan<- core.Event
}

// NewTui creates a new Tui.
func NewTui() *Tui {
	return &Tui{}
}

// Init initializes the Tui.
func (ui *Tui) Init(ch chan<- core.Event) error {
	ui.ch = ch
	return termbox.Init()
}

// Start starts the Tui.
func (ui *Tui) Start(km *core.KeyManager) error {
	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()
	km.Register(core.Quit, "q")
	km.Register(core.Quit, "c-c")
loop:
	for {
		select {
		case e := <-events:
			if e.Type == termbox.EventKey {
				if event := km.Press(eventToKey(e)); event.Type != core.Nop {
					if event.Type == core.Quit {
						break loop
					}
					ui.ch <- event
					continue
				}
			}
		}
	}
	return nil
}

// Height returns the height for the hex view.
func (ui *Tui) Height() int {
	_, height := termbox.Size()
	return height - 2
}

func (ui *Tui) setLine(line int, offset int, str string, attr termbox.Attribute) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	for i, c := range str {
		termbox.SetCell(offset+i, line, c, fg|attr, bg)
	}
}

// Redraw redraws the state.
func (ui *Tui) Redraw(state core.State) error {
	height, width := ui.Height(), state.Width
	ui.setLine(0, 0, strings.Repeat(" ", 4*width+15), termbox.AttrUnderline)
	for i := 0; i < width; i++ {
		ui.setLine(0, 3*i+11, fmt.Sprintf("%2x", i), termbox.AttrUnderline)
	}
	ui.setLine(0, 9, "|", termbox.AttrUnderline)
	ui.setLine(0, 3*width+11, "|", termbox.AttrUnderline)
	var attr termbox.Attribute
	for i := 0; i < height; i++ {
		if i == height-1 {
			attr = termbox.AttrUnderline
		}
		w := new(bytes.Buffer)
		fmt.Fprintf(w, "%08x |", state.Offset+int64(i*width))
		buf := make([]byte, width)
		for j := 0; j < width; j++ {
			k := i*width + j
			if k >= state.Size {
				fmt.Fprintf(w, "   ")
				continue
			}
			fmt.Fprintf(w, " %02x", state.Bytes[k])
			buf[j] = prettyByte(state.Bytes[k])
		}
		fmt.Fprintf(w, " | %s\n", buf)
		ui.setLine(i+1, 0, w.String(), attr)
	}
	i, j := int(state.Cursor%int64(width)), int(state.Cursor-state.Offset)
	cursorLine := j / width
	ui.setLine(0, 3*i+11, fmt.Sprintf("%2x", i), termbox.AttrBold|termbox.AttrUnderline)
	if cursorLine == height-1 {
		attr = termbox.AttrUnderline
	} else {
		attr = 0
	}
	ui.setLine(cursorLine+1, 0, fmt.Sprintf("%08x", state.Offset+int64(cursorLine*width)), termbox.AttrBold|attr)
	ui.setLine(cursorLine+1, 3*i+11, fmt.Sprintf("%02x", state.Bytes[j]), termbox.AttrReverse|attr)
	ui.setLine(cursorLine+1, 3*width+13+i, string([]byte{prettyByte(state.Bytes[j])}), termbox.AttrBold|attr)
	termbox.SetCursor(3*i+11, cursorLine+1)
	ui.drawScrollBar(state, 4*width+14)
	ui.drawFooter(state)
	return termbox.Flush()
}

func (ui *Tui) drawScrollBar(state core.State, offset int) {
	height := int64((state.Size + state.Width - 1) / state.Width)
	len := util.MaxInt64((state.Length+int64(state.Width)-1)/int64(state.Width), 1)
	size := util.MaxInt64(height*height/len, 1)
	pad := (height*height + len - len*size - 1) / util.MaxInt64(height-size+1, 1)
	top := (state.Offset / int64(state.Width) * height) / (len - pad)
	for i := 0; i < int(height); i++ {
		if int(top) <= i && i < int(top+size) {
			ui.setLine(i+1, offset, " ", termbox.AttrReverse)
		} else if i < int(height-1) {
			ui.setLine(i+1, offset, "|", 0)
		} else {
			ui.setLine(i+1, offset, "|", termbox.AttrUnderline)
		}
	}
}

func (ui *Tui) drawFooter(state core.State) {
	j := int(state.Cursor - state.Offset)
	line := fmt.Sprintf("%s: %08x / %08x (%.2f%%) [0x%02x '%s']     ",
		state.Name, state.Cursor, state.Length, float64(state.Cursor*100)/float64(util.MaxInt64(state.Length, 1)),
		state.Bytes[j], prettyRune(state.Bytes[j]))
	ui.setLine(ui.Height()+1, 0, line, 0)
}

func prettyByte(b byte) byte {
	switch {
	case 0x20 <= b && b < 0x7f:
		return b
	default:
		return 0x2e
	}
}

func prettyRune(b byte) string {
	switch {
	case b == 0x07:
		return "\\a"
	case b == 0x08:
		return "\\b"
	case b == 0x09:
		return "\\t"
	case b == 0x0a:
		return "\\n"
	case b == 0x0b:
		return "\\v"
	case b == 0x0c:
		return "\\f"
	case b == 0x0d:
		return "\\r"
	case b < 0x20:
		return fmt.Sprintf("\\x%02x", b)
	case b == 0x27:
		return "\\'"
	case b < 0x7f:
		return string(rune(b))
	default:
		return fmt.Sprintf("\\u%04x", b)
	}
}

// Close terminates the Tui.
func (ui *Tui) Close() error {
	termbox.Close()
	close(ui.ch)
	return nil
}
