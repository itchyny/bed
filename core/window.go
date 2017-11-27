package core

import (
	"io"
	"os"
	"path/filepath"

	"github.com/itchyny/bed/util"
)

// Window represents an editor window.
type Window struct {
	buffer   *Buffer
	name     string
	basename string
	height   int64
	width    int64
	offset   int64
	cursor   int64
	length   int64
}

// NewWindow creates a new editor window.
func NewWindow(name string, width int64) (*Window, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	buffer := NewBuffer(file)
	length, err := buffer.Len()
	if err != nil {
		return nil, err
	}
	return &Window{
		buffer:   buffer,
		name:     name,
		basename: filepath.Base(name),
		width:    width,
		length:   length,
	}, nil
}

func (w *Window) readBytes() (int, []byte, error) {
	bytes := make([]byte, int(w.height*w.width))
	_, err := w.buffer.Seek(w.offset, io.SeekStart)
	if err != nil {
		return 0, bytes, err
	}
	n, err := w.buffer.Read(bytes)
	if err != nil && err != io.EOF {
		return 0, bytes, err
	}
	return n, bytes, nil
}

// State returns the current state of the buffer.
func (w *Window) State() (State, error) {
	n, bytes, err := w.readBytes()
	if err != nil {
		return State{}, err
	}
	return State{
		Name:   w.basename,
		Width:  int(w.width),
		Offset: w.offset,
		Cursor: w.cursor,
		Bytes:  bytes,
		Size:   n,
		Length: w.length,
	}, nil
}

// Close the window.
func (w *Window) Close() error {
	return w.buffer.Close()
}

func (w *Window) cursorUp(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor/w.width) * w.width
	if w.cursor < w.offset {
		w.offset = w.cursor / w.width * w.width
	}
}

func (w *Window) cursorDown(count int64) {
	w.cursor += util.MinInt64(
		util.MinInt64(util.MaxInt64(count, 1), (util.MaxInt64(w.length, 1)-1)/w.width-w.cursor/w.width)*w.width,
		util.MaxInt64(w.length, 1)-1-w.cursor)
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *Window) cursorLeft(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor%w.width)
}

func (w *Window) cursorRight(count int64) {
	w.cursor += util.MinInt64(util.MinInt64(util.MaxInt64(count, 1), w.width-1-w.cursor%w.width), util.MaxInt64(w.length, 1)-1-w.cursor)
}

func (w *Window) cursorPrev(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor)
	if w.cursor < w.offset {
		w.offset = w.cursor / w.width * w.width
	}
}

func (w *Window) cursorNext(count int64) {
	w.cursor += util.MinInt64(util.MaxInt64(count, 1), util.MaxInt64(w.length, 1)-1-w.cursor)
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *Window) cursorHead(_ int64) {
	w.cursor -= w.cursor % w.width
}

func (w *Window) cursorEnd(count int64) {
	w.cursor = util.MinInt64((w.cursor/w.width+util.MaxInt64(count, 1))*w.width-1, util.MaxInt64(w.length, 1)-1)
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *Window) scrollUp() {
	if w.offset > 0 {
		w.offset -= w.width
	}
	if w.cursor >= w.offset+w.height*w.width {
		w.cursor -= w.width
	}
}

func (w *Window) scrollDown() {
	offset := util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = util.MinInt64(w.offset+w.width, offset)
	if w.cursor < w.offset {
		w.cursor += w.width
	}
}

func (w *Window) pageUp() {
	w.offset = util.MaxInt64(w.offset-(w.height-2)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *Window) pageDown() {
	offset := util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = util.MinInt64(w.offset+(w.height-2)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *Window) pageUpHalf() {
	w.offset = util.MaxInt64(w.offset-util.MaxInt64(w.height/2, 1)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *Window) pageDownHalf() {
	offset := util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = util.MinInt64(w.offset+util.MaxInt64(w.height/2, 1)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *Window) pageTop() {
	w.offset = 0
	w.cursor = 0
}

func (w *Window) pageEnd() {
	w.offset = util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
}
