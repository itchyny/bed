package core

import (
	"io"
)

type Buffer struct {
	r io.Reader
}

func NewBuffer(r io.Reader) *Buffer {
	return &Buffer{r}
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	return b.r.Read(p)
}
