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
