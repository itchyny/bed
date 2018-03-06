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
	ui            UI
	window        *Window
	files         []file
	mode          Mode
	cmdline       Cmdline
	cmdlineCursor int
	err           error
	eventCh       chan Event
	cmdlineCh     chan Event
	quitCh        chan struct{}
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
	e.quitCh = make(chan struct{})
	e.cmdlineCh = make(chan Event, 1)
	if err := e.ui.Init(e.eventCh, e.quitCh); err != nil {
		return err
	}
	return e.cmdline.Init(e.eventCh, e.cmdlineCh)
}

func (e *Editor) listen() {
	for event := range e.eventCh {
		e.window.height = int64(e.ui.Height())
		switch event.Type {
		case EventQuit:
			if len(event.Args) > 0 {
				e.err = fmt.Errorf("too many arguments for %s", event.CmdName)
			} else {
				e.quitCh <- struct{}{}
				return
			}
		case EventCursorUp:
			e.window.cursorUp(event.Count)
		case EventCursorDown:
			e.window.cursorDown(event.Count)
		case EventCursorLeft:
			e.window.cursorLeft(event.Count)
		case EventCursorRight:
			e.window.cursorRight(event.Count)
		case EventCursorPrev:
			e.window.cursorPrev(event.Count)
		case EventCursorNext:
			e.window.cursorNext(event.Count)
		case EventCursorHead:
			e.window.cursorHead(event.Count)
		case EventCursorEnd:
			e.window.cursorEnd(event.Count)
		case EventScrollUp:
			e.window.scrollUp(event.Count)
		case EventScrollDown:
			e.window.scrollDown(event.Count)
		case EventPageUp:
			e.window.pageUp()
		case EventPageDown:
			e.window.pageDown()
		case EventPageUpHalf:
			e.window.pageUpHalf()
		case EventPageDownHalf:
			e.window.pageDownHalf()
		case EventPageTop:
			e.window.pageTop()
		case EventPageEnd:
			e.window.pageEnd()
		case EventJumpTo:
			e.window.jumpTo()
		case EventJumpBack:
			e.window.jumpBack()
		case EventDeleteByte:
			e.window.deleteByte(event.Count)
		case EventDeletePrevByte:
			e.window.deletePrevByte(event.Count)
		case EventIncrement:
			e.window.increment(event.Count)
		case EventDecrement:
			e.window.decrement(event.Count)

		case EventStartInsert:
			e.mode = ModeInsert
			e.window.startInsert()
		case EventStartInsertHead:
			e.mode = ModeInsert
			e.window.startInsertHead()
		case EventStartAppend:
			e.mode = ModeInsert
			e.window.startAppend()
		case EventStartAppendEnd:
			e.mode = ModeInsert
			e.window.startAppendEnd()
		case EventStartReplaceByte:
			e.mode = ModeReplace
			e.window.startReplaceByte()
		case EventStartReplace:
			e.mode = ModeReplace
			e.window.startReplace()
		case EventExitInsert:
			e.mode = ModeNormal
			e.window.exitInsert()
		case EventInsert0:
			e.window.insert0(e.mode)
		case EventInsert1:
			e.window.insert1(e.mode)
		case EventInsert2:
			e.window.insert2(e.mode)
		case EventInsert3:
			e.window.insert3(e.mode)
		case EventInsert4:
			e.window.insert4(e.mode)
		case EventInsert5:
			e.window.insert5(e.mode)
		case EventInsert6:
			e.window.insert6(e.mode)
		case EventInsert7:
			e.window.insert7(e.mode)
		case EventInsert8:
			e.window.insert8(e.mode)
		case EventInsert9:
			e.window.insert9(e.mode)
		case EventInsertA:
			e.window.insertA(e.mode)
		case EventInsertB:
			e.window.insertB(e.mode)
		case EventInsertC:
			e.window.insertC(e.mode)
		case EventInsertD:
			e.window.insertD(e.mode)
		case EventInsertE:
			e.window.insertE(e.mode)
		case EventInsertF:
			e.window.insertF(e.mode)
		case EventBackspace:
			e.window.backspace()
		case EventDelete:
			e.window.deleteByte(1)

		case EventStartCmdline:
			e.mode = ModeCmdline
			e.err = nil
			e.cmdline.Clear()
		case EventExitCmdline:
			e.mode = ModeNormal
		case EventExecuteCmdline:
			e.mode = ModeNormal
			e.cmdline.Execute()
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
		case EventWriteQuit:
			if len(event.Args) > 0 {
				e.err = fmt.Errorf("too many arguments for %s", event.CmdName)
			} else {
				e.err = e.writeFile("")
				e.quitCh <- struct{}{}
				return
			}
		case EventError:
			e.err = event.Error
		case EventRedraw:
		default:
			if e.mode == ModeCmdline {
				e.cmdlineCh <- event
			} else {
				continue
			}
		}
		e.redraw()
	}
}

// Close terminates the editor.
func (e *Editor) Close() error {
	for _, f := range e.files {
		f.file.Close()
	}
	close(e.eventCh)
	close(e.quitCh)
	close(e.cmdlineCh)
	return e.ui.Close()
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if e.window, err = NewWindow(bytes.NewReader(nil), filename, filepath.Base(filename), int64(e.ui.Height()), 16); err != nil {
			return err
		}
		return nil
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	e.files = append(e.files, file{name: filename, file: f, perm: info.Mode().Perm()})
	if e.window, err = NewWindow(f, filename, filepath.Base(filename), int64(e.ui.Height()), 16); err != nil {
		return err
	}
	return nil
}

// OpenEmpty creates a new window.
func (e *Editor) OpenEmpty() (err error) {
	if e.window, err = NewWindow(bytes.NewReader(nil), "", "", int64(e.ui.Height()), 16); err != nil {
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
