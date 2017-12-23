package core

import (
	"reflect"
	"strings"
	"testing"
)

func TestWindowState(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	height, width := int64(10), int64(16)
	window, err := NewWindow(r, "test", height, width)
	if err != nil {
		t.Fatal(err)
	}

	state, err := window.State()
	if err != nil {
		t.Fatal(err)
	}

	if state.Name != "test" {
		t.Errorf("state.Name should be %q but got %q", "test", state.Name)
	}

	if state.Width != int(width) {
		t.Errorf("state.Width should be %d but got %d", int(width), state.Width)
	}

	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}

	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	if state.Size != 13 {
		t.Errorf("state.Size should be %d but got %d", 13, state.Size)
	}

	if state.Length != 13 {
		t.Errorf("state.Length should be %d but got %d", 13, state.Length)
	}

	if state.Mode != ModeNormal {
		t.Errorf("state.Mode should be %d but got %d", ModeNormal, state.Mode)
	}

	if state.Pending != false {
		t.Errorf("state.Pending should be %b but got %b", false, state.Pending)
	}

	if state.PendingByte != '\x00' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x00', state.PendingByte)
	}

	if !reflect.DeepEqual(state.EditedIndices, []int64{}) {
		t.Errorf("state.EditedIndices should be empty but got %v", state.EditedIndices)
	}

	expected := []byte("Hello, world!" + strings.Repeat("\x00", int(height*width)-13))
	if !reflect.DeepEqual(state.Bytes, expected) {
		t.Errorf("state.Bytes should be %q but got %q", expected, state.Bytes)
	}
}

func TestWindowEmptyState(t *testing.T) {
	r := strings.NewReader("")
	height, width := int64(10), int64(16)
	window, err := NewWindow(r, "test", height, width)
	if err != nil {
		t.Fatal(err)
	}

	state, err := window.State()
	if err != nil {
		t.Fatal(err)
	}

	if state.Name != "test" {
		t.Errorf("state.Name should be %q but got %q", "test", state.Name)
	}

	if state.Width != int(width) {
		t.Errorf("state.Width should be %d but got %d", int(width), state.Width)
	}

	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}

	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	if state.Size != 0 {
		t.Errorf("state.Size should be %d but got %d", 0, state.Size)
	}

	if state.Length != 0 {
		t.Errorf("state.Length should be %d but got %d", 0, state.Length)
	}

	if state.Mode != ModeNormal {
		t.Errorf("state.Mode should be %d but got %d", ModeNormal, state.Mode)
	}

	if state.Pending != false {
		t.Errorf("state.Pending should be %b but got %b", false, state.Pending)
	}

	if state.PendingByte != '\x00' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x00', state.PendingByte)
	}

	if !reflect.DeepEqual(state.EditedIndices, []int64{}) {
		t.Errorf("state.EditedIndices should be empty but got %v", state.EditedIndices)
	}

	expected := []byte(strings.Repeat("\x00", int(height*width)))
	if !reflect.DeepEqual(state.Bytes, expected) {
		t.Errorf("state.Bytes should be %q but got %q", expected, state.Bytes)
	}
}

