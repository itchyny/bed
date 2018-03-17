package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"

	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/util"
)

// Tui implements UI
type Tui struct {
	width   int
	height  int
	eventCh chan<- Event
	mode    Mode
	screen  tcell.Screen
}

// NewTui creates a new Tui.
func NewTui() *Tui {
	return &Tui{}
}

// Init initializes the Tui.
func (ui *Tui) Init(eventCh chan<- Event) (err error) {
	ui.eventCh = eventCh
	ui.mode = ModeNormal
	if ui.screen, err = tcell.NewScreen(); err != nil {
		return
	}
	return ui.screen.Init()
}

// Run the Tui.
func (ui *Tui) Run(kms map[Mode]*KeyManager) {
	kms[ModeNormal].Register(EventQuit, "q")
	kms[ModeNormal].Register(EventQuit, "c-c")
	for {
		e := ui.screen.PollEvent()
		switch ev := e.(type) {
		case *tcell.EventKey:
			if event := kms[ui.mode].Press(eventToKey(ev)); event.Type != EventNop {
				ui.eventCh <- event
			} else {
				ui.eventCh <- Event{Type: EventRune, Rune: ev.Rune()}
			}
		case nil:
			return
		}
	}
}

// Height returns the height for the hex view.
func (ui *Tui) Height() int {
	_, height := ui.screen.Size()
	return height - 3
}

// Width returns the width for the hex view.
func (ui *Tui) Width() int {
	width, _ := ui.screen.Size()
	return width
}

func (ui *Tui) setLine(line int, offset int, str string, style tcell.Style) {
	for _, c := range str {
		ui.screen.SetContent(offset, line, c, nil, style)
		offset += runewidth.RuneWidth(c)
	}
}

// Redraw redraws the state.
func (ui *Tui) Redraw(state State) error {
	ui.mode = state.Mode
	ui.drawWindow(state.Windows[0])
	ui.drawCmdline(state)
	ui.screen.Show()
	return nil
}

func (ui *Tui) drawWindow(state WindowState) {
	height, width := ui.Height(), state.Width
	bytes, styles := ui.bytesArray(height, width, state)
	cursorPos := int(state.Cursor - state.Offset)
	cursorLine := cursorPos / width
	for i := 0; i < height; i++ {
		style := tcell.StyleDefault.Underline(i == height-1)
		for j := 0; j < width; j++ {
			if styles[i][j] == math.MaxUint16 {
				ui.setLine(i+1, 3*j+10, "   ", style)
				ui.setLine(i+1, 3*width+j+13, " ", style)
			} else {
				ui.setLine(i+1, 3*j+10, " ", styles[i][j]|style)
				if i*width+j == cursorPos {
					styles[i][j] = styles[i][j].Reverse(!state.FocusText).Bold(state.FocusText).Underline(state.FocusText)
				}
				ui.setLine(i+1, 3*j+11, fmt.Sprintf("%02x", bytes[i][j]), styles[i][j]|style)
				if i*width+j == cursorPos {
					styles[i][j] = styles[i][j].Reverse(state.FocusText).Bold(!state.FocusText).Underline(!state.FocusText)
				}
				ui.setLine(i+1, 3*width+j+13, string(prettyByte(bytes[i][j])), styles[i][j]|style)
			}
		}
		ui.setLine(i+1, 0, fmt.Sprintf("%08x", state.Offset+int64(i*width)), style.Bold(i == cursorLine))
		ui.setLine(i+1, 8, " | ", style)
		ui.setLine(i+1, 3*width+10, " | ", style)
		ui.setLine(i+1, 4*width+13, " ", style)
	}
	i := int(state.Cursor % int64(width))
	if state.FocusText {
		ui.screen.ShowCursor(3*width+i+13, cursorLine+1)
	} else if state.Pending {
		ui.screen.ShowCursor(3*i+12, cursorLine+1)
	} else {
		ui.screen.ShowCursor(3*i+11, cursorLine+1)
	}
	ui.drawHeader(state)
	ui.drawScrollBar(state, height, 4*width+14)
	ui.drawFooter(state)
}

