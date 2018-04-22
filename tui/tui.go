package tui

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"

	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/key"
	"github.com/itchyny/bed/layout"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/state"
)

// Tui implements UI
type Tui struct {
	eventCh chan<- event.Event
	mode    mode.Mode
	screen  tcell.Screen
	waitCh  chan struct{}
}

// NewTui creates a new Tui.
func NewTui() *Tui {
	return &Tui{}
}

// Init initializes the Tui.
func (ui *Tui) Init(eventCh chan<- event.Event) (err error) {
	ui.eventCh = eventCh
	ui.mode = mode.Normal
	if ui.screen, err = tcell.NewScreen(); err != nil {
		return
	}
	ui.waitCh = make(chan struct{})
	return ui.screen.Init()
}

// Run the Tui.
func (ui *Tui) Run(kms map[mode.Mode]*key.Manager) {
	for {
		e := ui.screen.PollEvent()
		switch ev := e.(type) {
		case *tcell.EventKey:
			if e := kms[ui.mode].Press(eventToKey(ev)); e.Type != event.Nop {
				ui.eventCh <- e
			} else {
				ui.eventCh <- event.Event{Type: event.Rune, Rune: ev.Rune()}
			}
		case *tcell.EventResize:
			if ui.eventCh != nil {
				ui.eventCh <- event.Event{Type: event.Redraw}
			}
		case nil:
			close(ui.waitCh)
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
func (ui *Tui) Redraw(s state.State) error {
	ui.mode = s.Mode
	ui.screen.Clear()
	ui.drawWindows(s.WindowStates, s.Layout)
	ui.drawCmdline(s)
	ui.screen.Show()
	return nil
}

func (ui *Tui) drawWindows(windowStates map[int]*state.WindowState, l layout.Layout) {
	switch l := l.(type) {
	case layout.Window:
		r := fromLayout(l)
		if r.valid() {
			ui.newTuiWindow(r).drawWindow(
				windowStates[l.Index],
				l.Active && ui.mode != mode.Cmdline && ui.mode != mode.Search,
			)
		}
	case layout.Horizontal:
		ui.drawWindows(windowStates, l.Top)
		ui.drawWindows(windowStates, l.Bottom)
	case layout.Vertical:
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

func (ui *Tui) drawCmdline(s state.State) {
	width, height := ui.Size()
	if s.Error != nil {
		style := tcell.StyleDefault.Foreground(tcell.ColorRed)
		if s.ErrorType == state.MessageInfo {
			style = style.Foreground(tcell.ColorYellow)
		}
		ui.setLine(height-1, 0, s.Error.Error(), style)
	} else if s.Mode == mode.Cmdline || s.PrevMode == mode.Cmdline && len(s.Cmdline) > 0 {
		ui.setLine(height-1, 0, ":"+string(s.Cmdline), tcell.StyleDefault)
		if s.Mode == mode.Cmdline {
			ui.drawCompletionResults(s, width, height)
			ui.screen.ShowCursor(1+runewidth.StringWidth(string(s.Cmdline[:s.CmdlineCursor])), height-1)
		}
	} else if s.SearchMode != '\x00' {
		ui.setLine(height-1, 0, string(s.SearchMode)+string(s.Cmdline), tcell.StyleDefault)
		if s.Mode == mode.Search {
			ui.screen.ShowCursor(1+runewidth.StringWidth(string(s.Cmdline[:s.CmdlineCursor])), height-1)
		}
	}
}

func (ui *Tui) drawCompletionResults(s state.State, width int, height int) {
	if len(s.CompletionResults) > 0 {
		var line string
		var pos, lineWidth int
		for i, result := range s.CompletionResults {
			w := runewidth.StringWidth(result)
			if lineWidth+w+2 > width && i <= s.CompletionIndex {
				line, lineWidth = "", 0
			}
			if s.CompletionIndex == i {
				pos = lineWidth
			}
			line += " " + result + " "
			lineWidth += w + 2
		}
		ui.setLine(height-2, 0, line+strings.Repeat(" ", width), tcell.StyleDefault.Reverse(true))
		if s.CompletionIndex >= 0 {
			ui.setLine(height-2, pos, " "+s.CompletionResults[s.CompletionIndex]+" ",
				tcell.StyleDefault.Foreground(tcell.ColorGrey).Reverse(true))
		}
	}
}

// Close terminates the Tui.
func (ui *Tui) Close() error {
	ui.eventCh = nil
	ui.screen.Fini()
	<-ui.waitCh
	return nil
}
