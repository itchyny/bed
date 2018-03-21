package history

import (
	"strings"
	"testing"

	"github.com/itchyny/bed/buffer"
)

func TestHistoryUndo(t *testing.T) {
	history := NewHistory()
	b, index, offset, cursor := history.Undo()
	if b != nil {
		t.Errorf("history.Undo should return nil buffer but got %q", b)
	}
	if index != -1 {
		t.Errorf("history.Undo should return index -1 but got %d", index)
	}
	if offset != 0 {
		t.Errorf("history.Undo should return offset 0 but got %d", offset)
	}
	if cursor != 0 {
		t.Errorf("history.Undo should return cursor 0 but got %d", cursor)
	}

	buffer1 := buffer.NewBuffer(strings.NewReader("test1"))
	history.Push(buffer1, 2, 1)

	buffer2 := buffer.NewBuffer(strings.NewReader("test2"))
	history.Push(buffer2, 3, 2)

	buf := make([]byte, 8)
	b, index, offset, cursor = history.Undo()
	b.Read(buf)
	if string(buf) != "test1\x00\x00\x00" {
		t.Errorf("buf should be %q but got %q", "test1\x00\x00\x00", string(buf))
	}
	if index != 0 {
		t.Errorf("push should return index 0 but got %d", index)
	}
	if offset != 2 {
		t.Errorf("push should return offset 2 but got %d", offset)
	}
	if cursor != 1 {
		t.Errorf("push should return cursor 1 but got %d", cursor)
	}

	buf = make([]byte, 8)
	b, offset, cursor = history.Redo()
	b.Read(buf)
	if string(buf) != "test2\x00\x00\x00" {
		t.Errorf("buf should be %q but got %q", "test2\x00\x00\x00", string(buf))
	}
	if offset != 3 {
		t.Errorf("history.Redo should return offset 3 but got %d", offset)
	}
	if cursor != 2 {
		t.Errorf("history.Redo should return cursor 2 but got %d", cursor)
	}

	history.Undo()
	buffer3 := buffer.NewBuffer(strings.NewReader("test2"))
	history.Push(buffer3, 3, 2)

	b, offset, cursor = history.Redo()
	if b != nil {
		t.Errorf("history.Redo should return nil buffer but got %q", b)
	}
	if offset != 0 {
		t.Errorf("history.Redo should return offset 0 but got %d", offset)
	}
	if cursor != 0 {
		t.Errorf("history.Redo should return cursor 0 but got %d", cursor)
	}
}
