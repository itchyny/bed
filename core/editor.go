package core

// Editor is the main struct for this command.
type Editor struct {
	ui     UI
	window *Window
	files  []*File
}

// NewEditor creates a new editor.
func NewEditor(ui UI) *Editor {
	return &Editor{ui: ui}
}

// Init initializes the editor.
func (e *Editor) Init() error {
	ch := make(chan Event)
	if err := e.ui.Init(ch); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case event := <-ch:
				e.window.height = int64(e.ui.Height())
				switch event.Type {
				case CursorUp:
					e.window.cursorUp(event.Count)
				case CursorDown:
					e.window.cursorDown(event.Count)
				case CursorLeft:
					e.window.cursorLeft(event.Count)
				case CursorRight:
					e.window.cursorRight(event.Count)
				case CursorPrev:
					e.window.cursorPrev(event.Count)
				case CursorNext:
					e.window.cursorNext(event.Count)
				case CursorHead:
					e.window.cursorHead(event.Count)
				case CursorEnd:
					e.window.cursorEnd(event.Count)
				case ScrollUp:
					e.window.scrollUp(event.Count)
				case ScrollDown:
					e.window.scrollDown(event.Count)
				case PageUp:
					e.window.pageUp()
				case PageDown:
					e.window.pageDown()
				case PageUpHalf:
					e.window.pageUpHalf()
				case PageDownHalf:
					e.window.pageDownHalf()
				case PageTop:
					e.window.pageTop()
				case PageEnd:
					e.window.pageEnd()
				case JumpTo:
					e.window.jumpTo()
				case JumpBack:
					e.window.jumpBack()
				case DeleteByte:
					e.window.deleteByte(event.Count)
				case DeletePrevByte:
					e.window.deletePrevByte(event.Count)

				case StartInsert:
					e.window.startInsert()
				case StartInsertHead:
					e.window.startInsertHead()
				case StartAppend:
					e.window.startAppend()
				case StartAppendEnd:
					e.window.startAppendEnd()
				case StartReplaceByte:
					e.window.startReplaceByte()
				case StartReplace:
					e.window.startReplace()
				case ExitInsert:
					e.window.exitInsert()
				case Insert0:
					e.window.insert0()
				case Insert1:
					e.window.insert1()
				case Insert2:
					e.window.insert2()
				case Insert3:
					e.window.insert3()
				case Insert4:
					e.window.insert4()
				case Insert5:
					e.window.insert5()
				case Insert6:
					e.window.insert6()
				case Insert7:
					e.window.insert7()
				case Insert8:
					e.window.insert8()
				case Insert9:
					e.window.insert9()
				case InsertA:
					e.window.insertA()
				case InsertB:
					e.window.insertB()
				case InsertC:
					e.window.insertC()
				case InsertD:
					e.window.insertD()
				case InsertE:
					e.window.insertE()
				case InsertF:
					e.window.insertF()
				case Backspace:
					e.window.backspace()
				case Delete:
					e.window.deleteByte(1)
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
	km.Register(Quit, "Z", "Q")
	km.Register(CursorUp, "up")
	km.Register(CursorDown, "down")
	km.Register(CursorLeft, "left")
	km.Register(CursorRight, "right")
	km.Register(PageUp, "pgup")
	km.Register(PageDown, "pgdn")
	km.Register(PageTop, "home")
	km.Register(PageEnd, "end")
	km.Register(CursorUp, "k")
	km.Register(CursorDown, "j")
	km.Register(CursorLeft, "h")
	km.Register(CursorRight, "l")
	km.Register(CursorPrev, "b")
	km.Register(CursorNext, "w")
	km.Register(CursorHead, "0")
	km.Register(CursorHead, "^")
	km.Register(CursorEnd, "$")
	km.Register(ScrollUp, "c-y")
	km.Register(ScrollDown, "c-e")
	km.Register(PageUp, "c-b")
	km.Register(PageDown, "c-f")
	km.Register(PageUpHalf, "c-u")
	km.Register(PageDownHalf, "c-d")
	km.Register(PageTop, "g", "g")
	km.Register(PageEnd, "G")
	km.Register(JumpTo, "c-]")
	km.Register(JumpBack, "c-t")
	km.Register(DeleteByte, "x")
	km.Register(DeletePrevByte, "X")

	km.Register(StartInsert, "i")
	km.Register(StartInsertHead, "I")
	km.Register(StartAppend, "a")
	km.Register(StartAppendEnd, "A")
	km.Register(StartReplaceByte, "r")
	km.Register(StartReplace, "R")
	kms[ModeNormal] = km

	km = NewKeyManager(false)
	km.Register(ExitInsert, "escape")
	km.Register(ExitInsert, "c-c")
	km.Register(CursorUp, "up")
	km.Register(CursorDown, "down")
	km.Register(CursorLeft, "left")
	km.Register(CursorRight, "right")
	km.Register(PageUp, "pgup")
	km.Register(PageDown, "pgdn")
	km.Register(PageTop, "home")
	km.Register(PageEnd, "end")
	km.Register(Insert0, "0")
	km.Register(Insert1, "1")
	km.Register(Insert2, "2")
	km.Register(Insert3, "3")
	km.Register(Insert4, "4")
	km.Register(Insert5, "5")
	km.Register(Insert6, "6")
	km.Register(Insert7, "7")
	km.Register(Insert8, "8")
	km.Register(Insert9, "9")
	km.Register(InsertA, "a")
	km.Register(InsertB, "b")
	km.Register(InsertC, "c")
	km.Register(InsertD, "d")
	km.Register(InsertE, "e")
	km.Register(InsertF, "f")
	km.Register(Backspace, "backspace")
	km.Register(Backspace, "backspace2")
	km.Register(Delete, "delete")
	kms[ModeInsert] = km
	kms[ModeReplace] = km
	return kms
}

// Close terminates the editor.
func (e *Editor) Close() error {
	for _, f := range e.files {
		f.Close()
	}
	return e.ui.Close()
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	f, err := NewFile(filename)
	if err != nil {
		return err
	}
	e.files = append(e.files, f)
	if e.window, err = NewWindow(f, 16); err != nil {
		return err
	}
	e.window.height = int64(e.ui.Height())
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
	return e.ui.Redraw(state)
}
