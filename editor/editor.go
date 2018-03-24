package editor

import (
	"errors"
	"fmt"

	. "github.com/itchyny/bed/common"
)

// Editor is the main struct for this command.
type Editor struct {
	ui            UI
	wm            Manager
	cmdline       Cmdline
	mode          Mode
	searchTarget  string
	searchMode    rune
	prevEventType EventType
	err           error
	errtyp        int
	eventCh       chan Event
	redrawCh      chan struct{}
	cmdlineCh     chan Event
}

// NewEditor creates a new editor.
func NewEditor(ui UI, wm Manager, cmdline Cmdline) *Editor {
	return &Editor{ui: ui, wm: wm, cmdline: cmdline, mode: ModeNormal}
}

// Init initializes the editor.
func (e *Editor) Init() error {
	e.eventCh = make(chan Event, 1)
	e.redrawCh = make(chan struct{})
	e.cmdlineCh = make(chan Event)
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
	for event := range e.eventCh {
		if event.Type != EventRedraw {
			e.prevEventType = event.Type
		}
		switch event.Type {
		case EventQuitAll:
			if len(event.Arg) > 0 {
				e.err, e.errtyp = fmt.Errorf("too many arguments for %s", event.CmdName), MessageError
				e.redrawCh <- struct{}{}
			} else {
				return
			}
		case EventInfo:
			e.err, e.errtyp = event.Error, MessageInfo
			e.redrawCh <- struct{}{}
		case EventError:
			e.err, e.errtyp = event.Error, MessageError
			e.redrawCh <- struct{}{}
		case EventRedraw:
			width, height := e.ui.Size()
			e.wm.Resize(width, height-1)
			e.redrawCh <- struct{}{}
		default:
			switch event.Type {
			case EventStartInsert, EventStartInsertHead, EventStartAppend, EventStartAppendEnd:
				e.mode = ModeInsert
			case EventStartReplaceByte, EventStartReplace:
				e.mode = ModeReplace
			case EventExitInsert:
				e.mode = ModeNormal
			case EventStartCmdlineCommand:
				e.mode = ModeCmdline
				e.err = nil
			case EventStartCmdlineSearchForward:
				e.mode = ModeSearch
				e.err = nil
				e.searchMode = '/'
			case EventStartCmdlineSearchBackward:
				e.mode = ModeSearch
				e.err = nil
				e.searchMode = '?'
			case EventExitCmdline:
				e.mode = ModeNormal
			case EventExecuteCmdline:
				e.mode = ModeNormal
			case EventExecuteSearch:
				e.searchTarget, e.searchMode = event.Arg, event.Rune
			case EventNextSearch:
				event.Arg, event.Rune = e.searchTarget, e.searchMode
			case EventPreviousSearch:
				event.Arg, event.Rune = e.searchTarget, e.searchMode
			}
			if e.mode == ModeCmdline || e.mode == ModeSearch ||
				event.Type == EventExitCmdline || event.Type == EventExecuteCmdline {
				e.cmdlineCh <- event
			} else {
				event.Mode = e.mode
				width, height := e.ui.Size()
				e.wm.Resize(width, height-1)
				e.wm.Emit(event)
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
	if e.mode == ModeSearch || e.prevEventType == EventExecuteSearch {
		state.SearchMode = e.searchMode
	} else if e.prevEventType == EventNextSearch {
		state.SearchMode, state.Cmdline = e.searchMode, []rune(e.searchTarget)
	} else if e.prevEventType == EventPreviousSearch {
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
