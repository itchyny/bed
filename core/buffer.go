package core

import (
	"io"
)

// Buffer represents a buffer.
type Buffer struct {
	r ReadSeekCloser
}

// ReadSeekCloser is the interface that groups the basic Read, Seek and Close methods.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// NewBuffer creates a new buffer.
func NewBuffer(r ReadSeekCloser) *Buffer {
	return &Buffer{r}
}

// Read reads bytes.
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

// Len returns the total size of the buffer.
func (b *Buffer) Len() (int64, error) {
	return b.r.Seek(0, io.SeekEnd)
}
