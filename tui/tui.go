package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"

	. "github.com/itchyny/bed/common"
)

// Tui implements UI
type Tui struct {
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

func (ui *Tui) initForTest(eventCh chan<- Event, screen tcell.SimulationScreen) (err error) {
	ui.eventCh = eventCh
	ui.mode = ModeNormal
	ui.screen = screen
	return ui.screen.Init()
}

// Run the Tui.
func (ui *Tui) Run(kms map[Mode]*KeyManager) {
	for {
		e := ui.screen.PollEvent()
		switch ev := e.(type) {
		case *tcell.EventKey:
			if event := kms[ui.mode].Press(eventToKey(ev)); event.Type != EventNop {
				ui.eventCh <- event
			} else {
				ui.eventCh <- Event{Type: EventRune, Rune: ev.Rune()}
			}
		case *tcell.EventResize:
			ui.eventCh <- Event{Type: EventRedraw}
		case nil:
			return
		}
	}
}

// Size returns the size for the screen.
func (ui *Tui) Size() (int, int) {
	return ui.screen.Size()
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
	ui.screen.Clear()
	ui.drawWindows(state.WindowStates, state.Layout)
	ui.drawCmdline(state)
	ui.screen.Show()
	return nil
}

func (ui *Tui) drawWindows(windowStates map[int]*WindowState, layout Layout) {
	switch l := layout.(type) {
	case LayoutWindow:
		r := fromLayout(layout)
		if r.valid() {
			ui.newTuiWindow(r).drawWindow(windowStates[l.Index], l.Active)
		}
	case LayoutHorizontal:
		ui.drawWindows(windowStates, l.Top)
		ui.drawWindows(windowStates, l.Bottom)
	case LayoutVertical:
		ui.drawWindows(windowStates, l.Left)
		ui.drawWindows(windowStates, l.Right)
		ui.drawVerticalSplit(fromLayout(l.Left))
	}
}

func (ui *Tui) newTuiWindow(region region) *tuiWindow {
	return &tuiWindow{region: region, screen: ui.screen}
}

func (ui *Tui) drawVerticalSplit(region region) {
	for i := 0; i < region.height; i++ {
		ui.setLine(region.top+i, region.left+region.width, "|", tcell.StyleDefault.Reverse(true))
	}
}

func (ui *Tui) drawCmdline(state State) {
	width, height := ui.Size()
	if state.Error != nil {
		style := tcell.StyleDefault.Foreground(tcell.ColorRed)
		if state.ErrorType == MessageInfo {
			style = style.Foreground(tcell.ColorYellow)
		}
		ui.setLine(height-1, 0, state.Error.Error(), style)
	} else if state.Mode == ModeCmdline {
		if len(state.CompletionResults) > 0 {
			var line string
			var pos int
			for i, result := range state.CompletionResults {
				if len(line)+len(result)+2 > width && i <= state.CompletionIndex {
					line = ""
				}
				if state.CompletionIndex == i {
					pos = len(line)
				}
				line += " " + result + " "
			}
			ui.setLine(height-2, 0, line+strings.Repeat(" ", width), tcell.StyleDefault.Reverse(true))
			if state.CompletionIndex >= 0 {
				ui.setLine(height-2, pos, " "+state.CompletionResults[state.CompletionIndex]+" ",
					tcell.StyleDefault.Foreground(tcell.ColorGrey).Reverse(true))
			}
		}
		ui.setLine(height-1, 0, ":"+string(state.Cmdline), tcell.StyleDefault)
		ui.screen.ShowCursor(1+runewidth.StringWidth(string(state.Cmdline[:state.CmdlineCursor])), height-1)
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
