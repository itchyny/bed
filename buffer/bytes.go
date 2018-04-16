package buffer

import (
	"errors"
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
func (r *bytesReader) ReadAt(b []byte, offset int64) (n int, err error) {
	if offset < 0 {
		return 0, errors.New("buffer.bytesReader.ReadAt: negative offset")
	}
	if offset >= int64(len(r.bs)) {
		return 0, io.EOF
	}
	n = copy(b, r.bs[offset:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

func (r *bytesReader) replaceByte(offset int64, b byte) {
	l := int64(len(r.bs))
	switch {
	case offset == l:
		r.bs = append(r.bs, b)
	case 0 <= offset && offset < l:
		r.bs[offset] = b
	default:
		panic("buffer.bytesReader.replaceByte: unreachable")
	}
}

func (r *bytesReader) deleteByte(offset int64) {
	copy(r.bs[offset:], r.bs[offset+1:])
	r.bs = r.bs[:len(r.bs)-1]
}

func (r *bytesReader) clone() *bytesReader {
	bs := make([]byte, len(r.bs))
	copy(bs, r.bs)
	return newBytesReader(bs)
}
