package cmdline

import (
	"testing"

	. "github.com/itchyny/bed/common"
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

func TestCmdlineCursorMotion(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.insert(ch)
	}
	cmdline, cursor := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	_, cursor = c.Get()
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}

	for i := 0; i < 10; i++ {
		c.cursorLeft()
	}
	_, cursor = c.Get()
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}

	c.cursorRight()
	_, cursor = c.Get()
	if cursor != 1 {
		t.Errorf("cursor should be 1 but got %v", cursor)
	}

	for i := 0; i < 10; i++ {
		c.cursorRight()
	}
	_, cursor = c.Get()
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorHead()
	_, cursor = c.Get()
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}

	c.cursorEnd()
	_, cursor = c.Get()
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}
}

func TestCmdlineCursorBackspaceDelete(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.insert(ch)
	}
	cmdline, cursor := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.backspace()

	cmdline, cursor = c.Get()
	if string(cmdline) != "abce" {
		t.Errorf("cmdline should be %v but got %v", "abce", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.deleteRune()

	cmdline, cursor = c.Get()
	if string(cmdline) != "abc" {
		t.Errorf("cmdline should be %v but got %v", "abc", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.deleteRune()

	cmdline, cursor = c.Get()
	if string(cmdline) != "abc" {
		t.Errorf("cmdline should be %v but got %v", "abc", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.cursorLeft()
	c.cursorLeft()
	c.backspace()
	c.backspace()

	cmdline, cursor = c.Get()
	if string(cmdline) != "bc" {
		t.Errorf("cmdline should be %v but got %v", "bc", string(cmdline))
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}
}

func TestCmdlineCursorDeleteWord(t *testing.T) {
	c := NewCmdline()
	for _, ch := range "abcde" {
		c.insert(ch)
	}

	c.cursorLeft()
	c.cursorLeft()
	c.deleteWord()

	cmdline, cursor := c.Get()
	if string(cmdline) != "de" {
		t.Errorf("cmdline should be %v but got %v", "de", string(cmdline))
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}

	for _, ch := range "x0z!123  " {
		c.insert(ch)
	}
	c.cursorLeft()
	c.deleteWord()

	cmdline, cursor = c.Get()
	if string(cmdline) != "x0z! de" {
		t.Errorf("cmdline should be %v but got %v", "x0z! de", string(cmdline))
	}
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}

	c.deleteWord()

	cmdline, cursor = c.Get()
	if string(cmdline) != "x0z de" {
		t.Errorf("cmdline should be %v but got %v", "x0z de", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}
}

func TestCmdlineCursorClear(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.insert(ch)
	}
	cmdline, cursor := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.clear()

	cmdline, cursor = c.Get()
	if string(cmdline) != "" {
		t.Errorf("cmdline should be %v but got %v", "", string(cmdline))
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}
}

func TestCmdlineCursorClearToHead(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.insert(ch)
	}
	cmdline, cursor := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.cursorLeft()
	c.clearToHead()

	cmdline, cursor = c.Get()
	if string(cmdline) != "de" {
		t.Errorf("cmdline should be %v but got %v", "de", string(cmdline))
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}
}

func TestCmdlineCursorInsert(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.insert(ch)
	}

	c.cursorLeft()
	c.cursorLeft()
	c.backspace()
	c.insert('x')
	c.insert('y')

	cmdline, cursor := c.Get()
	if string(cmdline) != "abxyde" {
		t.Errorf("cmdline should be %v but got %v", "abxyde", string(cmdline))
	}
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}
}

func TestCmdlineQuit(t *testing.T) {
	c := NewCmdline()
	ch := make(chan Event, 1)
	c.Init(ch, make(chan Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
	}{
		{"exi", "exi[t]"},
		{"quit", "q[uit]"},
		{"q", "q[uit]"},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != EventQuit {
			t.Errorf("cmdline should emit quit event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteQuitAll(t *testing.T) {
	c := NewCmdline()
	ch := make(chan Event, 1)
	c.Init(ch, make(chan Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
	}{
		{"qall", "qa[ll]"},
		{"qa", "qa[ll]"},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != EventQuitAll {
			t.Errorf("cmdline should emit quit all event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteWriteQuit(t *testing.T) {
	c := NewCmdline()
	ch := make(chan Event, 1)
	c.Init(ch, make(chan Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
	}{
		{"wq", "wq"},
		{"x", "x[it]"},
		{"xit", "x[it]"},
		{"xa", "xa[ll]"},
		{"xall", "xa[ll]"},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != EventWriteQuit {
			t.Errorf("cmdline should emit quit event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteGoto(t *testing.T) {
	c := NewCmdline()
	ch := make(chan Event, 1)
	c.Init(ch, make(chan Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
		typ  EventType
	}{
		{"  :  :  $  ", "$", EventCursorGotoAbs},
		{"  :  123456789abcdef  ", "123456789abcdef", EventCursorGotoAbs},
		{"  fedcba  ", "fedcba", EventCursorGotoAbs},
		{"  +44ef ", "+44ef", EventCursorGotoRel},
		{"  -ff ", "-ff", EventCursorGotoRel},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != cmd.typ {
			t.Errorf("cmdline should emit %q but got %q with %q", cmd.typ, e.Type, cmd.cmd)
		}
	}
}
