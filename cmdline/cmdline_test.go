package cmdline

import (
	"testing"
)

func TestNewCmdline(t *testing.T) {
	c := NewCmdline()
	cmdline, cursor := c.Get()
	if len(cmdline) != 0 {
		t.Errorf("cmdline should be empty but got %v", cmdline)
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}
}

func TestCursorMotion(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.Insert(ch)
	}
	cmdline, cursor := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.CursorLeft()
	_, cursor = c.Get()
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}

	for i := 0; i < 10; i++ {
		c.CursorLeft()
	}
	_, cursor = c.Get()
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}

	c.CursorRight()
	_, cursor = c.Get()
	if cursor != 1 {
		t.Errorf("cursor should be 1 but got %v", cursor)
	}

	for i := 0; i < 10; i++ {
		c.CursorRight()
	}
	_, cursor = c.Get()
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.CursorHead()
	_, cursor = c.Get()
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}

	c.CursorEnd()
	_, cursor = c.Get()
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}
}

func TestCursorBackspaceDelete(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.Insert(ch)
	}
	cmdline, cursor := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.CursorLeft()
	c.Backspace()

	cmdline, cursor = c.Get()
	if string(cmdline) != "abce" {
		t.Errorf("cmdline should be %v but got %v", "abce", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.Delete()

	cmdline, cursor = c.Get()
	if string(cmdline) != "abc" {
		t.Errorf("cmdline should be %v but got %v", "abc", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.Delete()

	cmdline, cursor = c.Get()
	if string(cmdline) != "abc" {
		t.Errorf("cmdline should be %v but got %v", "abc", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.CursorLeft()
	c.CursorLeft()
	c.Backspace()
	c.Backspace()

	cmdline, cursor = c.Get()
	if string(cmdline) != "bc" {
		t.Errorf("cmdline should be %v but got %v", "bc", string(cmdline))
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}
}

func TestCursorClear(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.Insert(ch)
	}
	cmdline, cursor := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.CursorLeft()
	c.Clear()

	cmdline, cursor = c.Get()
	if string(cmdline) != "" {
		t.Errorf("cmdline should be %v but got %v", "", string(cmdline))
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}
}

func TestCursorClearToHead(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.Insert(ch)
	}
	cmdline, cursor := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.CursorLeft()
	c.CursorLeft()
	c.ClearToHead()

	cmdline, cursor = c.Get()
	if string(cmdline) != "de" {
		t.Errorf("cmdline should be %v but got %v", "de", string(cmdline))
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}
}

func TestCursorInsert(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.Insert(ch)
	}

	c.CursorLeft()
	c.CursorLeft()
	c.Backspace()
	c.Insert('x')
	c.Insert('y')

	cmdline, cursor := c.Get()
	if string(cmdline) != "abxyde" {
		t.Errorf("cmdline should be %v but got %v", "abxyde", string(cmdline))
	}
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}
}
