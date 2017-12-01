package buffer

import (
	"io"
	"strings"
	"testing"
)

type StringBuffer struct {
	*strings.Reader
}

func NewStringBuffer(str string) *StringBuffer {
	return &StringBuffer{strings.NewReader(str)}
}

func (b *StringBuffer) Close() error {
	return nil
}

func TestBufferEmpty(t *testing.T) {
	b := NewBuffer(NewStringBuffer(""))

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
	b := NewBuffer(NewStringBuffer("0123456789abcdef"))

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
