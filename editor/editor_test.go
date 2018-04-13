package editor

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
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
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
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
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := ioutil.TempFile("", "bed-test-editor-open-write-quit")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
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

func TestEditorWritePartial(t *testing.T) {
	f, err := ioutil.TempFile("", "bed-test-editor-write-partial")
	defer os.Remove(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	str := "Hello, world! こんにちは、世界！"
	n, err := f.WriteString(str)
	if n != 41 {
		t.Errorf("WriteString should return %d but got %d", 41, n)
	}
	if err != nil {
		t.Errorf("err should be nil but got %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	for i, testCase := range []struct {
		cmdRange string
		count    int
		expected string
	}{
		{"", 41, str},
		{"-10,$+10", 41, str},
		{"10,25", 16, str[10:26]},
		{".+3+3+3+5+5 , .+0xa-0x6", 16, str[4:20]},
		{"$-20,.+28", 9, str[20:29]},
	} {
		ui := newTestUI()
		editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
		if err := editor.Init(); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if err := editor.Open(f.Name()); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		name := "editor-partial-test-" + strconv.Itoa(i)
		defer os.Remove(name)
		go func(name string) {
			ui.Emit(event.Event{Type: event.StartCmdlineCommand})
			for _, c := range testCase.cmdRange + "w " + name {
				ui.Emit(event.Event{Type: event.Rune, Rune: c})
			}
			ui.Emit(event.Event{Type: event.ExecuteCmdline})
			time.Sleep(100 * time.Millisecond)
			ui.Emit(event.Event{Type: event.WriteQuit})
		}(name)
		if err := editor.Run(); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		expectedErr := fmt.Sprintf("%d (0x%x) bytes written", testCase.count, testCase.count)
		if editor.err == nil || !strings.Contains(editor.err.Error(), expectedErr) {
			t.Errorf("err should be contain %q but got: %v", expectedErr, editor.err)
		}
		if err := editor.Close(); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		bs, err := ioutil.ReadFile(name)
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if string(bs) != testCase.expected {
			t.Errorf("file contents should be %q but got %q", testCase.expected, string(bs))
		}
	}
}

func TestEditorWriteVisualSelection(t *testing.T) {
	f, err := ioutil.TempFile("", "bed-test-editor-write-visual-selection")
	defer os.Remove(f.Name())
	defer os.Remove(f.Name() + "_")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	str := "Hello, world!"
	n, err := f.WriteString(str)
	if n != 13 {
		t.Errorf("WriteString should return %d but got %d", 13, n)
	}
	if err != nil {
		t.Errorf("err should be nil but got %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, e := range []struct {
			typ   event.Type
			ch    rune
			count int64
		}{
			{event.CursorNext, 'w', 4}, {event.StartVisual, 'v', 0},
			{event.CursorNext, 'w', 5}, {event.StartCmdlineCommand, ':', 0},
			{event.Rune, 'w', 0}, {event.Rune, ' ', 0},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch, Count: e.count})
		}
		for _, ch := range f.Name() + "-" {
			ui.Emit(event.Event{Type: event.Rune, Rune: ch})
		}
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
		time.Sleep(100 * time.Millisecond)
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.err; !strings.HasSuffix(err.Error(), "6 (0x6) bytes written") {
		t.Errorf("err should be ends with %q but got: %v", "6 (0x6) bytes written", err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := ioutil.ReadFile(f.Name() + "-")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if string(bs) != "o, wor" {
		t.Errorf("file contents should be %q but got %q", "o, wor", string(bs))
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
