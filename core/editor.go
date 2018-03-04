package core

import (
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
	ch := make(chan Event, 1)
	quit := make(chan struct{})
	if err := e.ui.Init(ch, quit); err != nil {
		return err
	}
	if err := e.cmdline.Init(ch); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event := <-ch:
				e.window.height = int64(e.ui.Height())
				switch event.Type {
				case EventQuit:
					if len(event.Args) > 0 {
						e.err = fmt.Errorf("too many arguments for %s", event.CmdName)
					} else {
						quit <- struct{}{}
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
				case EventCursorLeftCmdline:
					e.cmdline.CursorLeft()
				case EventCursorRightCmdline:
					e.cmdline.CursorRight()
				case EventCursorHeadCmdline:
					e.cmdline.CursorHead()
				case EventCursorEndCmdline:
					e.cmdline.CursorEnd()
				case EventBackspaceCmdline:
					e.cmdline.Backspace()
				case EventDeleteCmdline:
					e.cmdline.Delete()
				case EventDeleteWordCmdline:
					e.cmdline.DeleteWord()
				case EventClearToHeadCmdline:
					e.cmdline.ClearToHead()
				case EventClearCmdline:
					e.cmdline.Clear()
				case EventExitCmdline:
					e.mode = ModeNormal
				case EventExecuteCmdline:
					e.mode = ModeNormal
					e.cmdline.Execute()
				case EventSpaceCmdline:
					event.Rune = ' '
					fallthrough
				case EventRune:
					if e.mode == ModeCmdline {
						e.cmdline.Insert(event.Rune)
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
				case EventWriteQuit:
					if len(event.Args) > 0 {
						e.err = fmt.Errorf("too many arguments for %s", event.CmdName)
					} else {
						e.err = e.writeFile("")
						quit <- struct{}{}
					}
				case EventError:
					e.err = event.Error
				default:
					continue
				}
				e.redraw()
			}
		}
	}()
	return nil
}

func defaultKeyManagers() map[Mode]*KeyManager {
	kms := make(map[Mode]*KeyManager)
	km := NewKeyManager(true)
	km.Register(EventQuit, "Z", "Q")
	km.Register(EventCursorUp, "up")
	km.Register(EventCursorDown, "down")
	km.Register(EventCursorLeft, "left")
	km.Register(EventCursorRight, "right")
	km.Register(EventPageUp, "pgup")
	km.Register(EventPageDown, "pgdn")
	km.Register(EventPageTop, "home")
	km.Register(EventPageEnd, "end")
	km.Register(EventCursorUp, "k")
	km.Register(EventCursorDown, "j")
	km.Register(EventCursorLeft, "h")
	km.Register(EventCursorRight, "l")
	km.Register(EventCursorPrev, "b")
	km.Register(EventCursorNext, "w")
	km.Register(EventCursorHead, "0")
	km.Register(EventCursorHead, "^")
	km.Register(EventCursorEnd, "$")
	km.Register(EventScrollUp, "c-y")
	km.Register(EventScrollDown, "c-e")
	km.Register(EventPageUp, "c-b")
	km.Register(EventPageDown, "c-f")
	km.Register(EventPageUpHalf, "c-u")
	km.Register(EventPageDownHalf, "c-d")
	km.Register(EventPageTop, "g", "g")
	km.Register(EventPageEnd, "G")
	km.Register(EventJumpTo, "c-]")
	km.Register(EventJumpBack, "c-t")
	km.Register(EventDeleteByte, "x")
	km.Register(EventDeletePrevByte, "X")
	km.Register(EventIncrement, "c-a")
	km.Register(EventIncrement, "+")
	km.Register(EventDecrement, "c-x")
	km.Register(EventDecrement, "-")

	km.Register(EventStartInsert, "i")
	km.Register(EventStartInsertHead, "I")
	km.Register(EventStartAppend, "a")
	km.Register(EventStartAppendEnd, "A")
	km.Register(EventStartReplaceByte, "r")
	km.Register(EventStartReplace, "R")

	km.Register(EventStartCmdline, ":")
	kms[ModeNormal] = km

	km = NewKeyManager(false)
	km.Register(EventExitInsert, "escape")
	km.Register(EventExitInsert, "c-c")
	km.Register(EventCursorUp, "up")
	km.Register(EventCursorDown, "down")
	km.Register(EventCursorLeft, "left")
	km.Register(EventCursorRight, "right")
	km.Register(EventPageUp, "pgup")
	km.Register(EventPageDown, "pgdn")
	km.Register(EventPageTop, "home")
	km.Register(EventPageEnd, "end")
	km.Register(EventInsert0, "0")
	km.Register(EventInsert1, "1")
	km.Register(EventInsert2, "2")
	km.Register(EventInsert3, "3")
	km.Register(EventInsert4, "4")
	km.Register(EventInsert5, "5")
	km.Register(EventInsert6, "6")
	km.Register(EventInsert7, "7")
	km.Register(EventInsert8, "8")
	km.Register(EventInsert9, "9")
	km.Register(EventInsertA, "a")
	km.Register(EventInsertB, "b")
	km.Register(EventInsertC, "c")
	km.Register(EventInsertD, "d")
	km.Register(EventInsertE, "e")
	km.Register(EventInsertF, "f")
	km.Register(EventBackspace, "backspace")
	km.Register(EventBackspace, "backspace2")
	km.Register(EventDelete, "delete")
	kms[ModeInsert] = km
	kms[ModeReplace] = km

	km = NewKeyManager(false)
	km.Register(EventSpaceCmdline, "space")
	km.Register(EventCursorLeftCmdline, "left")
	km.Register(EventCursorLeftCmdline, "c-b")
	km.Register(EventCursorRightCmdline, "right")
	km.Register(EventCursorRightCmdline, "c-f")
	km.Register(EventCursorHeadCmdline, "home")
	km.Register(EventCursorHeadCmdline, "c-a")
	km.Register(EventCursorEndCmdline, "end")
	km.Register(EventCursorEndCmdline, "c-e")
	km.Register(EventBackspaceCmdline, "c-h")
	km.Register(EventBackspaceCmdline, "backspace")
	km.Register(EventBackspaceCmdline, "backspace2")
	km.Register(EventDeleteCmdline, "delete")
	km.Register(EventDeleteWordCmdline, "c-w")
	km.Register(EventClearToHeadCmdline, "c-u")
	km.Register(EventClearCmdline, "c-k")
	km.Register(EventExitCmdline, "escape")
	km.Register(EventExitCmdline, "c-c")
	km.Register(EventExecuteCmdline, "enter")
	km.Register(EventExecuteCmdline, "c-j")
	km.Register(EventExecuteCmdline, "c-m")
	kms[ModeCmdline] = km
	return kms
}

// Close terminates the editor.
func (e *Editor) Close() error {
	for _, f := range e.files {
		f.file.Close()
	}
	return e.ui.Close()
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
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

// Start starts the editor.
func (e *Editor) Start() error {
	if err := e.redraw(); err != nil {
		return err
	}
	return e.ui.Start(defaultKeyManagers())
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
	for _, f := range e.files {
		if f.name == name {
			perm = f.perm
		}
	}
	tmpf, err := os.OpenFile(
		name+"-"+strconv.FormatUint(rand.Uint64(), 16), os.O_RDWR|os.O_CREATE, perm,
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
