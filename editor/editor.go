package editor

import (
	"errors"
	"fmt"

	. "github.com/itchyny/bed/common"
)

// Editor is the main struct for this command.
type Editor struct {
	ui        UI
	wm        Manager
	cmdline   Cmdline
	mode      Mode
	err       error
	errtyp    int
	eventCh   chan Event
	redrawCh  chan struct{}
	cmdlineCh chan Event
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
	if err := e.ui.Init(e.eventCh); err != nil {
		return err
	}
	if err := e.cmdline.Init(e.eventCh, e.cmdlineCh, e.redrawCh); err != nil {
		return err
	}
	return e.wm.Init(e.eventCh, e.redrawCh)
}

func (e *Editor) listen() {
	go func() {
		for {
			<-e.redrawCh
			e.redraw()
		}
	}()
	for event := range e.eventCh {
		switch event.Type {
		case EventQuitAll:
			if len(event.Args) > 0 {
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
			e.redrawCh <- struct{}{}
		default:
			switch event.Type {
			case EventStartInsert, EventStartInsertHead, EventStartAppend, EventStartAppendEnd:
				e.mode = ModeInsert
			case EventStartReplaceByte, EventStartReplace:
				e.mode = ModeReplace
			case EventExitInsert:
				e.mode = ModeNormal
			case EventStartCmdline:
				e.mode = ModeCmdline
				e.err = nil
			case EventExitCmdline, EventExecuteCmdline:
				e.mode = ModeNormal
			}
			if e.mode == ModeCmdline || event.Type == EventExitCmdline || event.Type == EventExecuteCmdline {
				e.cmdlineCh <- event
			} else {
				event.Mode = e.mode
				width, height := e.ui.Size()
				e.wm.Resize(width, height)
				e.wm.Emit(event)
			}
		}
	}
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	width, height := e.ui.Size()
	e.wm.SetSize(width, height)
	return e.wm.Open(filename)
}

// OpenEmpty creates a new window.
func (e *Editor) OpenEmpty() (err error) {
	width, height := e.ui.Size()
	e.wm.SetSize(width, height)
	return e.wm.Open("")
}

// Run the editor.
func (e *Editor) Run() error {
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
	var index int
	state.Windows, state.Layout, index, err = e.wm.State()
	if index < 0 || len(state.Windows) <= index {
		return errors.New("index out of windows")
	}
	state.Windows[index].Mode = e.mode
	if err != nil {
		return err
	}
	state.Mode, state.Error, state.ErrorType = e.mode, e.err, e.errtyp
	state.Cmdline, state.CmdlineCursor = e.cmdline.Get()
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
