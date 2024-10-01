package cmdline

import (
	"reflect"
	"runtime"
	"strings"
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
		{Type: event.StartCmdlineCommand},
		{Type: event.Nop},
		{Type: event.Rune, Rune: 't'},
		{Type: event.Rune, Rune: 'e'},
		{Type: event.CursorLeft},
		{Type: event.CursorRight},
		{Type: event.CursorHead},
		{Type: event.CursorEnd},
		{Type: event.BackspaceCmdline},
		{Type: event.DeleteCmdline},
		{Type: event.DeleteWordCmdline},
		{Type: event.ClearToHeadCmdline},
		{Type: event.ClearCmdline},
		{Type: event.Rune, Rune: 't'},
		{Type: event.Rune, Rune: 'e'},
		{Type: event.ExecuteCmdline},
		{Type: event.StartCmdlineCommand},
		{Type: event.ExecuteCmdline},
	}
	go func() {
		for _, e := range events {
			cmdlineCh <- e
		}
	}()
	for range len(events) - 4 {
		<-redrawCh
	}
	e := <-eventCh
	if e.Type != event.Error {
		t.Errorf("cmdline should emit Error event but got %v", e)
	}
	<-redrawCh
	cmdline, cursor, _, _ := c.Get()
	if expected := "te"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q got %q", expected, string(cmdline))
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
	if expected := "abcde"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	_, cursor, _, _ = c.Get()
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}

	for range 10 {
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

	for range 10 {
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
	if expected := "abcde"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.backspace()

	cmdline, cursor, _, _ = c.Get()
	if expected := "abce"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.deleteRune()

	cmdline, cursor, _, _ = c.Get()
	if expected := "abc"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.deleteRune()

	cmdline, cursor, _, _ = c.Get()
	if expected := "abc"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
	}
	if cursor != 3 {
		t.Errorf("cursor should be 3 but got %v", cursor)
	}

	c.cursorLeft()
	c.cursorLeft()
	c.backspace()
	c.backspace()

	cmdline, cursor, _, _ = c.Get()
	if expected := "bc"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
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
	if expected := "de"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
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
	if expected := "x0z! de"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
	}
	if cursor != 4 {
		t.Errorf("cursor should be 4 but got %v", cursor)
	}

	c.deleteWord()

	cmdline, cursor, _, _ = c.Get()
	if expected := "x0z de"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
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
	if expected := "abcde"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.clear()

	cmdline, cursor, _, _ = c.Get()
	if expected := ""; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
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
	if expected := "abcde"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
	}
	if cursor != 5 {
		t.Errorf("cursor should be 5 but got %v", cursor)
	}

	c.cursorLeft()
	c.cursorLeft()
	c.clearToHead()

	cmdline, cursor, _, _ = c.Get()
	if expected := "de"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
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
	if expected := "abxyde"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q but got %q", expected, string(cmdline))
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
		if e.Bang {
			t.Errorf("cmdline should emit quit event without bang")
		}
	}
}