func (ui *Tui) bytesArray(height, width int, state WindowState) ([][]byte, [][]tcell.Style) {
	var k int
	eis := state.EditedIndices
	bytes := make([][]byte, height)
	styles := make([][]tcell.Style, height)
	for i := 0; i < height; i++ {
		bytes[i] = make([]byte, width)
		styles[i] = make([]tcell.Style, width)
		for j := 0; j < width; j++ {
			if k >= state.Size {
				styles[i][j] = tcell.Style(math.MaxUint16)
			}
			if state.Pending && i*width+j == int(state.Cursor-state.Offset) {
				bytes[i][j] = state.PendingByte
				styles[i][j] = styles[i][j].Foreground(tcell.ColorSkyblue)
				if state.Mode == ModeReplace {
					k++
				}
				continue
			}
			bytes[i][j] = state.Bytes[k]
			if 0 < len(eis) && eis[0] <= int64(k)+state.Offset && int64(k)+state.Offset < eis[1] {
				styles[i][j] = styles[i][j].Foreground(tcell.ColorSkyblue)
			} else if 0 < len(eis) && eis[1] <= int64(k)+state.Offset {
				eis = eis[2:]
			}
			k++
		}
	}
	return bytes, styles
}

func (ui *Tui) drawHeader(state WindowState) {
	style := tcell.StyleDefault.Underline(true)
	ui.setLine(0, 0, strings.Repeat(" ", 4*state.Width+15), style)
	cursor := int(state.Cursor % int64(state.Width))
	for i := 0; i < state.Width; i++ {
		ui.setLine(0, 3*i+11, fmt.Sprintf("%2x", i), style.Bold(cursor == i))
	}
	ui.setLine(0, 9, "|", style)
	ui.setLine(0, 3*state.Width+11, "|", style)
}

func (ui *Tui) drawScrollBar(state WindowState, height int, offset int) {
	stateSize := state.Size
	if state.Cursor+1 == state.Length && state.Cursor == state.Offset+int64(state.Size) {
		stateSize++
	}
	total := int64((stateSize + state.Width - 1) / state.Width)
	len := util.MaxInt64((state.Length+int64(state.Width)-1)/int64(state.Width), 1)
	size := util.MaxInt64(total*total/len, 1)
	pad := (total*total + len - len*size - 1) / util.MaxInt64(total-size+1, 1)
	top := (state.Offset / int64(state.Width) * total) / (len - pad)
	for i := 0; i < height; i++ {
		if int(top) <= i && i < int(top+size) {
			ui.setLine(i+1, offset, " ", tcell.StyleDefault.Reverse(true))
		} else if i < height-1 {
			ui.setLine(i+1, offset, "|", 0)
		} else {
			ui.setLine(i+1, offset, "|", tcell.StyleDefault.Underline(true))
		}
	}
}

func (ui *Tui) drawFooter(state WindowState) {
	j := int(state.Cursor - state.Offset)
	name := state.Name
	if name == "" {
		name = "[No name]"
	}
	line := fmt.Sprintf("%s%s: %08x / %08x (%.2f%%) [0x%02x '%s']"+strings.Repeat(" ", ui.Width()),
		prettyMode(state.Mode), name, state.Cursor, state.Length, float64(state.Cursor*100)/float64(util.MaxInt64(state.Length, 1)),
		state.Bytes[j], prettyRune(state.Bytes[j]))
	ui.setLine(ui.Height()+1, 0, line, 0)
}

func (ui *Tui) drawCmdline(state State) {
	if state.Error != nil {
		style := tcell.StyleDefault.Foreground(tcell.ColorRed)
		if state.ErrorType == MessageInfo {
			style = style.Foreground(tcell.ColorYellow)
		}
		ui.setLine(ui.Height()+2, 0, state.Error.Error()+strings.Repeat(" ", ui.Width()), style)
	} else if state.Mode == ModeCmdline {
		ui.setLine(ui.Height()+2, 0, ":"+string(state.Cmdline)+strings.Repeat(" ", ui.Width()), 0)
		ui.screen.ShowCursor(1+runewidth.StringWidth(string(state.Cmdline[:state.CmdlineCursor])), ui.Height()+2)
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
	ui.screen.Fini()
	return nil
}
