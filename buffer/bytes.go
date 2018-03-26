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

// ReadAt implements the io.ReaderAt interface.
func (r *bytesReader) ReadAt(b []byte, offset int64) (int, error) {
	if offset >= int64(len(r.bs)) {
		return 0, io.EOF
	}
	n := copy(b, r.bs[offset:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

func (r *bytesReader) appendByte(b byte) {
	r.bs = append(r.bs, b)
}

func (r *bytesReader) replaceByte(offset int64, b byte) {
	r.bs[offset] = b
}

func (r *bytesReader) deleteByte(offset int64) {
	copy(r.bs[offset:], r.bs[offset+1:])
	r.bs = r.bs[:len(r.bs)-1]
}
