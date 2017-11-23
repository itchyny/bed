package core

import (
	"io"
	"os"

	"github.com/itchyny/bed/util"
)

// Editor is the main struct for this command.
type Editor struct {
	ui     UI
	buffer *Buffer
	width  int
	offset int64
	cursor int64
	length int64
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
				len, err := e.buffer.Len()
				if err != nil {
					return
				}
				e.length = len
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
	len, err := e.buffer.Len()
	if err != nil {
		return err
	}
	e.length = len
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
		if e.cursor < e.offset {
			e.offset = e.offset - int64(e.width)
		}
	}
	return e.redraw()
}

func (e *Editor) cursorDown() error {
	if e.cursor < e.length-int64(e.width) {
		e.cursor += int64(e.width)
	}
	if e.cursor >= e.offset+int64(e.ui.Height()*e.width) {
		return e.scrollDown()
	}
	return e.redraw()
}

func (e *Editor) cursorLeft() error {
	if e.cursor%int64(e.width) > 0 {
		e.cursor--
	}
	return e.redraw()
}

func (e *Editor) cursorRight() error {
	if e.cursor < e.length-1 && int(e.cursor%int64(e.width)) < e.width-1 {
		e.cursor++
	}
	return e.redraw()
}

func (e *Editor) cursorPrev() error {
	if e.cursor > 0 {
		e.cursor--
		if e.cursor < e.offset {
			e.offset -= int64(e.width)
		}
	}
	return e.redraw()
}

func (e *Editor) cursorNext() error {
	if e.cursor < e.length-1 {
		e.cursor++
		if e.cursor >= e.offset+int64(e.ui.Height()*e.width) {
			e.offset += int64(e.width)
		}
	}
	return e.redraw()
}

func (e *Editor) scrollUp() error {
	if e.offset > 0 {
		e.offset -= int64(e.width)
	}
	if e.cursor >= e.offset+int64(e.ui.Height()*e.width) {
		e.cursor -= int64(e.width)
	}
	return e.redraw()
}

func (e *Editor) scrollDown() error {
	offset := util.MaxInt64(((e.length+int64(e.width)-1)/int64(e.width)-int64(e.ui.Height()))*int64(e.width), 0)
	e.offset = util.MinInt64(e.offset+int64(e.width), offset)
	if e.cursor < e.offset {
		e.cursor += int64(e.width)
	}
	return e.redraw()
}

func (e *Editor) pageUp() error {
	e.offset = util.MaxInt64(e.offset-int64((e.ui.Height()-2)*e.width), 0)
	if e.offset == 0 {
		e.cursor = 0
	} else if e.cursor >= e.offset+int64(e.ui.Height()*e.width) {
		e.cursor = e.offset + int64(e.ui.Height()-1)*int64(e.width)
	}
	return e.redraw()
}

func (e *Editor) pageDown() error {
	offset := util.MaxInt64(((e.length+int64(e.width)-1)/int64(e.width)-int64(e.ui.Height()))*int64(e.width), 0)
	e.offset = util.MinInt64(e.offset+int64((e.ui.Height()-2)*e.width), offset)
	if e.cursor < e.offset {
		e.cursor = e.offset
	} else if e.offset == offset {
		e.cursor = ((util.MaxInt64(e.length, 1)+int64(e.width)-1)/int64(e.width) - 1) * int64(e.width)
	}
	return e.redraw()
}

func (e *Editor) pageTop() error {
	e.offset = 0
	e.cursor = 0
	return e.redraw()
}

func (e *Editor) pageLast() error {
	e.offset = util.MaxInt64(((e.length+int64(e.width)-1)/int64(e.width)-int64(e.ui.Height()))*int64(e.width), 0)
	e.cursor = ((util.MaxInt64(e.length, 1)+int64(e.width)-1)/int64(e.width) - 1) * int64(e.width)
	return e.redraw()
}

func (e *Editor) redraw() error {
	b := make([]byte, e.ui.Height()*e.width)
	n, err := e.buffer.Read(e.offset, b)
	if err != nil && err != io.EOF {
		return err
	}
	return e.ui.Redraw(State{
		Width:  e.width,
		Offset: e.offset,
		Cursor: e.cursor,
		Bytes:  b,
		Size:   n,
		Length: e.length,
	})
}
