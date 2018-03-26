package cmdline

import (
	"testing"

	"github.com/itchyny/bed/event"
)

func TestNewCmdline(t *testing.T) {
	c := NewCmdline()
	cmdline, cursor, _, _ := c.Get()
	if len(cmdline) != 0 {
		t.Errorf("cmdline should be empty but got %v", cmdline)
	}
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}
}

func TestCmdlineRun(t *testing.T) {
	c := NewCmdline()
	eventCh, cmdlineCh, redrawCh := make(chan event.Event), make(chan event.Event), make(chan struct{})
	c.Init(eventCh, cmdlineCh, redrawCh)
	go c.Run()
	events := []event.Event{
		event.Event{Type: event.StartCmdlineCommand}, event.Event{Type: event.Nop},
		event.Event{Type: event.Rune, Rune: 't'}, event.Event{Type: event.Rune, Rune: 'e'},
		event.Event{Type: event.CursorLeft}, event.Event{Type: event.CursorRight},
		event.Event{Type: event.CursorHead}, event.Event{Type: event.CursorEnd},
		event.Event{Type: event.BackspaceCmdline}, event.Event{Type: event.DeleteCmdline},
		event.Event{Type: event.DeleteWordCmdline}, event.Event{Type: event.ClearToHeadCmdline},
		event.Event{Type: event.ClearCmdline}, event.Event{Type: event.Rune, Rune: 't'},
		event.Event{Type: event.Rune, Rune: 'e'}, event.Event{Type: event.ExecuteCmdline},
		event.Event{Type: event.StartCmdlineCommand}, event.Event{Type: event.ExecuteCmdline},
	}
	go func() {
		for _, e := range events {
			cmdlineCh <- e
		}
	}()
	for i := 0; i < len(events)-4; i++ {
		<-redrawCh
	}
	e := <-eventCh
	if e.Type != event.Error {
		t.Errorf("cmdline should emit event.Error but got %v", e)
	}
	<-redrawCh
	cmdline, cursor, _, _ := c.Get()
	if string(cmdline) != "te" {
		t.Errorf("cmdline should be %q got %q", "te", string(cmdline))
	}
	if cursor != 2 {
		t.Errorf("cursor should be 2 but got %v", cursor)
	}
	<-redrawCh
}

func TestCmdlineCursorMotion(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.insert(ch)
	}
	cmdline, cursor, _, _ := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	_, cursor, _, _ = c.Get()
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}

	for i := 0; i < 10; i++ {
		c.cursorLeft()
	}
	_, cursor, _, _ = c.Get()
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}

	c.cursorRight()
	_, cursor, _, _ = c.Get()
	if cursor != 1 {
		t.Errorf("cursor should be 1 but got %v", cursor)
	}

	for i := 0; i < 10; i++ {
		c.cursorRight()
	}
	_, cursor, _, _ = c.Get()
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorHead()
	_, cursor, _, _ = c.Get()
	if cursor != 0 {
		t.Errorf("cursor should be 0 but got %v", cursor)
	}

	c.cursorEnd()
	_, cursor, _, _ = c.Get()
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}
}

