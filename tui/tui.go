package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/itchyny/bed/core"
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
	return height
}

// SetLine sets the line.
func (ui *Tui) SetLine(line int, str string) error {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	for i, c := range str {
		termbox.SetCell(i, line, c, fg, bg)
	}
	return nil
}

// Redraw redraws the state.
func (ui *Tui) Redraw(state core.State) error {
	height, width := ui.Height(), 16
	for i := 0; i < height; i++ {
		if i*width >= state.Size {
			ui.SetLine(i, strings.Repeat(" ", 11+4*width))
			continue
		}
		w := new(bytes.Buffer)
		fmt.Fprintf(w, "%08x:", int64(state.Line+i)*int64(width))
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
		fmt.Fprintf(w, "  %s\n", buf)
		ui.SetLine(i, w.String())
	}
	termbox.SetCursor(3*state.Cursor.Y+10, state.Cursor.X)
	return termbox.Flush()
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