func TestCmdlineForceQuit(t *testing.T) {
	c := NewCmdline()
	ch := make(chan event.Event, 1)
	c.Init(ch, make(chan event.Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
	}{
		{"exit!", "exi[t]"},
		{"q!", "q[uit]"},
		{"quit!", "q[uit]"},
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
		if !e.Bang {
			t.Errorf("cmdline should emit quit event with bang")
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
			t.Errorf("cmdline should emit QuitAll event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteQuitErr(t *testing.T) {
	c := NewCmdline()
	ch := make(chan event.Event, 1)
	c.Init(ch, make(chan event.Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
	}{
		{"cquit", "cq[uit]"},
		{"cq", "cq[uit]"},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.typ = ':'
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != event.QuitErr {
			t.Errorf("cmdline should emit QuitErr event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteWrite(t *testing.T) {
	c := NewCmdline()
	ch := make(chan event.Event, 1)
	c.Init(ch, make(chan event.Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd  string
		name string
	}{
		{"w", "w[rite]"},
		{" :   : write  sample.txt", "w[rite]"},
		{"'<,'>write sample.txt", "w[rite]"},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.typ = ':'
		c.execute()
		e := <-ch
		if e.CmdName != cmd.name {
			t.Errorf("cmdline should report command name %q but got %q", cmd.name, e.CmdName)
		}
		if e.Type != event.Write {
			t.Errorf("cmdline should emit Write event with %q", cmd.cmd)
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
			t.Errorf("cmdline should emit WriteQuit event with %q", cmd.cmd)
		}
	}
}

func TestCmdlineExecuteGoto(t *testing.T) {
	c := NewCmdline()
	ch := make(chan event.Event, 1)
	c.Init(ch, make(chan event.Event), make(chan struct{}))
	for _, cmd := range []struct {
		cmd string
		pos event.Position
		typ event.Type
	}{
		{"  :  :  $  ", event.End{}, event.CursorGoto},
		{"  :123456789  ", event.Absolute{Offset: 123456789}, event.CursorGoto},
		{"  +16777216 ", event.Relative{Offset: 16777216}, event.CursorGoto},
		{"  -256 ", event.Relative{Offset: -256}, event.CursorGoto},
		{"  :  0x123456789abcdef  ", event.Absolute{Offset: 0x123456789abcdef}, event.CursorGoto},
		{"  0xfedcba  ", event.Absolute{Offset: 0xfedcba}, event.CursorGoto},
		{"  +0x44ef ", event.Relative{Offset: 0x44ef}, event.CursorGoto},
		{"  -0xff ", event.Relative{Offset: -0xff}, event.CursorGoto},
		{"10go", event.Absolute{Offset: 10}, event.CursorGoto},
		{"+10 got", event.Relative{Offset: 10}, event.CursorGoto},
		{"$-10 goto", event.End{Offset: -10}, event.CursorGoto},
		{"10%", event.Absolute{Offset: 10}, event.CursorGoto},
		{"+10%", event.Relative{Offset: 10}, event.CursorGoto},
		{"$-10%", event.End{Offset: -10}, event.CursorGoto},
	} {
		c.clear()
		c.cmdline = []rune(cmd.cmd)
		c.typ = ':'
		c.execute()
		e := <-ch
		expected := "goto"
		if strings.HasSuffix(cmd.cmd, "%") {
			expected = "%"
		} else if strings.Contains(cmd.cmd, "go") {
			expected = "go[to]"
		}
		if e.CmdName != expected {
			t.Errorf("cmdline should report command name %q but got %q", expected, e.CmdName)
		}
		if !reflect.DeepEqual(e.Range.From, cmd.pos) {
			t.Errorf("cmdline should report command with position %#v but got %#v", cmd.pos, e.Range.From)
		}
		if e.Type != cmd.typ {
			t.Errorf("cmdline should emit %d but got %d with %q", cmd.typ, e.Type, cmd.cmd)
		}
	}
}

func TestCmdlineComplete(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
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
	for range 5 {
		<-redrawCh
	}
	cmdline, cursor, _, _ := c.Get()
	if expected := "e /bin/"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q got %q", expected, string(cmdline))
	}
	if cursor != 7 {
		t.Errorf("cursor should be 7 but got %v", cursor)
	}
	waitCh <- struct{}{}
	<-redrawCh
	cmdline, cursor, _, _ = c.Get()
	if expected := "e /tmp/"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q got %q", expected, string(cmdline))
	}
	if cursor != 7 {
		t.Errorf("cursor should be 7 but got %v", cursor)
	}
	waitCh <- struct{}{}
	<-redrawCh
	cmdline, cursor, _, _ = c.Get()
	if expected := "e /bin/"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q got %q", expected, string(cmdline))
	}
	if cursor != 7 {
		t.Errorf("cursor should be 7 but got %v", cursor)
	}
	waitCh <- struct{}{}
	<-redrawCh
	<-redrawCh
	<-redrawCh
	cmdline, cursor, _, _ = c.Get()
	if expected := "e /bin/echo"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q got %q", expected, string(cmdline))
	}
	if cursor != 11 {
		t.Errorf("cursor should be 11 but got %v", cursor)
	}
	waitCh <- struct{}{}
	go func() { <-redrawCh }()
	e := <-eventCh
	cmdline, cursor, _, _ = c.Get()
	if expected := "e /bin/echo"; string(cmdline) != expected {
		t.Errorf("cmdline should be %q got %q", expected, string(cmdline))
	}
	if cursor != 11 {
		t.Errorf("cursor should be 11 but got %v", cursor)
	}
	if e.Type != event.Edit {
		t.Errorf("cmdline should emit Edit event but got %v", e)
	}
	if expected := "/bin/echo"; e.Arg != expected {
		t.Errorf("cmdline should emit event with arg %q but got %v", expected, e)
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
		{Type: event.StartCmdlineSearchForward},
		{Type: event.Rune, Rune: 't'},
		{Type: event.Rune, Rune: 't'},
		{Type: event.CursorLeft},
		{Type: event.Rune, Rune: 'e'},
		{Type: event.Rune, Rune: 's'},
		{Type: event.ExecuteCmdline},
	}
	events2 := []event.Event{
		{Type: event.StartCmdlineSearchBackward},
		{Type: event.Rune, Rune: 'x'},
		{Type: event.Rune, Rune: 'y'},
		{Type: event.Rune, Rune: 'z'},
		{Type: event.ExecuteCmdline},
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
	for range len(events1) - 1 {
		<-redrawCh
	}
	e := <-eventCh
	<-redrawCh
	if e.Type != event.ExecuteSearch {
		t.Errorf("cmdline should emit ExecuteSearch event but got %v", e)
	}
	if expected := "test"; e.Arg != expected {
		t.Errorf("cmdline should emit search event with Arg %q but got %q", expected, e.Arg)
	}
	if e.Rune != '/' {
		t.Errorf("cmdline should emit search event with Rune %q but got %q", '/', e.Rune)
	}
	waitCh <- struct{}{}
	for range len(events2) - 1 {
		<-redrawCh
	}
	e = <-eventCh
	<-redrawCh
	if e.Type != event.ExecuteSearch {
		t.Errorf("cmdline should emit ExecuteSearch event but got %v", e)
	}
	if expected := "xyz"; e.Arg != expected {
		t.Errorf("cmdline should emit search event with Arg %q but got %q", expected, e.Arg)
	}
	if e.Rune != '?' {
		t.Errorf("cmdline should emit search event with Rune %q but got %q", '?', e.Rune)
	}
}

func TestCmdlineHistory(t *testing.T) {
	c := NewCmdline()
	eventCh, cmdlineCh, redrawCh := make(chan event.Event), make(chan event.Event), make(chan struct{})
	c.Init(eventCh, cmdlineCh, redrawCh)
	go c.Run()
	events0 := []event.Event{
		{Type: event.StartCmdlineCommand},
		{Type: event.Rune, Rune: 'n'},
		{Type: event.Rune, Rune: 'e'},
		{Type: event.Rune, Rune: 'w'},
		{Type: event.ExecuteCmdline},
	}
	events1 := []event.Event{
		{Type: event.StartCmdlineCommand},
		{Type: event.Rune, Rune: 'v'},
		{Type: event.Rune, Rune: 'n'},
		{Type: event.Rune, Rune: 'e'},
		{Type: event.Rune, Rune: 'w'},
		{Type: event.ExecuteCmdline},
	}
	events2 := []event.Event{
		{Type: event.StartCmdlineCommand},
		{Type: event.CursorUp},
		{Type: event.ExecuteCmdline},
	}
	events3 := []event.Event{
		{Type: event.StartCmdlineCommand},
		{Type: event.CursorUp},
		{Type: event.CursorUp},
		{Type: event.CursorUp},
		{Type: event.CursorDown},
		{Type: event.ExecuteCmdline},
	}
	events4 := []event.Event{
		{Type: event.StartCmdlineCommand},
		{Type: event.CursorUp},
		{Type: event.ExecuteCmdline},
	}
	events5 := []event.Event{
		{Type: event.StartCmdlineSearchForward},
		{Type: event.Rune, Rune: 't'},
		{Type: event.Rune, Rune: 'e'},
		{Type: event.Rune, Rune: 's'},
		{Type: event.Rune, Rune: 't'},
		{Type: event.ExecuteCmdline},
	}
	events6 := []event.Event{
		{Type: event.StartCmdlineSearchForward},
		{Type: event.CursorUp},
		{Type: event.CursorDown},
		{Type: event.Rune, Rune: 'n'},
		{Type: event.Rune, Rune: 'e'},
		{Type: event.Rune, Rune: 'w'},
		{Type: event.ExecuteCmdline},
	}
	events7 := []event.Event{
		{Type: event.StartCmdlineSearchBackward},
		{Type: event.CursorUp},
		{Type: event.CursorUp},
		{Type: event.ExecuteCmdline},
	}
	events8 := []event.Event{
		{Type: event.StartCmdlineCommand},
		{Type: event.CursorUp},
		{Type: event.CursorUp},
		{Type: event.ExecuteCmdline},
	}
	events9 := []event.Event{
		{Type: event.StartCmdlineSearchForward},
		{Type: event.CursorUp},
		{Type: event.ExecuteCmdline},
	}
	go func() {
		for _, events := range [][]event.Event{
			events0, events1, events2, events3, events4,
			events5, events6, events7, events8, events9,
		} {
			for _, e := range events {
				cmdlineCh <- e
			}
		}
	}()
	for range len(events0) - 1 {
		<-redrawCh
	}
	e := <-eventCh
	if e.Type != event.New {
		t.Errorf("cmdline should emit New event but got %v", e)
	}
	for range len(events1) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.Vnew {
		t.Errorf("cmdline should emit Vnew event but got %v", e)
	}
	for range len(events2) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.Vnew {
		t.Errorf("cmdline should emit Vnew event but got %v", e)
	}
	for range len(events3) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.New {
		t.Errorf("cmdline should emit New event but got %v", e.Type)
	}
	for range len(events4) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.New {
		t.Errorf("cmdline should emit New event but got %v", e.Type)
	}
	for range len(events5) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.ExecuteSearch {
		t.Errorf("cmdline should emit ExecuteSearch event but got %v", e)
	}
	if expected := "test"; e.Arg != expected {
		t.Errorf("cmdline should emit search event with Arg %q but got %q", expected, e.Arg)
	}
	for range len(events6) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.ExecuteSearch {
		t.Errorf("cmdline should emit ExecuteSearch event but got %v", e)
	}
	if expected := "new"; e.Arg != expected {
		t.Errorf("cmdline should emit search event with Arg %q but got %q", expected, e.Arg)
	}
	for range len(events7) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.ExecuteSearch {
		t.Errorf("cmdline should emit ExecuteSearch event but got %v", e)
	}
	if expected := "test"; e.Arg != expected {
		t.Errorf("cmdline should emit search event with Arg %q but got %q", expected, e.Arg)
	}
	for range len(events8) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.Vnew {
		t.Errorf("cmdline should emit Vnew event but got %v", e.Type)
	}
	for range len(events9) {
		<-redrawCh
	}
	e = <-eventCh
	if e.Type != event.ExecuteSearch {
		t.Errorf("cmdline should emit ExecuteSearch event but got %v", e)
	}
	if expected := "test"; e.Arg != expected {
		t.Errorf("cmdline should emit search event with Arg %q but got %q", expected, e.Arg)
	}
	<-redrawCh
}
