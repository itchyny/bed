package editor

import (
	"fmt"

	. "github.com/itchyny/bed/core"
)

// Editor is the main struct for this command.
type Editor struct {
	ui        UI
	wm        Manager
	cmdline   Cmdline
	mode      Mode
	err       error
	eventCh   chan Event
	redrawCh  chan struct{}
	cmdlineCh chan Event
	quitCh    chan struct{}
}

// NewEditor creates a new editor.
func NewEditor(ui UI, wm Manager, cmdline Cmdline) *Editor {
	return &Editor{ui: ui, wm: wm, cmdline: cmdline, mode: ModeNormal}
}

// Init initializes the editor.
func (e *Editor) Init() error {
	e.eventCh = make(chan Event, 1)
	e.redrawCh = make(chan struct{})
	e.quitCh = make(chan struct{})
	e.cmdlineCh = make(chan Event)
	if err := e.ui.Init(e.eventCh, e.quitCh); err != nil {
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
		case EventQuit:
			if len(event.Args) > 0 {
				e.err = fmt.Errorf("too many arguments for %s", event.CmdName)
				e.redrawCh <- struct{}{}
			} else {
				e.quitCh <- struct{}{}
				return
			}
		case EventError:
			e.err = event.Error
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
				e.wm.SetHeight(e.ui.Height())
				e.wm.Emit(event)
			}
		}
	}
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	e.wm.SetHeight(e.ui.Height())
	return e.wm.Open(filename)
}

// OpenEmpty creates a new window.
func (e *Editor) OpenEmpty() (err error) {
	e.wm.SetHeight(e.ui.Height())
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

func (e *Editor) redraw() error {
	state, err := e.wm.State()
	if err != nil {
		return err
	}
	state.Mode, state.Error = e.mode, e.err
	state.Cmdline, state.CmdlineCursor = e.cmdline.Get()
	return e.ui.Redraw(state)
}

// Close terminates the editor.
func (e *Editor) Close() error {
	close(e.eventCh)
	close(e.redrawCh)
	close(e.quitCh)
	close(e.cmdlineCh)
	e.wm.Close()
	return e.ui.Close()
}
