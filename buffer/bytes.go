package buffer

import (
	"bytes"
)

type bytesReader struct {
	*bytes.Reader
}

func NewBytesReader(bs []byte) *bytesReader {
	return &bytesReader{bytes.NewReader(bs)}
}

func (b *bytesReader) Close() error {
	return nil
}
