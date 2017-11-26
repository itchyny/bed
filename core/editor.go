package core

// Editor is the main struct for this command.
type Editor struct {
	ui     UI
	buffer *Buffer
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
			case c := <-ch:
				e.buffer.height = int64(e.ui.Height())
				switch c {
				case CursorUp:
					e.buffer.cursorUp()
				case CursorDown:
					e.buffer.cursorDown()
				case CursorLeft:
					e.buffer.cursorLeft()
				case CursorRight:
					e.buffer.cursorRight()
				case CursorPrev:
					e.buffer.cursorPrev()
				case CursorNext:
					e.buffer.cursorNext()
				case CursorHead:
					e.buffer.cursorHead()
				case CursorEnd:
					e.buffer.cursorEnd()
				case ScrollUp:
					e.buffer.scrollUp()
				case ScrollDown:
					e.buffer.scrollDown()
				case PageUp:
					e.buffer.pageUp()
				case PageDown:
					e.buffer.pageDown()
				case PageUpHalf:
					e.buffer.pageUpHalf()
				case PageDownHalf:
					e.buffer.pageDownHalf()
				case PageTop:
					e.buffer.pageTop()
				case PageEnd:
					e.buffer.pageEnd()
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
	return km
}

// Close terminates the editor.
func (e *Editor) Close() error {
	_ = e.buffer.Close()
	return e.ui.Close()
}

// Open opens a new file.
func (e *Editor) Open(filename string) (err error) {
	if e.buffer, err = NewBuffer(filename, 16); err != nil {
		return err
	}
	e.buffer.height = int64(e.ui.Height())
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
	n, bytes, err := e.buffer.ReadBytes()
	if err != nil {
		return err
	}
	return e.ui.Redraw(State{
		Name:   e.buffer.basename,
		Width:  int(e.buffer.width),
		Offset: e.buffer.offset,
		Cursor: e.buffer.cursor,
		Bytes:  bytes,
		Size:   n,
		Length: e.buffer.length,
	})
}
