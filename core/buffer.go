package core

import (
	"io"
	"path/filepath"

	"github.com/itchyny/bed/util"
)

// Buffer represents a buffer.
type Buffer struct {
	r        io.ReadSeeker
	name     string
	basename string
	height   int64
	width    int64
	offset   int64
	cursor   int64
	length   int64
}

// NewBuffer creates a new buffer.
func NewBuffer(name string, r io.ReadSeeker, width int64) *Buffer {
	return &Buffer{r: r, name: name, basename: filepath.Base(name), width: width}
}

// Seek sets the offset.
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	return b.r.Seek(offset, whence)
}

// Read reads the bytes.
func (b *Buffer) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

// Len returns the total size of the buffer.
func (b *Buffer) Len() (int64, error) {
	return b.r.Seek(0, io.SeekEnd)
}

// ReadBytes reads the bytes at the current offset.
func (b *Buffer) ReadBytes() (int, []byte, error) {
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
