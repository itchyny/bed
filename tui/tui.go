package tui

import (
	"bytes"
	"strings"
	"sync"

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
	mu      *sync.Mutex
}

// NewTui creates a new Tui.
func NewTui() *Tui {
	return &Tui{mu: new(sync.Mutex)}
}

// Init initializes the Tui.
func (ui *Tui) Init(eventCh chan<- event.Event) (err error) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
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
			var e event.Event
			if km, ok := kms[ui.getMode()]; ok {
				e = km.Press(eventToKey(ev))
			}
			if e.Type != event.Nop {
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

func (ui *Tui) getMode() mode.Mode {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	return ui.mode
}

// Size returns the size for the screen.
func (ui *Tui) Size() (int, int) {
	return ui.screen.Size()
}

// Redraw redraws the state.
func (ui *Tui) Redraw(s state.State) error {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.mode = s.Mode
	ui.screen.Clear()
	ui.drawWindows(s.WindowStates, s.Layout)
	ui.drawCmdline(s)
	ui.screen.Show()
	return nil
}

func (ui *Tui) setLine(line, offset int, str string, style tcell.Style) {
	for _, c := range str {
		ui.screen.SetContent(offset, line, c, nil, style)
		offset += runewidth.RuneWidth(c)
	}
}

func (ui *Tui) drawWindows(windowStates map[int]*state.WindowState, l layout.Layout) {
	switch l := l.(type) {
	case layout.Window:
		r := fromLayout(l)
		if ws, ok := windowStates[l.Index]; ok && r.valid() {
			ui.newTuiWindow(r).drawWindow(ws,
				l.Active && ui.mode != mode.Cmdline && ui.mode != mode.Search)
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
	for i := range region.height {
		ui.setLine(region.top+i, region.left+region.width, "|", tcell.StyleDefault.Reverse(true))
	}
}

func (ui *Tui) drawCmdline(s state.State) {
	var cmdline string
	style := tcell.StyleDefault
	width, height := ui.Size()
	switch {
	case s.Error != nil:
		cmdline = s.Error.Error()
		if s.ErrorType == state.MessageInfo {
			style = style.Foreground(tcell.ColorYellow)
		} else {
			style = style.Foreground(tcell.ColorRed)
		}
	case s.Mode == mode.Cmdline:
		if len(s.CompletionResults) > 0 {
			ui.drawCompletionResults(s.CompletionResults, s.CompletionIndex, width, height)
		}
		ui.screen.ShowCursor(1+runewidth.StringWidth(string(s.Cmdline[:s.CmdlineCursor])), height-1)
		fallthrough
	case s.PrevMode == mode.Cmdline && len(s.Cmdline) > 0:
		cmdline = ":" + string(s.Cmdline)
	case s.Mode == mode.Search:
		ui.screen.ShowCursor(1+runewidth.StringWidth(string(s.Cmdline[:s.CmdlineCursor])), height-1)
		fallthrough
	case s.SearchMode != '\x00':
		cmdline = string(s.SearchMode) + string(s.Cmdline)
	default:
		return
	}
	ui.setLine(height-1, 0, cmdline, style)
}

func (ui *Tui) drawCompletionResults(results []string, index, width, height int) {
	var line bytes.Buffer
	var left, right int
	for i, result := range results {
		size := runewidth.StringWidth(result) + 2
		if i <= index {
			left, right = right, right+size
			if right > width {
				line.Reset()
				left, right = 0, size
			}
		} else if right < width {
			right += size
		} else {
			break
		}
		line.WriteString(" ")
		line.WriteString(result)
		line.WriteString(" ")
	}
	line.WriteString(strings.Repeat(" ", max(width-right, 0)))
	ui.setLine(height-2, 0, line.String(), tcell.StyleDefault.Reverse(true))
	if index >= 0 {
		ui.setLine(height-2, left, " "+results[index]+" ",
			tcell.StyleDefault.Foreground(tcell.ColorGrey).Reverse(true))
	}
}

// Close terminates the Tui.
func (ui *Tui) Close() error {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.eventCh = nil
	ui.screen.Fini()
	<-ui.waitCh
	return nil
}
