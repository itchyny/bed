package core

import (
	"bytes"
	"fmt"
	"os"
)

// Editor is the main struct for this command.
type Editor struct {
	ui     UI
	buffer *Buffer
	line   int
	height int
	width  int
}

// NewEditor creates a new Editor.
func NewEditor(ui UI) *Editor {
	return &Editor{ui: ui, line: 0, width: 16}
}

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
				case ScrollDown:
					e.ScrollDown()
				case ScrollUp:
					e.ScrollUp()
				case PageUp:
					e.PageUp()
				case PageDown:
					e.PageDown()
				}
			}
		}
	}()
	return nil
}

func (e *Editor) Close() error {
	return e.ui.Close()
}

func (e *Editor) Open(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	e.buffer = NewBuffer(file)
	return nil
}

func (e *Editor) Start() error {
	return e.ui.Start(func(height, _ int) error {
		e.height = height
		return e.Redraw()
	})
}

func (e *Editor) ScrollUp() error {
	if e.line > 0 {
		e.line = e.line - 1
	}
	return e.Redraw()
}

func (e *Editor) ScrollDown() error {
	e.line = e.line + 1
	return e.Redraw()
}

func (e *Editor) PageUp() error {
	e.line = e.line - e.height + 2
	if e.line < 0 {
		e.line = 0
	}
	return e.Redraw()
}

func (e *Editor) PageDown() error {
	e.line = e.line + e.height - 2
	return e.Redraw()
}

func (e *Editor) Redraw() error {
	width := e.width
	b := make([]byte, e.height*width)
	n, err := e.buffer.Read(int64(e.line)*int64(width), b)
	if err != nil {
		return err
	}
	for i := 0; i < e.height; i++ {
		w := new(bytes.Buffer)
		fmt.Fprintf(w, "%08x:", (e.line+i)*width)
		buf := make([]byte, width)
		for j := 0; j < width; j++ {
			k := i*width + j
			if k >= n {
				break
			}
			fmt.Fprintf(w, " %02x", b[k])
			buf[j] = prettyByte(b[k])
		}
		fmt.Fprintf(w, "  %s\n", buf)
		e.ui.SetLine(i, w.String())
		if (i+1)*width >= n {
			break
		}
	}
	return nil
}

func prettyByte(b byte) byte {
	switch {
	case 0x20 <= b && b < 0x7f:
		return b
	default:
		return 0x2e
	}
}
