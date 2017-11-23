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
func (ui *Tui) Start() error {
	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()
loop:
	for {
		select {
		case e := <-events:
			if e.Type == termbox.EventKey {
				if e.Ch == 'q' || e.Key == termbox.KeyCtrlC || e.Key == termbox.KeyCtrlD {
					break loop
				}
				if e.Ch == 'k' {
					ui.ch <- core.CursorUp
				}
				if e.Ch == 'j' {
					ui.ch <- core.CursorDown
				}
				if e.Ch == 'h' {
					ui.ch <- core.CursorLeft
				}
				if e.Ch == 'l' {
					ui.ch <- core.CursorRight
				}
				if e.Ch == 'b' {
					ui.ch <- core.CursorPrev
				}
				if e.Ch == 'w' {
					ui.ch <- core.CursorNext
				}
				if e.Key == termbox.KeyCtrlY {
					ui.ch <- core.ScrollUp
				}
				if e.Key == termbox.KeyCtrlE {
					ui.ch <- core.ScrollDown
				}
				if e.Key == termbox.KeyCtrlB {
					ui.ch <- core.PageUp
				}
				if e.Key == termbox.KeyCtrlF {
					ui.ch <- core.PageDown
				}
				if e.Ch == 'g' {
					ui.ch <- core.PageTop
				}
				if e.Ch == 'G' {
					ui.ch <- core.PageLast
				}
			}
		}
	}
	return nil
}

// Height returns the height for the hex view.
func (ui *Tui) Height() int {
	_, height := termbox.Size()
	return height - 1
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
	for i := 0; i < height; i++ {
		if i*width >= state.Size {
			ui.setLine(i+1, 0, strings.Repeat(" ", 4*width+13), 0)
			continue
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
		ui.setLine(i+1, 0, w.String(), 0)
	}
	i, j := int(state.Cursor%int64(width)), int(state.Cursor-state.Offset)
	cursorLine := j / width
	ui.setLine(0, 3*i+11, fmt.Sprintf("%2x", i), termbox.AttrBold|termbox.AttrUnderline)
	ui.setLine(cursorLine+1, 0, fmt.Sprintf("%08x", state.Offset+int64(cursorLine*width)), termbox.AttrBold)
	ui.setLine(cursorLine+1, 3*i+11, fmt.Sprintf("%02x", state.Bytes[j]), termbox.AttrReverse)
	ui.setLine(cursorLine+1, 3*width+13+i, string([]byte{prettyByte(state.Bytes[j])}), termbox.AttrBold)
	termbox.SetCursor(3*i+11, cursorLine+1)
	ui.drawScrollBar(state, 4*width+14)
	return termbox.Flush()
}

func (ui *Tui) drawScrollBar(state core.State, offset int) {
	height := (state.Size + state.Width - 1) / state.Width
	len := util.MaxInt(int((state.Length+int64(state.Width)-1)/int64(state.Width)), 1)
	size := util.MaxInt((state.Size+state.Width-1)/state.Width*height/len, 1)
	pad := util.MinInt(height*height-len*size+len/2, len/2)
	top := (int(state.Offset/int64(state.Width))*height + pad) / len
	for i := 0; i < height; i++ {
		if top <= i && i < top+size {
			ui.setLine(i+1, offset, " ", termbox.AttrReverse)
		} else {
			ui.setLine(i+1, offset, "|", 0)
		}
	}
}

func prettyByte(b byte) byte {
	switch {
	case 0x20 <= b && b < 0x7f:
		return b
	default:
		return 0x2e
	}
}

// Close terminates the Tui.
func (ui *Tui) Close() error {
	termbox.Close()
	close(ui.ch)
	return nil
}
