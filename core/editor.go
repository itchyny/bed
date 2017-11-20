package core

import (
	"bytes"
	"fmt"
	"os"
	"strings"
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
				case PageTop:
					e.PageTop()
				case PageLast:
					e.PageLast()
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
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	e.line = e.line + 1
	if e.line > line {
		e.line = line
	}
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
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	e.line = e.line + e.height - 2
	if e.line > line {
		e.line = line
	}
	return e.Redraw()
}

func (e *Editor) PageTop() error {
	e.line = 0
	return e.Redraw()
}

func (e *Editor) PageLast() error {
	line, err := e.lastLine()
	if err != nil {
		return err
	}
	e.line = line
	return e.Redraw()
}

func (e *Editor) lastLine() (int, error) {
	len, err := e.buffer.Len()
	if err != nil {
		return 0, err
	}
	line := int((len+int64(e.width)-1)/int64(e.width)) - e.height
	if line < 0 {
		line = 0
	}
	return line, nil
}

func (e *Editor) Redraw() error {
	width := e.width
	b := make([]byte, e.height*width)
	n, err := e.buffer.Read(int64(e.line)*int64(width), b)
	if err != nil {
		return err
	}
	for i := 0; i < e.height; i++ {
		if i*width >= n {
			e.ui.SetLine(i, strings.Repeat(" ", 75))
			continue
		}
		w := new(bytes.Buffer)
		fmt.Fprintf(w, "%08x:", int64(e.line+i)*int64(width))
		buf := make([]byte, width)
		for j := 0; j < width; j++ {
			k := i*width + j
			if k >= n {
				fmt.Fprintf(w, "   ")
				continue
			}
			fmt.Fprintf(w, " %02x", b[k])
			buf[j] = prettyByte(b[k])
		}
		fmt.Fprintf(w, "  %s\n", buf)
		e.ui.SetLine(i, w.String())
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
