package core

import (
	"fmt"
	"os"
)

type Editor struct {
	buffer *Buffer
}

func NewEditor() *Editor {
	return &Editor{}
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
	height, width := 30, 16
	w := os.Stdout
	b := make([]byte, height*width)
	n, err := e.buffer.Read(b)
	if err != nil {
		return err
	}
	for i := 0; i < height; i++ {
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
