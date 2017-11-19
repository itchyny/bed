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
}

// NewEditor creates a new Editor.
func NewEditor(ui UI) *Editor {
	return &Editor{ui: ui, buffer: nil}
}

func (e *Editor) Init() error {
	return e.ui.Init()
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
		width := 16
		b := make([]byte, height*width)
		n, err := e.buffer.Read(b)
		if err != nil {
			return err
		}
		for i := 0; i < height; i++ {
			w := new(bytes.Buffer)
			fmt.Fprintf(w, "%08x:", i*width)
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
	})
}

func prettyByte(b byte) byte {
	switch {
	case 0x20 <= b && b < 0x7f:
		return b
	default:
		return 0x2e
	}
}
