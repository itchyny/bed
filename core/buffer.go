package core

import (
	"io"
	"os"
	"path/filepath"

	"github.com/itchyny/bed/util"
)

// Buffer represents a buffer.
type Buffer struct {
	r        ReadSeekCloser
	name     string
	basename string
	height   int64
	width    int64
	offset   int64
	cursor   int64
	length   int64
}

// ReadSeekCloser is the interface that groups the basic Read, Seek and Close methods.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// NewBuffer creates a new buffer.
func NewBuffer(name string, width int64) (*Buffer, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	b := &Buffer{
		r:        file,
		name:     name,
		basename: filepath.Base(name),
		width:    width,
	}
	b.length, err = b.r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	return b, err
}

// Read reads the bytes.
func (b *Buffer) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

// Seek sets the offset.
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	return b.r.Seek(offset, whence)
}

// Close the buffer.
func (b *Buffer) Close() error {
	return b.r.Close()
}

func (b *Buffer) readBytes() (int, []byte, error) {
	bytes := make([]byte, int(b.height*b.width))
	_, err := b.Seek(b.offset, io.SeekStart)
	if err != nil {
		return 0, bytes, err
	}
	n, err := b.Read(bytes)
	if err != nil && err != io.EOF {
		return 0, bytes, err
	}
	return n, bytes, nil
}

// State returns the current state of the buffer.
func (b *Buffer) State() (State, error) {
	n, bytes, err := b.readBytes()
	if err != nil {
		return State{}, err
	}
	return State{
		Name:   b.basename,
		Width:  int(b.width),
		Offset: b.offset,
		Cursor: b.cursor,
		Bytes:  bytes,
		Size:   n,
		Length: b.length,
	}, nil
}

func (b *Buffer) cursorUp() {
	if b.cursor >= b.width {
		b.cursor -= b.width
		if b.cursor < b.offset {
			b.offset = b.offset - b.width
		}
	}
}

func (b *Buffer) cursorDown() {
	if b.cursor < b.length-b.width {
		b.cursor += b.width
	} else if b.cursor < ((util.MaxInt64(b.length, 1)+b.width-1)/b.width-1)*b.width {
		b.cursor = b.length - 1
	}
	if b.cursor >= b.offset+b.height*b.width {
		b.scrollDown()
	}
}

func (b *Buffer) cursorLeft() {
	if b.cursor%b.width > 0 {
		b.cursor--
	}
}

func (b *Buffer) cursorRight() {
	if b.cursor < b.length-1 && b.cursor%b.width < b.width-1 {
		b.cursor++
	}
}

func (b *Buffer) cursorPrev() {
	if b.cursor > 0 {
		b.cursor--
		if b.cursor < b.offset {
			b.offset -= b.width
		}
	}
}

func (b *Buffer) cursorNext() {
	if b.cursor < b.length-1 {
		b.cursor++
		if b.cursor >= b.offset+b.height*b.width {
			b.offset += b.width
		}
	}
}

func (b *Buffer) cursorHead() {
	b.cursor -= b.cursor % b.width
}

func (b *Buffer) cursorEnd() {
	b.cursor = util.MinInt64((b.cursor/b.width+1)*b.width-1, b.length-1)
}

func (b *Buffer) scrollUp() {
	if b.offset > 0 {
		b.offset -= b.width
	}
	if b.cursor >= b.offset+b.height*b.width {
		b.cursor -= b.width
	}
}

func (b *Buffer) scrollDown() {
	offset := util.MaxInt64(((b.length+b.width-1)/b.width-b.height)*b.width, 0)
	b.offset = util.MinInt64(b.offset+b.width, offset)
	if b.cursor < b.offset {
		b.cursor += b.width
	}
}

func (b *Buffer) pageUp() {
	b.offset = util.MaxInt64(b.offset-(b.height-2)*b.width, 0)
	if b.offset == 0 {
		b.cursor = 0
	} else if b.cursor >= b.offset+b.height*b.width {
		b.cursor = b.offset + (b.height-1)*b.width
	}
}

func (b *Buffer) pageDown() {
	offset := util.MaxInt64(((b.length+b.width-1)/b.width-b.height)*b.width, 0)
	b.offset = util.MinInt64(b.offset+(b.height-2)*b.width, offset)
	if b.cursor < b.offset {
		b.cursor = b.offset
	} else if b.offset == offset {
		b.cursor = ((util.MaxInt64(b.length, 1)+b.width-1)/b.width - 1) * b.width
	}
}

func (b *Buffer) pageUpHalf() {
	b.offset = util.MaxInt64(b.offset-util.MaxInt64(b.height/2, 1)*b.width, 0)
	if b.offset == 0 {
		b.cursor = 0
	} else if b.cursor >= b.offset+b.height*b.width {
		b.cursor = b.offset + (b.height-1)*b.width
	}
}

func (b *Buffer) pageDownHalf() {
	offset := util.MaxInt64(((b.length+b.width-1)/b.width-b.height)*b.width, 0)
	b.offset = util.MinInt64(b.offset+util.MaxInt64(b.height/2, 1)*b.width, offset)
	if b.cursor < b.offset {
		b.cursor = b.offset
	} else if b.offset == offset {
		b.cursor = ((util.MaxInt64(b.length, 1)+b.width-1)/b.width - 1) * b.width
	}
}

func (b *Buffer) pageTop() {
	b.offset = 0
	b.cursor = 0
}

func (b *Buffer) pageEnd() {
	b.offset = util.MaxInt64(((b.length+b.width-1)/b.width-b.height)*b.width, 0)
	b.cursor = ((util.MaxInt64(b.length, 1)+b.width-1)/b.width - 1) * b.width
}
