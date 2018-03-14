package editor

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/itchyny/bed/cmdline"
	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/window"
)

type testUI struct {
	eventCh chan<- Event
	quitCh  <-chan struct{}
}

func (ui *testUI) Init(eventCh chan<- Event, quitCh <-chan struct{}) error {
	ui.eventCh, ui.quitCh = eventCh, quitCh
	return nil
}

func (ui *testUI) Run(km map[Mode]*KeyManager) { <-ui.quitCh }

func (ui *testUI) Height() int { return 0 }

func (ui *testUI) Redraw(state State) error { return nil }

func (ui *testUI) Close() error { return nil }

func (ui *testUI) Emit(e Event) { ui.eventCh <- e }

type testCmdline struct{}

func (c *testCmdline) Init(eventCh chan<- Event, cmdlineCh <-chan Event, redrawCh chan<- struct{}) error {
	return nil
}

func (c *testCmdline) Run() {}

func (c *testCmdline) Get() ([]rune, int) { return nil, 0 }

func (c *testCmdline) Execute() {}

func TestEditorOpenEmptyWriteQuit(t *testing.T) {
	ui := &testUI{}
	editor := NewEditor(ui, window.NewManager(), &testCmdline{})
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := ioutil.TempFile("", "bed-test-editor-open-empty-write-quit")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f.Close()
	if err := editor.OpenEmpty(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, t := range []EventType{EventIncrement, EventIncrement, EventDecrement} {
			ui.Emit(Event{Type: t})
		}
		ui.Emit(Event{Type: EventWrite, Args: []string{f.Name()}})
		time.Sleep(100 * time.Millisecond)
		ui.Emit(Event{Type: EventQuit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if editor.errtyp != MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if string(bs) != "\x01" {
		t.Errorf("file contents should be %q but got %q", "\x12\x48\xff", string(bs))
	}
}

func TestEditorOpenWriteQuit(t *testing.T) {
	ui := &testUI{}
	editor := NewEditor(ui, window.NewManager(), &testCmdline{})
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := ioutil.TempFile("", "bed-test-editor-open-write-quit")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f.Close()
	if err := editor.Open(f.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	defer os.Remove(f.Name())
	go func() {
		for _, e := range []struct {
			typ EventType
			ch  rune
		}{
			{EventStartInsert, '-'}, {EventRune, '4'}, {EventRune, '8'}, {EventRune, '0'}, {EventRune, '0'},
			{EventRune, 'f'}, {EventRune, 'a'}, {EventExitInsert, '-'}, {EventCursorLeft, '-'}, {EventDecrement, '-'},
			{EventStartInsertHead, '-'}, {EventRune, '1'}, {EventRune, '2'}, {EventExitInsert, '-'},
			{EventCursorEnd, '-'}, {EventDelete, '-'}, {EventWriteQuit, '-'},
		} {
			ui.Emit(Event{Type: e.typ, Rune: e.ch})
		}
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.err; err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if string(bs) != "\x12\x48\xff" {
		t.Errorf("file contents should be %q but got %q", "\x12\x48\xff", string(bs))
	}
}

func TestEditorCmdlineQuit(t *testing.T) {
	ui := &testUI{}
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, e := range []struct {
			typ EventType
			ch  rune
		}{
			{EventStartCmdline, '-'}, {EventRune, 'q'}, {EventRune, 'u'}, {EventRune, 'i'},
			{EventRune, 't'}, {EventExecuteCmdline, '-'},
		} {
			ui.Emit(Event{Type: e.typ, Rune: e.ch})
		}
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.err; err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}
