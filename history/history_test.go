package history

import (
	"strings"
	"testing"

	"github.com/itchyny/bed/buffer"
)

func TestHistoryUndo(t *testing.T) {
	history := NewHistory()
	b, index, offset, cursor, tick := history.Undo()
	if b != nil {
		t.Errorf("history.Undo should return nil buffer but got %v", b)
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
	if tick != 0 {
		t.Errorf("history.Undo should return tick 0 but got %d", tick)
	}

	buffer1 := buffer.NewBuffer(strings.NewReader("test1"))
	history.Push(buffer1, 2, 1, 1)

	buffer2 := buffer.NewBuffer(strings.NewReader("test2"))
	history.Push(buffer2, 3, 2, 2)

	buf := make([]byte, 8)
	b, index, offset, cursor, tick = history.Undo()
	if b == nil {
		t.Fatalf("history.Undo should return buffer but got nil")
	}
	_, err := b.Read(buf)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "test1\x00\x00\x00"; string(buf) != expected {
		t.Errorf("buf should be %q but got %q", expected, string(buf))
	}
	if index != 0 {
		t.Errorf("history.Undo should return index 0 but got %d", index)
	}
	if offset != 2 {
		t.Errorf("history.Undo should return offset 2 but got %d", offset)
	}
	if cursor != 1 {
		t.Errorf("history.Undo should return cursor 1 but got %d", cursor)
	}
	if tick != 1 {
		t.Errorf("history.Undo should return tick 1 but got %d", tick)
	}

	buf = make([]byte, 8)
	b, offset, cursor, tick = history.Redo()
	if b == nil {
		t.Fatalf("history.Redo should return buffer but got nil")
	}
	_, err = b.Read(buf)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "test2\x00\x00\x00"; string(buf) != expected {
		t.Errorf("buf should be %q but got %q", expected, string(buf))
	}
	if offset != 3 {
		t.Errorf("history.Redo should return offset 3 but got %d", offset)
	}
	if cursor != 2 {
		t.Errorf("history.Redo should return cursor 2 but got %d", cursor)
	}
	if tick != 2 {
		t.Errorf("history.Redo should return cursor 2 but got %d", tick)
	}

	history.Undo()
	buffer3 := buffer.NewBuffer(strings.NewReader("test2"))
	history.Push(buffer3, 3, 2, 3)

	b, offset, cursor, tick = history.Redo()
	if b != nil {
		t.Errorf("history.Redo should return nil buffer but got %v", b)
	}
	if offset != 0 {
		t.Errorf("history.Redo should return offset 0 but got %d", offset)
	}
	if cursor != 0 {
		t.Errorf("history.Redo should return cursor 0 but got %d", cursor)
	}
	if tick != 0 {
		t.Errorf("history.Redo should return tick 0 but got %d", tick)
	}
}
