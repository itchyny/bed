package core

import (
	"os"
)

// Editor is the main struct for this command.
type Editor struct {
	ui     UI
	buffer *Buffer
	line   int64
	width  int
	cursor int64
}

// NewEditor creates a new editor.
func NewEditor(ui UI) *Editor {
	return &Editor{ui: ui, width: 16}
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
				switch c {
				case CursorUp:
					e.cursorUp()
				case CursorDown:
					e.cursorDown()
				case CursorLeft:
					e.cursorLeft()
				case CursorRight:
					e.cursorRight()
				case CursorPrev:
					e.cursorPrev()
				case CursorNext:
					e.cursorNext()
				case ScrollUp:
					e.scrollUp()
				case ScrollDown:
					e.scrollDown()
				case PageUp:
					e.pageUp()
				case PageDown:
					e.pageDown()
				case PageTop:
					e.pageTop()
				case PageLast:
					e.pageLast()
				}
			}
		}
	}()
	return nil
}

// Close terminates the editor.
func (e *Editor) Close() error {
	return e.ui.Close()
}

// Open opens a new file.
func (e *Editor) Open(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	e.buffer = NewBuffer(file)
	return nil
}

// Start starts the editor.
func (e *Editor) Start() error {
	if err := e.redraw(); err != nil {
		return err
	}
	return e.ui.Start()
}

func (e *Editor) cursorUp() error {
	if e.cursor >= int64(e.width) {
		e.cursor -= int64(e.width)
		if e.cursor < e.line*int64(e.width) {
			e.line = e.line - 1
		}
	}
	return e.redraw()
}

func (e *Editor) cursorDown() error {
	e.cursor += int64(e.width)
	if e.cursor >= (e.line+int64(e.ui.Height()))*int64(e.width) {
		return e.scrollDown()
	}
	return e.redraw()
}

func (e *Editor) cursorLeft() error {
	if e.cursor%int64(e.width) > 0 {
		e.cursor -= 1
	}
	return e.redraw()
}

func (e *Editor) cursorRight() error {
	len, err := e.buffer.Len()
	if err != nil {
		return err
	}
	if e.cursor < len-1 && int(e.cursor%int64(e.width)) < e.width-1 {
		e.cursor += 1
	}
	return e.redraw()
}

func (e *Editor) cursorPrev() error {
	if e.cursor > 0 {
		e.cursor -= 1
		if e.cursor < e.line*int64(e.width) {
			e.line -= 1
		}
	}
	return e.redraw()
}

func (e *Editor) cursorNext() error {
	len, err := e.buffer.Len()
	if err != nil {
		return err
	}
	if e.cursor < len-1 {
		e.cursor += 1
		if e.cursor >= (e.line+int64(e.ui.Height()))*int64(e.width) {
			e.line += 1
		}
	}
	return e.redraw()
}

func (e *Editor) scrollUp() error {
	if e.line > 0 {
		e.line = e.line - 1
	}
	if e.cursor > (e.line+int64(e.ui.Height()))*int64(e.width) {
		e.cursor -= int64(e.width)
	}
	return e.redraw()
}

func (e *Editor) scrollDown() error {
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	e.line = e.line + 1
	if e.line > line {
		e.line = line
	}
	if e.cursor < e.line*int64(e.width) {
		e.cursor += int64(e.width)
	}
	return e.redraw()
}

func (e *Editor) pageUp() error {
	e.line = e.line - int64(e.ui.Height()) + 2
	if e.line < 0 {
		e.line = 0
	}
	if e.line == 0 {
		e.cursor = 0
	} else if e.cursor >= (e.line+int64(e.ui.Height()))*int64(e.width) {
		e.cursor = (e.line + int64(e.ui.Height()) - 1) * int64(e.width)
	}
	return e.redraw()
}

func (e *Editor) pageDown() error {
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	len, err := e.buffer.Len()
	if err != nil {
		return err
	}
	e.line = e.line + int64(e.ui.Height()) - 2
	if e.line > line {
		e.line = line
	}
	if e.cursor < e.line*int64(e.width) {
		e.cursor = e.line * int64(e.width)
	} else if e.line == line {
		e.cursor = ((len+int64(e.width)-1)/int64(e.width) - 1) * int64(e.width)
	}
	return e.redraw()
}

func (e *Editor) pageTop() error {
	e.line = 0
	e.cursor = 0
	return e.redraw()
}

func (e *Editor) pageLast() error {
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	len, err := e.buffer.Len()
	if err != nil {
		return err
	}
	e.line = line
	e.cursor = ((len+int64(e.width)-1)/int64(e.width) - 1) * int64(e.width)
	return e.redraw()
}

func (e *Editor) lastLine() (int64, error) {
	len, err := e.buffer.Len()
	if err != nil {
		return 0, err
	}
	line := (len+int64(e.width)-1)/int64(e.width) - int64(e.ui.Height())
	if line < 0 {
		line = 0
	}
	return line, nil
}

func (e *Editor) redraw() error {
	b := make([]byte, e.ui.Height()*e.width)
	n, err := e.buffer.Read(e.line*int64(e.width), b)
	if err != nil {
		return err
	}
	return e.ui.Redraw(State{Line: e.line, Width: e.width, Cursor: e.cursor, Bytes: b, Size: n})
}
