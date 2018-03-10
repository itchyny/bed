package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"

	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/util"
)

// Tui implements UI
type Tui struct {
	width   int
	height  int
	eventCh chan<- Event
	quitCh  <-chan struct{}
	mode    Mode
}

// NewTui creates a new Tui.
func NewTui() *Tui {
	return &Tui{}
}

// Init initializes the Tui.
func (ui *Tui) Init(eventCh chan<- Event, quitCh <-chan struct{}) error {
	ui.eventCh = eventCh
	ui.quitCh = quitCh
	ui.mode = ModeNormal
	return termbox.Init()
}

// Run the Tui.
func (ui *Tui) Run(kms map[Mode]*KeyManager) {
	kms[ModeNormal].Register(EventQuit, "q")
	kms[ModeNormal].Register(EventQuit, "c-c")
	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()
	go func() {
		for {
			select {
			case e := <-events:
				if e.Type == termbox.EventKey {
					if event := kms[ui.mode].Press(eventToKey(e)); event.Type != EventNop {
						ui.eventCh <- event
					} else {
						ui.eventCh <- Event{Type: EventRune, Rune: e.Ch}
					}
				}
			case <-ui.quitCh:
				close(events)
				return
			}
		}
	}()
}

// Height returns the height for the hex view.
func (ui *Tui) Height() int {
	_, height := termbox.Size()
	return height - 3
}

// Width returns the width for the hex view.
func (ui *Tui) Width() int {
	width, _ := termbox.Size()
	return width
}

func (ui *Tui) setLine(line int, offset int, str string, attr termbox.Attribute) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	for _, c := range str {
		termbox.SetCell(offset, line, c, fg|attr, bg)
		offset += runewidth.RuneWidth(c)
	}
}

// Redraw redraws the state.
func (ui *Tui) Redraw(state State) error {
	ui.mode = state.Mode
	height, width := ui.Height(), state.Width
	bytes, attrs := ui.bytesArray(height, width, state)
	var attr termbox.Attribute
	for i := 0; i < height; i++ {
		attr := termbox.ColorDefault
		if i == height-1 {
			attr = termbox.AttrUnderline
		}
		for j := 0; j < width; j++ {
			if attrs[i][j] == math.MaxUint16 {
				ui.setLine(i+1, 3*j+10, "   ", attr)
				ui.setLine(i+1, 3*width+j+13, " ", attr)
			} else {
				ui.setLine(i+1, 3*j+10, " ", attrs[i][j]|attr)
				if i*width+j == int(state.Cursor-state.Offset) {
					attrs[i][j] |= termbox.AttrReverse
				}
				ui.setLine(i+1, 3*j+11, fmt.Sprintf("%02x", bytes[i][j]), attrs[i][j]|attr)
				if i*width+j == int(state.Cursor-state.Offset) {
					attrs[i][j] ^= termbox.AttrReverse | termbox.AttrBold
				}
				ui.setLine(i+1, 3*width+j+13, string(prettyByte(bytes[i][j])), attrs[i][j]|attr)
			}
		}
		ui.setLine(i+1, 0, fmt.Sprintf("%08x |", state.Offset+int64(i*width)), attr)
		ui.setLine(i+1, 3*width+10, " | ", attr)
		ui.setLine(i+1, 4*width+13, " ", attr)
	}
	i, j := int(state.Cursor%int64(width)), int(state.Cursor-state.Offset)
	cursorLine := j / width
	if cursorLine == height-1 {
		attr = termbox.AttrUnderline
	} else {
		attr = 0
	}
	ui.setLine(cursorLine+1, 0, fmt.Sprintf("%08x", state.Offset+int64(cursorLine*width)), termbox.AttrBold|attr)
	if state.Pending {
		termbox.SetCursor(3*i+12, cursorLine+1)
	} else {
		termbox.SetCursor(3*i+11, cursorLine+1)
	}
	ui.drawHeader(state)
	ui.drawScrollBar(state, height, 4*width+14)
	ui.drawFooter(state)
	return termbox.Flush()
}

