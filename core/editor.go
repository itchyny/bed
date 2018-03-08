package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
)

// Editor is the main struct for this command.
type Editor struct {
	ui        UI
	window    *Window
	files     []file
	mode      Mode
	cmdline   Cmdline
	err       error
	eventCh   chan Event
	redrawCh  chan struct{}
	cmdlineCh chan Event
	quitCh    chan struct{}
}

type file struct {
	name string
	file *os.File
	perm os.FileMode
}

// NewEditor creates a new editor.
func NewEditor(ui UI, cmdline Cmdline) *Editor {
	return &Editor{ui: ui, cmdline: cmdline}
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
	return e.cmdline.Init(e.eventCh, e.cmdlineCh, e.redrawCh)
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
		case EventWrite:
			if len(event.Args) > 1 {
				e.err = fmt.Errorf("too many arguments for %s", event.CmdName)
			} else {
				var name string
				if len(event.Args) > 0 {
					name = event.Args[0]
				}
				e.err = e.writeFile(name)
			}
			e.redrawCh <- struct{}{}
		case EventWriteQuit:
			if len(event.Args) > 0 {
				e.err = fmt.Errorf("too many arguments for %s", event.CmdName)
				e.redrawCh <- struct{}{}
			} else {
				e.err = e.writeFile("")
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
				e.window.height = int64(e.ui.Height())
				event.Mode = e.mode
				e.window.eventCh <- event
			}
		}
	}
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if e.window, err = NewWindow(bytes.NewReader(nil), filename, filepath.Base(filename), int64(e.ui.Height()), 16, e.redrawCh); err != nil {
			return err
		}
		return nil
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	e.files = append(e.files, file{name: filename, file: f, perm: info.Mode().Perm()})
	if e.window, err = NewWindow(f, filename, filepath.Base(filename), int64(e.ui.Height()), 16, e.redrawCh); err != nil {
		return err
	}
	return nil
}

// OpenEmpty creates a new window.
func (e *Editor) OpenEmpty() (err error) {
	if e.window, err = NewWindow(bytes.NewReader(nil), "", "", int64(e.ui.Height()), 16, e.redrawCh); err != nil {
		return err
	}
	return nil
}

// Run the editor.
func (e *Editor) Run() error {
	if err := e.redraw(); err != nil {
		return err
	}
	go e.ui.Run(defaultKeyManagers())
	go e.cmdline.Run()
	go e.window.Run()
	e.listen()
	return nil
}

func (e *Editor) redraw() error {
	state, err := e.window.State()
	if err != nil {
		return err
	}
	state.Mode, state.Error = e.mode, e.err
	state.Cmdline, state.CmdlineCursor = e.cmdline.Get()
	return e.ui.Redraw(state)
}

// Close terminates the editor.
func (e *Editor) Close() error {
	for _, f := range e.files {
		f.file.Close()
	}
	close(e.eventCh)
	close(e.redrawCh)
	close(e.quitCh)
	close(e.cmdlineCh)
	close(e.window.eventCh)
	return e.ui.Close()
}

func (e *Editor) writeFile(name string) error {
	perm := os.FileMode(0644)
	if name == "" {
		name = e.window.filename
	}
	if name == "" {
		return errors.New("no file name")
	}
	if e.window.filename == "" {
		e.window.filename = name
	}
	for _, f := range e.files {
		if f.name == name {
			perm = f.perm
		}
	}
	tmpf, err := os.OpenFile(
		name+"-"+strconv.FormatUint(rand.Uint64(), 16), os.O_RDWR|os.O_CREATE|os.O_EXCL, perm,
	)
	if err != nil {
		return err
	}
	defer os.Remove(tmpf.Name())
	e.window.buffer.Seek(0, io.SeekStart)
	_, err = io.Copy(tmpf, e.window.buffer)
	tmpf.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpf.Name(), name)
}
