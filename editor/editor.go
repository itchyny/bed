package editor

import (
	"errors"
	"fmt"
	"sync"

	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/state"
)

// Editor is the main struct for this command.
type Editor struct {
	ui            UI
	wm            Manager
	cmdline       Cmdline
	mode          mode.Mode
	prevMode      mode.Mode
	searchTarget  string
	searchMode    rune
	prevEventType event.Type
	err           error
	errtyp        int
	eventCh       chan event.Event
	redrawCh      chan struct{}
	cmdlineCh     chan event.Event
	mu            *sync.Mutex
}

// NewEditor creates a new editor.
func NewEditor(ui UI, wm Manager, cmdline Cmdline) *Editor {
	return &Editor{
		ui:       ui,
		wm:       wm,
		cmdline:  cmdline,
		mode:     mode.Normal,
		prevMode: mode.Normal,
	}
}

// Init initializes the editor.
func (e *Editor) Init() error {
	e.eventCh = make(chan event.Event, 1)
	e.redrawCh = make(chan struct{})
	e.cmdlineCh = make(chan event.Event)
	e.cmdline.Init(e.eventCh, e.cmdlineCh, e.redrawCh)
	e.wm.Init(e.eventCh, e.redrawCh)
	e.mu = new(sync.Mutex)
	return nil
}

func (e *Editor) listen() {
	go func() {
		for {
			<-e.redrawCh
			e.redraw()
		}
	}()
	for ev := range e.eventCh {
		if redraw, finish := e.emit(ev); redraw {
			e.redrawCh <- struct{}{}
		} else if finish {
			break
		}
	}
}

func (e *Editor) emit(ev event.Event) (redraw bool, finish bool) {
	e.mu.Lock()
	if ev.Type != event.Redraw {
		e.prevEventType = ev.Type
	}
	switch ev.Type {
	case event.QuitAll:
		if len(ev.Arg) > 0 {
			e.err, e.errtyp = fmt.Errorf("too many arguments for %s", ev.CmdName), state.MessageError
			redraw = true
		} else {
			finish = true
		}
	case event.Suspend:
		if len(ev.Arg) > 0 {
			e.err, e.errtyp = fmt.Errorf("too many arguments for %s", ev.CmdName), state.MessageError
		} else {
			e.mu.Unlock()
			if err := e.suspend(); err != nil {
				e.mu.Lock()
				e.err, e.errtyp = err, state.MessageError
				e.mu.Unlock()
			}
			redraw = true
			return
		}
		redraw = true
	case event.Info:
		e.err, e.errtyp = ev.Error, state.MessageInfo
		redraw = true
	case event.Error:
		e.err, e.errtyp = ev.Error, state.MessageError
		redraw = true
	case event.Redraw:
		width, height := e.ui.Size()
		e.wm.Resize(width, height-1)
		redraw = true
	default:
		switch ev.Type {
		case event.StartInsert, event.StartInsertHead, event.StartAppend, event.StartAppendEnd:
			e.mode, e.prevMode = mode.Insert, e.mode
		case event.StartReplaceByte, event.StartReplace:
			e.mode, e.prevMode = mode.Replace, e.mode
		case event.ExitInsert:
			e.mode, e.prevMode = mode.Normal, e.mode
		case event.StartVisual:
			e.mode, e.prevMode = mode.Visual, e.mode
		case event.ExitVisual:
			e.mode, e.prevMode = mode.Normal, e.mode
		case event.StartCmdlineCommand:
			if e.mode == mode.Visual {
				ev.Arg = "'<,'>"
			} else if ev.Count > 0 {
				ev.Arg = fmt.Sprintf(".,.+%d", ev.Count-1)
			}
			e.mode, e.prevMode = mode.Cmdline, e.mode
			e.err = nil
		case event.StartCmdlineSearchForward:
			e.mode, e.prevMode = mode.Search, e.mode
			e.err = nil
			e.searchMode = '/'
		case event.StartCmdlineSearchBackward:
			e.mode, e.prevMode = mode.Search, e.mode
			e.err = nil
			e.searchMode = '?'
		case event.ExitCmdline:
			e.mode, e.prevMode = mode.Normal, e.mode
		case event.ExecuteCmdline:
			e.mode, e.prevMode = mode.Normal, e.mode
		case event.ExecuteSearch:
			e.searchTarget, e.searchMode = ev.Arg, ev.Rune
		case event.NextSearch:
			ev.Arg, ev.Rune = e.searchTarget, e.searchMode
		case event.PreviousSearch:
			ev.Arg, ev.Rune = e.searchTarget, e.searchMode
		}
		if e.mode == mode.Cmdline || e.mode == mode.Search ||
			ev.Type == event.ExitCmdline || ev.Type == event.ExecuteCmdline {
			e.mu.Unlock()
			e.cmdlineCh <- ev
		} else {
			if event.ScrollUp <= ev.Type && ev.Type <= event.SwitchFocus {
				e.prevMode, e.err = e.mode, nil
			}
			ev.Mode = e.mode
			width, height := e.ui.Size()
			e.wm.Resize(width, height-1)
			e.mu.Unlock()
			e.wm.Emit(ev)
		}
		return
	}
	e.mu.Unlock()
	return
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	return e.wm.Open(filename)
}

// OpenEmpty creates a new window.
func (e *Editor) OpenEmpty() (err error) {
	return e.wm.Open("")
}

// Run the editor.
func (e *Editor) Run() error {
	if err := e.ui.Init(e.eventCh); err != nil {
		return err
	}
	if err := e.redraw(); err != nil {
		return err
	}
	go e.ui.Run(defaultKeyManagers())
	go e.cmdline.Run()
	e.listen()
	return nil
}

func (e *Editor) redraw() (err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var s state.State
	var windowIndex int
	s.WindowStates, s.Layout, windowIndex, err = e.wm.State()
	if err != nil {
		return err
	}
	if s.WindowStates[windowIndex] == nil {
		return errors.New("index out of windows")
	}
	s.WindowStates[windowIndex].Mode = e.mode
	s.Mode, s.PrevMode, s.Error, s.ErrorType = e.mode, e.prevMode, e.err, e.errtyp
	if s.Mode != mode.Visual && s.PrevMode != mode.Visual {
		for _, ws := range s.WindowStates {
			ws.VisualStart = -1
		}
	}
	s.Cmdline, s.CmdlineCursor, s.CompletionResults, s.CompletionIndex = e.cmdline.Get()
	if e.mode == mode.Search || e.prevEventType == event.ExecuteSearch {
		s.SearchMode = e.searchMode
	} else if e.prevEventType == event.NextSearch {
		s.SearchMode, s.Cmdline = e.searchMode, []rune(e.searchTarget)
	} else if e.prevEventType == event.PreviousSearch {
		if e.searchMode == '/' {
			s.SearchMode, s.Cmdline = '?', []rune(e.searchTarget)
		} else {
			s.SearchMode, s.Cmdline = '/', []rune(e.searchTarget)
		}
	}
	return e.ui.Redraw(s)
}

func (e *Editor) suspend() error {
	return suspend(e)
}

// Close terminates the editor.
func (e *Editor) Close() error {
	close(e.eventCh)
	close(e.redrawCh)
	close(e.cmdlineCh)
	e.wm.Close()
	return e.ui.Close()
}