func TestWindowCursorMotions(t *testing.T) {
	r := strings.NewReader(strings.Repeat("Hello, world!", 100))
	height, width := int64(10), int64(16)
	window, err := NewWindow(r, "test", height, width)
	if err != nil {
		t.Fatal(err)
	}

	state, _ := window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	window.cursorDown(0)
	state, _ = window.State()
	if state.Cursor != width {
		t.Errorf("state.Cursor should be %d but got %d", width, state.Cursor)
	}

	window.cursorDown(1)
	state, _ = window.State()
	if state.Cursor != width*2 {
		t.Errorf("state.Cursor should be %d but got %d", width*2, state.Cursor)
	}

	window.cursorUp(0)
	state, _ = window.State()
	if state.Cursor != width {
		t.Errorf("state.Cursor should be %d but got %d", width, state.Cursor)
	}

	window.cursorDown(10)
	state, _ = window.State()
	if state.Cursor != width*11 {
		t.Errorf("state.Cursor should be %d but got %d", width*11, state.Cursor)
	}
	if state.Offset != width*2 {
		t.Errorf("state.Offset should be %d but got %d", width*2, state.Offset)
	}
	if !strings.HasPrefix(string(state.Bytes), " world!") {
		t.Errorf("state.Bytes should start with %q but got %q", " world!", string(state.Bytes))
	}

	window.cursorRight(3)
	state, _ = window.State()
	if state.Cursor != width*11+3 {
		t.Errorf("state.Cursor should be %d but got %d", width*11+3, state.Cursor)
	}

	window.cursorRight(20)
	state, _ = window.State()
	if state.Cursor != width*12-1 {
		t.Errorf("state.Cursor should be %d but got %d", width*12-1, state.Cursor)
	}

	window.cursorLeft(3)
	state, _ = window.State()
	if state.Cursor != width*12-4 {
		t.Errorf("state.Cursor should be %d but got %d", width*12-4, state.Cursor)
	}

	window.cursorLeft(20)
	state, _ = window.State()
	if state.Cursor != width*11 {
		t.Errorf("state.Cursor should be %d but got %d", width*11, state.Cursor)
	}

	window.cursorPrev(154)
	state, _ = window.State()
	if state.Cursor != 22 {
		t.Errorf("state.Cursor should be %d but got %d", 22, state.Cursor)
	}
	if state.Offset != width {
		t.Errorf("state.Offset should be %d but got %d", width, state.Offset)
	}

	window.cursorNext(200)
	state, _ = window.State()
	if state.Cursor != 222 {
		t.Errorf("state.Cursor should be %d but got %d", 222, state.Cursor)
	}
	if state.Offset != width*4 {
		t.Errorf("state.Offset should be %d but got %d", width*4, state.Offset)
	}

	window.cursorNext(2000)
	state, _ = window.State()
	if state.Cursor != 1299 {
		t.Errorf("state.Cursor should be %d but got %d", 1299, state.Cursor)
	}
	if state.Offset != width*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorHead(1)
	state, _ = window.State()
	if state.Cursor != 1296 {
		t.Errorf("state.Cursor should be %d but got %d", 1296, state.Cursor)
	}
	if state.Offset != width*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorEnd(1)
	state, _ = window.State()
	if state.Cursor != 1299 {
		t.Errorf("state.Cursor should be %d but got %d", 1299, state.Cursor)
	}
	if state.Offset != width*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorUp(20)
	window.cursorEnd(1)
	state, _ = window.State()
	if state.Cursor != 991 {
		t.Errorf("state.Cursor should be %d but got %d", 991, state.Cursor)
	}
	if state.Offset != width*61 {
		t.Errorf("state.Offset should be %d but got %d", width*61, state.Offset)
	}

	window.cursorEnd(11)
	state, _ = window.State()
	if state.Cursor != 1151 {
		t.Errorf("state.Cursor should be %d but got %d", 1151, state.Cursor)
	}
	if state.Offset != width*62 {
		t.Errorf("state.Offset should be %d but got %d", width*62, state.Offset)
	}

	window.cursorDown(30)
	state, _ = window.State()
	if state.Cursor != 1299 {
		t.Errorf("state.Cursor should be %d but got %d", 1299, state.Cursor)
	}
	if state.Offset != width*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorPrev(2000)
	state, _ = window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}
	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}

	window.cursorDown(2000)
	state, _ = window.State()
	if state.Cursor != width*81 {
		t.Errorf("state.Cursor should be %d but got %d", width*81, state.Cursor)
	}
	if state.Offset != width*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorRight(1000)
	state, _ = window.State()
	if state.Cursor != 1299 {
		t.Errorf("state.Cursor should be %d but got %d", 1299, state.Cursor)
	}
	if state.Offset != width*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorUp(2000)
	state, _ = window.State()
	if state.Cursor != 3 {
		t.Errorf("state.Cursor should be %d but got %d", 3, state.Cursor)
	}
	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}
}