func TestCmdlineCursorBackspaceDelete(t *testing.T) {
	c := NewCmdline()

	for _, ch := range "abcde" {
		c.insert(ch)
	}
	cmdline, cursor, _, _ := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.backspace()

	cmdline, cursor, _, _ = c.Get()
	if string(cmdline) != "abce" {
		t.Errorf("cmdline should be %v but got %v", "abce", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.deleteRune()

	cmdline, cursor, _, _ = c.Get()
	if string(cmdline) != "abc" {
		t.Errorf("cmdline should be %v but got %v", "abc", string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.deleteRune()

	cmdline, cursor, _, _ = c.Get()
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

	cmdline, cursor, _, _ = c.Get()
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

	cmdline, cursor, _, _ := c.Get()
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

	cmdline, cursor, _, _ = c.Get()
	if string(cmdline) != "x0z! de" {
		t.Errorf("cmdline should be %v but got %v", "x0z! de", string(cmdline))
	}
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}

	c.deleteWord()

	cmdline, cursor, _, _ = c.Get()
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
	cmdline, cursor, _, _ := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.clear()

	cmdline, cursor, _, _ = c.Get()
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
	cmdline, cursor, _, _ := c.Get()
	if string(cmdline) != "abcde" {
		t.Errorf("cmdline should be %v but got %v", "abcde", string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.cursorLeft()
	c.clearToHead()

	cmdline, cursor, _, _ = c.Get()
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

	cmdline, cursor, _, _ := c.Get()
	if string(cmdline) != "abxyde" {
		t.Errorf("cmdline should be %v but got %v", "abxyde", string(cmdline))
	}
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}
}

func TestCmdlineQuit(t *testing.T) {
	c := NewCmdline()
	ch := make(chan event.Event, 1)
	c.Init(ch, make(chan event.Event), make(chan struct{}))
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
		c.typ = ':'
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != event.Quit {
			t.Errorf("cmdline should emit quit event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteQuitAll(t *testing.T) {
	c := NewCmdline()
	ch := make(chan event.Event, 1)
	c.Init(ch, make(chan event.Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
	}{
		{"qall", "qa[ll]"},
		{"qa", "qa[ll]"},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.typ = ':'
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != event.QuitAll {
			t.Errorf("cmdline should emit quit all event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteWriteQuit(t *testing.T) {
	c := NewCmdline()
	ch := make(chan event.Event, 1)
	c.Init(ch, make(chan event.Event), make(chan struct{}))
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
		c.typ = ':'
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != event.WriteQuit {
			t.Errorf("cmdline should emit quit event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteGoto(t *testing.T) {
	c := NewCmdline()
	ch := make(chan event.Event, 1)
	c.Init(ch, make(chan event.Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
		typ  event.Type
	}{
		{"  :  :  $  ", "$", event.CursorGotoAbs},
		{"  :  123456789abcdef  ", "123456789abcdef", event.CursorGotoAbs},
		{"  fedcba  ", "fedcba", event.CursorGotoAbs},
		{"  +44ef ", "+44ef", event.CursorGotoRel},
		{"  -ff ", "-ff", event.CursorGotoRel},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.typ = ':'
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

func TestCmdlineComplete(t *testing.T) {
	c := NewCmdline()
	c.completor = newCompletor(&mockFilesystem{})
	eventCh, cmdlineCh, redrawCh := make(chan event.Event), make(chan event.Event), make(chan struct{})
	c.Init(eventCh, cmdlineCh, redrawCh)
	waitCh := make(chan struct{})
	go c.Run()
	go func() {
		cmdlineCh <- event.Event{Type: event.StartCmdlineCommand}
		cmdlineCh <- event.Event{Type: event.Rune, Rune: 'e'}
		cmdlineCh <- event.Event{Type: event.Rune, Rune: ' '}
		cmdlineCh <- event.Event{Type: event.Rune, Rune: '/'}
		cmdlineCh <- event.Event{Type: event.CompleteForwardCmdline}
		<-waitCh
		cmdlineCh <- event.Event{Type: event.CompleteForwardCmdline}
		<-waitCh
		cmdlineCh <- event.Event{Type: event.CompleteBackCmdline}
		<-waitCh
		cmdlineCh <- event.Event{Type: event.CursorEnd}
		cmdlineCh <- event.Event{Type: event.CompleteForwardCmdline}
		cmdlineCh <- event.Event{Type: event.CompleteForwardCmdline}
		<-waitCh
		cmdlineCh <- event.Event{Type: event.ExecuteCmdline}
	}()
	for i := 0; i < 5; i++ {
		<-redrawCh
	}
	cmdline, cursor, _, _ := c.Get()
	if string(cmdline) != "e /bin/" {
		t.Errorf("cmdline should be %q got %q", "e /bin/", string(cmdline))
	}
	if cursor != 7 {
		t.Errorf("cursor should be 7 but got %v", cursor)
	}
	waitCh <- struct{}{}
	<-redrawCh
	cmdline, cursor, _, _ = c.Get()
	if string(cmdline) != "e /tmp/" {
		t.Errorf("cmdline should be %q got %q", "e /tmp/", string(cmdline))
	}
	if cursor != 7 {
		t.Errorf("cursor should be 7 but got %v", cursor)
	}
	waitCh <- struct{}{}
	<-redrawCh
	cmdline, cursor, _, _ = c.Get()
	if string(cmdline) != "e /bin/" {
		t.Errorf("cmdline should be %q got %q", "e /bin/", string(cmdline))
	}
	if cursor != 7 {
		t.Errorf("cursor should be 7 but got %v", cursor)
	}
	waitCh <- struct{}{}
	<-redrawCh
	<-redrawCh
	<-redrawCh
	cmdline, cursor, _, _ = c.Get()
	if string(cmdline) != "e /bin/echo" {
		t.Errorf("cmdline should be %q got %q", "e /bin/echo", string(cmdline))
	}
	if cursor != 11 {
		t.Errorf("cursor should be 11 but got %v", cursor)
	}
	waitCh <- struct{}{}
	cmdline, cursor, _, _ = c.Get()
	if string(cmdline) != "e /bin/echo" {
		t.Errorf("cmdline should be %q got %q", "e /bin/echo", string(cmdline))
	}
	if cursor != 11 {
		t.Errorf("cursor should be 11 but got %v", cursor)
	}
	e := <-eventCh
	<-redrawCh
	if e.Type != event.Edit {
		t.Errorf("cmdline should emit event.Edit but got %v", e)
	}
	if e.Arg != "/bin/echo" {
		t.Errorf("cmdline should emit event with arg %q but got %q", "/bin/echo", e)
	}
}

func TestCmdlineSearch(t *testing.T) {
	c := NewCmdline()
	eventCh, cmdlineCh, redrawCh := make(chan event.Event), make(chan event.Event), make(chan struct{})
	waitCh := make(chan struct{})
	c.Init(eventCh, cmdlineCh, redrawCh)
	defer func() {
		close(eventCh)
		close(cmdlineCh)
		close(redrawCh)
	}()
	go c.Run()
	events1 := []event.Event{
		event.Event{Type: event.StartCmdlineSearchForward},
		event.Event{Type: event.Rune, Rune: 't'}, event.Event{Type: event.Rune, Rune: 't'},
		event.Event{Type: event.CursorLeft}, event.Event{Type: event.Rune, Rune: 'e'},
		event.Event{Type: event.Rune, Rune: 's'}, event.Event{Type: event.ExecuteCmdline},
	}
	events2 := []event.Event{
		event.Event{Type: event.StartCmdlineSearchBackward},
		event.Event{Type: event.Rune, Rune: 'x'}, event.Event{Type: event.Rune, Rune: 'y'},
		event.Event{Type: event.Rune, Rune: 'z'}, event.Event{Type: event.ExecuteCmdline},
	}
	go func() {
		for _, e := range events1 {
			cmdlineCh <- e
		}
		<-waitCh
		for _, e := range events2 {
			cmdlineCh <- e
		}
	}()
	for i := 0; i < len(events1)-1; i++ {
		<-redrawCh
	}
	e := <-eventCh
	<-redrawCh
	if e.Type != event.ExecuteSearch {
		t.Errorf("cmdline should emit event.ExecuteSearch but got %v", e)
	}
	if e.Arg != "test" {
		t.Errorf("cmdline should emit search event with Arg %q but got %q", "test", e.Arg)
	}
	if e.Rune != '/' {
		t.Errorf("cmdline should emit search event with Rune %q but got %q", '/', e.Rune)
	}
	waitCh <- struct{}{}
	for i := 0; i < len(events2)-1; i++ {
		<-redrawCh
	}
	e = <-eventCh
	<-redrawCh
	if e.Type != event.ExecuteSearch {
		t.Errorf("cmdline should emit event.ExecuteSearch but got %v", e)
	}
	if e.Arg != "xyz" {
		t.Errorf("cmdline should emit search event with Arg %q but got %q", "xyz", e.Arg)
	}
	if e.Rune != '?' {
		t.Errorf("cmdline should emit search event with Rune %q but got %q", '?', e.Rune)
	}
}
