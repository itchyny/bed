package core

import (
	"io"
)

// Buffer represents a buffer.
type Buffer struct {
	r io.ReadSeeker
}

// NewBuffer creates a new buffer.
func NewBuffer(r io.ReadSeeker) *Buffer {
	return &Buffer{r}
}

// Read reads the bytes at the specified offset.
func (b *Buffer) Read(offset int64, p []byte) (n int, err error) {
	if _, err := b.r.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}
	return b.r.Read(p)
}

// Len returns the total size of the buffer.
func (b *Buffer) Len() (int64, error) {
	return b.r.Seek(0, io.SeekEnd)
}
