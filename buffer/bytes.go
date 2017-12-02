package buffer

import (
	"io"
)

type bytesReader struct {
	bs    []byte
	index int64
}

func newBytesReader(bs []byte) *bytesReader {
	return &bytesReader{bs: bs, index: 0}
}

// Read implements the io.Reader interface.
func (r *bytesReader) Read(b []byte) (n int, err error) {
	if r.index >= int64(len(r.bs)) {
		return 0, io.EOF
	}
	n = copy(b, r.bs[r.index:])
	r.index += int64(n)
	return
}

// Seek implements the io.Seeker interface.
func (r *bytesReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.index = offset
	case io.SeekCurrent:
		r.index += offset
	case io.SeekEnd:
		r.index = int64(len(r.bs)) + offset
	}
	return r.index, nil
}

// Close implements the io.Closer interface.
func (r *bytesReader) Close() error {
	return nil
}

func (r *bytesReader) insertByte(offset int64, b byte) {
	r.bs = append(r.bs, 0)
	copy(r.bs[offset+1:], r.bs[offset:])
	r.bs[offset] = b
}

func (r *bytesReader) appendByte(b byte) {
	r.bs = append(r.bs, b)
}

func (r *bytesReader) replaceByte(offset int64, b byte) {
	r.bs[offset] = b
}
