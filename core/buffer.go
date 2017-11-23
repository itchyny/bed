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
