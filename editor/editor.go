package editor

import (
	"errors"
	"fmt"

	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/event"
)

// Editor is the main struct for this command.
type Editor struct {
	ui            UI
	wm            Manager
	cmdline       Cmdline
	mode          Mode
	searchTarget  string
	searchMode    rune
	prevEventType event.Type
	err           error
	errtyp        int
	eventCh       chan event.Event
	redrawCh      chan struct{}
	cmdlineCh     chan event.Event
}

// NewEditor creates a new editor.
func NewEditor(ui UI, wm Manager, cmdline Cmdline) *Editor {
	return &Editor{ui: ui, wm: wm, cmdline: cmdline, mode: ModeNormal}
}

// Init initializes the editor.
func (e *Editor) Init() error {
	e.eventCh = make(chan event.Event, 1)
	e.redrawCh = make(chan struct{})
	e.cmdlineCh = make(chan event.Event)
	e.cmdline.Init(e.eventCh, e.cmdlineCh, e.redrawCh)
	e.wm.Init(e.eventCh, e.redrawCh)
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
		if ev.Type != event.Redraw {
			e.prevEventType = ev.Type
		}
		switch ev.Type {
		case event.QuitAll:
			if len(ev.Arg) > 0 {
				e.err, e.errtyp = fmt.Errorf("too many arguments for %s", ev.CmdName), MessageError
				e.redrawCh <- struct{}{}
			} else {
				return
			}
		case event.Info:
			e.err, e.errtyp = ev.Error, MessageInfo
			e.redrawCh <- struct{}{}
		case event.Error:
			e.err, e.errtyp = ev.Error, MessageError
			e.redrawCh <- struct{}{}
		case event.Redraw:
			width, height := e.ui.Size()
			e.wm.Resize(width, height-1)
			e.redrawCh <- struct{}{}
		default:
			switch ev.Type {
			case event.StartInsert, event.StartInsertHead, event.StartAppend, event.StartAppendEnd:
				e.mode = ModeInsert
			case event.StartReplaceByte, event.StartReplace:
				e.mode = ModeReplace
			case event.ExitInsert:
				e.mode = ModeNormal
			case event.StartCmdlineCommand:
				e.mode = ModeCmdline
				e.err = nil
			case event.StartCmdlineSearchForward:
				e.mode = ModeSearch
				e.err = nil
				e.searchMode = '/'
			case event.StartCmdlineSearchBackward:
				e.mode = ModeSearch
				e.err = nil
				e.searchMode = '?'
			case event.ExitCmdline:
				e.mode = ModeNormal
			case event.ExecuteCmdline:
				e.mode = ModeNormal
			case event.ExecuteSearch:
				e.searchTarget, e.searchMode = ev.Arg, ev.Rune
			case event.NextSearch:
				ev.Arg, ev.Rune = e.searchTarget, e.searchMode
			case event.PreviousSearch:
				ev.Arg, ev.Rune = e.searchTarget, e.searchMode
			}
			if e.mode == ModeCmdline || e.mode == ModeSearch ||
				ev.Type == event.ExitCmdline || ev.Type == event.ExecuteCmdline {
				e.cmdlineCh <- ev
			} else {
				ev.Mode = e.mode
				width, height := e.ui.Size()
				e.wm.Resize(width, height-1)
				e.wm.Emit(ev)
			}
		}
	}
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
	go e.wm.Run()
	e.listen()
	return nil
}

func (e *Editor) redraw() (err error) {
	var state State
	var windowIndex int
	state.WindowStates, state.Layout, windowIndex, err = e.wm.State()
	if err != nil {
		return err
	}
	if state.WindowStates[windowIndex] == nil {
		return errors.New("index out of windows")
	}
	state.WindowStates[windowIndex].Mode = e.mode
	state.Mode, state.Error, state.ErrorType = e.mode, e.err, e.errtyp
	state.Cmdline, state.CmdlineCursor, state.CompletionResults, state.CompletionIndex = e.cmdline.Get()
	if e.mode == ModeSearch || e.prevEventType == event.ExecuteSearch {
		state.SearchMode = e.searchMode
	} else if e.prevEventType == event.NextSearch {
		state.SearchMode, state.Cmdline = e.searchMode, []rune(e.searchTarget)
	} else if e.prevEventType == event.PreviousSearch {
		if e.searchMode == '/' {
			state.SearchMode, state.Cmdline = '?', []rune(e.searchTarget)
		} else {
			state.SearchMode, state.Cmdline = '/', []rune(e.searchTarget)
		}
	}
	return e.ui.Redraw(state)
}

// Close terminates the editor.
func (e *Editor) Close() error {
	close(e.eventCh)
	close(e.redrawCh)
	close(e.cmdlineCh)
	e.wm.Close()
	return e.ui.Close()
}
