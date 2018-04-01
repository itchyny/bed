package editor

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/itchyny/bed/cmdline"
	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/key"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/state"
	"github.com/itchyny/bed/window"
)

type testUI struct {
	eventCh chan<- event.Event
	mu      *sync.Mutex
}

func newTestUI() *testUI {
	return &testUI{mu: new(sync.Mutex)}
}

func (ui *testUI) Init(eventCh chan<- event.Event) error {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.eventCh = eventCh
	return nil
}

func (ui *testUI) Run(km map[mode.Mode]*key.Manager) {}

func (ui *testUI) Height() int { return 10 }

func (ui *testUI) Size() (int, int) { return 10, 10 }

func (ui *testUI) Redraw(_ state.State) error { return nil }

func (ui *testUI) Close() error { return nil }

func (ui *testUI) Emit(e event.Event) {
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.eventCh <- e
}

func TestEditorOpenEmptyWriteQuit(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := ioutil.TempFile("", "bed-test-editor-open-empty-write-quit")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f.Close()
	defer os.Remove(f.Name())
	if err := editor.OpenEmpty(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, t := range []event.Type{event.Increment, event.Increment, event.Decrement} {
			ui.Emit(event.Event{Type: t})
		}
		time.Sleep(100 * time.Millisecond)
		ui.Emit(event.Event{Type: event.Write, Arg: f.Name()})
		time.Sleep(100 * time.Millisecond)
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if editor.errtyp != state.MessageInfo {
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
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
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
			typ event.Type
			ch  rune
		}{
			{event.StartInsert, '-'}, {event.Rune, '4'}, {event.Rune, '8'}, {event.Rune, '0'}, {event.Rune, '0'},
			{event.Rune, 'f'}, {event.Rune, 'a'}, {event.ExitInsert, '-'}, {event.CursorLeft, '-'}, {event.Decrement, '-'},
			{event.StartInsertHead, '-'}, {event.Rune, '1'}, {event.Rune, '2'}, {event.ExitInsert, '-'},
			{event.CursorEnd, '-'}, {event.Delete, '-'},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch})
		}
		time.Sleep(100 * time.Millisecond)
		ui.Emit(event.Event{Type: event.WriteQuit})
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
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, e := range []struct {
			typ event.Type
			ch  rune
		}{
			{event.StartCmdlineCommand, ':'}, {event.Rune, 'q'}, {event.Rune, 'u'}, {event.Rune, 'i'},
			{event.Rune, 't'},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch})
		}
		time.Sleep(100 * time.Millisecond)
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.err; err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}