func (ui *Tui) bytesArray(height, width int, state State) ([][]byte, [][]termbox.Attribute) {
	var k int
	eis := state.EditedIndices
	bytes := make([][]byte, height)
	attrs := make([][]termbox.Attribute, height)
	for i := 0; i < height; i++ {
		bytes[i] = make([]byte, width)
		attrs[i] = make([]termbox.Attribute, width)
		for j := 0; j < width; j++ {
			if k >= state.Size {
				attrs[i][j] = termbox.Attribute(math.MaxUint16)
			}
			if state.Pending && i*width+j == int(state.Cursor-state.Offset) {
				bytes[i][j] = state.PendingByte
				attrs[i][j] = termbox.ColorCyan
				if state.Mode == ModeReplace {
					k++
				}
				continue
			}
			bytes[i][j] = state.Bytes[k]
			if 0 < len(eis) && eis[0] <= int64(k)+state.Offset && int64(k)+state.Offset < eis[1] {
				attrs[i][j] = termbox.ColorCyan
			} else if 0 < len(eis) && eis[1] <= int64(k)+state.Offset {
				eis = eis[2:]
			}
			k++
		}
	}
	return bytes, attrs
}

func (ui *Tui) drawHeader(state State) {
	ui.setLine(0, 0, strings.Repeat(" ", 4*state.Width+15), termbox.AttrUnderline)
	cursor := int(state.Cursor % int64(state.Width))
	for i := 0; i < state.Width; i++ {
		if cursor == i {
			ui.setLine(0, 3*i+11, fmt.Sprintf("%2x", i), termbox.AttrBold|termbox.AttrUnderline)
		} else {
			ui.setLine(0, 3*i+11, fmt.Sprintf("%2x", i), termbox.AttrUnderline)
		}
	}
	ui.setLine(0, 9, "|", termbox.AttrUnderline)
	ui.setLine(0, 3*state.Width+11, "|", termbox.AttrUnderline)
}

func (ui *Tui) drawScrollBar(state State, height int, offset int) {
	total := int64((state.Size + state.Width - 1) / state.Width)
	len := util.MaxInt64((state.Length+int64(state.Width)-1)/int64(state.Width), 1)
	size := util.MaxInt64(total*total/len, 1)
	pad := (total*total + len - len*size - 1) / util.MaxInt64(total-size+1, 1)
	top := (state.Offset / int64(state.Width) * total) / (len - pad)
	for i := 0; i < height; i++ {
		if int(top) <= i && i < int(top+size) {
			ui.setLine(i+1, offset, " ", termbox.AttrReverse)
		} else if i < height-1 {
			ui.setLine(i+1, offset, "|", 0)
		} else {
			ui.setLine(i+1, offset, "|", termbox.AttrUnderline)
		}
	}
}

func (ui *Tui) drawFooter(state State) {
	j := int(state.Cursor - state.Offset)
	name := state.Name
	if name == "" {
		name = "[No name]"
	}
	line := fmt.Sprintf("%s%s: %08x / %08x (%.2f%%) [0x%02x '%s']               ",
		prettyMode(state.Mode), name, state.Cursor, state.Length, float64(state.Cursor*100)/float64(util.MaxInt64(state.Length, 1)),
		state.Bytes[j], prettyRune(state.Bytes[j]))
	ui.setLine(ui.Height()+1, 0, line, 0)
	if state.Error != nil {
		ui.setLine(ui.Height()+2, 0, state.Error.Error()+strings.Repeat(" ", ui.Width()), termbox.ColorRed)
	} else if state.Mode == ModeCmdline {
		ui.setLine(ui.Height()+2, 0, ":"+string(state.Cmdline)+strings.Repeat(" ", ui.Width()), 0)
		termbox.SetCursor(1+runewidth.StringWidth(string(state.Cmdline[:state.CmdlineCursor])), ui.Height()+2)
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

func prettyMode(mode Mode) string {
	switch mode {
	case ModeInsert:
		return "[INSERT] "
	case ModeReplace:
		return "[REPLACE] "
	default:
		return ""
	}
}

// Close terminates the Tui.
func (ui *Tui) Close() error {
	termbox.Close()
	return nil
}
