package core

import (
	"io"
)

type Buffer struct {
	r io.ReadSeeker
}

func NewBuffer(r io.ReadSeeker) *Buffer {
	return &Buffer{r}
}

func (b *Buffer) Read(offset int64, p []byte) (n int, err error) {
	if _, err := b.r.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}
	return b.r.Read(p)
}
