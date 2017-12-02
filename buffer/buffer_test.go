package buffer

import (
	"io"
	"strings"
	"testing"
)

type stringReader struct {
	*strings.Reader
}

func NewStringReader(str string) *stringReader {
	return &stringReader{strings.NewReader(str)}
}

func (b *stringReader) Close() error {
	return nil
}

func TestBufferEmpty(t *testing.T) {
	b := NewBuffer(NewStringReader(""))

	p := make([]byte, 10)
	n, err := b.Read(p)
	if err != io.EOF {
		t.Errorf("err should be EOF but got: %v", err)
	}
	if n != 0 {
		t.Errorf("n should be 0 but got: %d", n)
	}

	l, err := b.Len()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if l != 0 {
		t.Errorf("l should be 0 but got: %d", l)
	}
}

func TestBuffer(t *testing.T) {
	b := NewBuffer(NewStringReader("0123456789abcdef"))

	p := make([]byte, 8)
	n, err := b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "01234567" {
		t.Errorf("p should be 01234567 but got: %s", string(p))
	}

	l, err := b.Len()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if l != 16 {
		t.Errorf("l should be 16 but got: %d", l)
	}

	_, err = b.Seek(4, io.SeekStart)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err = b.Read(p)
	if err != nil {
		t.Errorf("err should be EOF but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "456789ab" {
		t.Errorf("p should be 456789ab but got: %s", string(p))
	}

	_, err = b.Seek(-4, io.SeekEnd)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err = b.Read(p)
	if err != nil {
		t.Errorf("err should be EOF but got: %v", err)
	}
	if n != 4 {
		t.Errorf("n should be 4 but got: %d", n)
	}
	if string(p) != "cdef89ab" {
		t.Errorf("p should be cdef89ab but got: %s", string(p))
	}

	err = b.Close()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestBufferInsertHead(t *testing.T) {
	b := NewBuffer(NewStringReader("0123456789abcdef"))

	err := b.Insert(0, 0x39)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	p := make([]byte, 8)
	n, err := b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "90123456" {
		t.Errorf("p should be 90123456 but got: %s", string(p))
	}

	err = b.Insert(0, 0x38)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	_, err = b.Seek(0, io.SeekStart)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err = b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "89012345" {
		t.Errorf("p should be 89012345 but got: %s", string(p))
	}

	l, err := b.Len()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if l != 18 {
		t.Errorf("l should be 18 but got: %d", l)
	}

	if len(b.rrs) != 2 {
		t.Errorf("len(b.rrs) should be 2 but got: %d", len(b.rrs))
	}
}

func TestBufferInsertMiddle(t *testing.T) {
	b := NewBuffer(NewStringReader("0123456789abcdef"))

	p := make([]byte, 8)
	err := b.Insert(4, 0x37)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	_, err = b.Seek(0, io.SeekStart)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err := b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "01237456" {
		t.Errorf("p should be 01237456 but got: %s", string(p))
	}

	err = b.Insert(8, 0x30)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	_, err = b.Seek(3, io.SeekStart)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err = b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "37456078" {
		t.Errorf("p should be 37456078 but got: %s", string(p))
	}

	err = b.Insert(9, 0x31)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	_, err = b.Seek(3, io.SeekStart)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err = b.Read(p)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "37456017" {
		t.Errorf("p should be 37456017 but got: %s", string(p))
	}

	l, err := b.Len()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if l != 19 {
		t.Errorf("l should be 19 but got: %d", l)
	}

	if len(b.rrs) != 5 {
		t.Errorf("len(b.rrs) should be 5 but got: %d", len(b.rrs))
	}
}

func TestBufferInsertLast(t *testing.T) {
	b := NewBuffer(NewStringReader("0123456789abcdef"))

	p := make([]byte, 8)
	err := b.Insert(16, 0x39)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	_, err = b.Seek(9, io.SeekStart)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}

	n, err := b.Read(p)
	if err != io.EOF {
		t.Errorf("err should be io.EOF but got: %v", err)
	}
	if n != 8 {
		t.Errorf("n should be 8 but got: %d", n)
	}
	if string(p) != "9abcdef9" {
		t.Errorf("p should be 9abcdef9 but got: %s", string(p))
	}

	l, err := b.Len()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if l != 17 {
		t.Errorf("l should be 17 but got: %d", l)
	}

	if len(b.rrs) != 3 {
		t.Errorf("len(b.rrs) should be 3 but got: %d", len(b.rrs))
	}
}
