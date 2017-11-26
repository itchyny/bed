package core

import (
	"io"
	"path/filepath"
)

// Buffer represents a buffer.
type Buffer struct {
	r        io.ReadSeeker
	name     string
	basename string
}

// NewBuffer creates a new buffer.
func NewBuffer(name string, r io.ReadSeeker) *Buffer {
	return &Buffer{r: r, name: name, basename: filepath.Base(name)}
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
