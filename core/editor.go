package core

// Editor is the main struct for this command.
type Editor struct {
	ui     UI
	window *Window
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
				default:
					continue
				}
				e.redraw()
			}
		}
	}()
	return nil
}

func defaultKeyManager() *KeyManager {
	km := NewKeyManager()
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
	return km
}

// Close terminates the editor.
func (e *Editor) Close() error {
	if e.window != nil {
		_ = e.window.Close()
	}
	return e.ui.Close()
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	if e.window, err = NewWindow(filename, 16); err != nil {
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
	return e.ui.Start(defaultKeyManager())
}

func (e *Editor) redraw() error {
	state, err := e.window.State()
	if err != nil {
		return err
	}
	return e.ui.Redraw(state)
}
