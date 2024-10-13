package editor

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
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
	initCh  chan struct{}
	mu      *sync.Mutex
}

func newTestUI() *testUI {
	return &testUI{initCh: make(chan struct{}), mu: new(sync.Mutex)}
}

func (ui *testUI) Init(eventCh chan<- event.Event) error {
	defer close(ui.initCh)
	ui.mu.Lock()
	defer ui.mu.Unlock()
	ui.eventCh = eventCh
	return nil
}

func (ui *testUI) Run(map[mode.Mode]*key.Manager) {}

func (ui *testUI) Height() int { return 10 }

func (ui *testUI) Size() (int, int) { return 90, 20 }

func (ui *testUI) Redraw(_ state.State) error { return nil }

func (ui *testUI) Close() error { return nil }

func (ui *testUI) Emit(e event.Event) {
	<-ui.initCh
	ui.mu.Lock()
	defer ui.mu.Unlock()
	if e.Type == event.ExecuteCmdline {
		time.Sleep(100 * time.Millisecond)
	}
	ui.eventCh <- e
	switch e.Type {
	case event.Write, event.WriteQuit, event.StartCmdlineCommand, event.ExecuteCmdline:
		time.Sleep(500 * time.Millisecond)
	case event.Rune:
	default:
		time.Sleep(10 * time.Millisecond)
	}
}

func TestEditorOpenEmptyWriteQuit(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := os.CreateTemp("", "bed-test-editor-open-empty-write-quit")
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
		ui.Emit(event.Event{Type: event.Write, Arg: f.Name()})
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
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
	bs, err := os.ReadFile(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "\x01"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
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
	f, err := os.CreateTemp("", "bed-test-editor-open-write-quit")
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
			{event.StartInsert, '-'},
			{event.Rune, '4'},
			{event.Rune, '8'},
			{event.Rune, '0'},
			{event.Rune, '0'},
			{event.Rune, 'f'},
			{event.Rune, 'a'},
			{event.ExitInsert, '-'},
			{event.CursorLeft, '-'},
			{event.Decrement, '-'},
			{event.StartInsertHead, '-'},
			{event.Rune, '1'},
			{event.Rune, '2'},
			{event.ExitInsert, '-'},
			{event.CursorEnd, '-'},
			{event.Delete, '-'},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch})
		}
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
	bs, err := os.ReadFile(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "\x12\x48\xff"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorOpenForceQuit(t *testing.T) {
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
			{event.StartInsert, '-'}, {event.Rune, '4'}, {event.Rune, '8'}, {event.ExitInsert, '-'},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch})
		}
		ui.Emit(event.Event{Type: event.Quit})
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err, expected := editor.err, "you have unsaved changes in [No Name], "+
		"add ! to force :quit"; err == nil || !strings.HasSuffix(err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorReadWriteQuit(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	r := strings.NewReader("Hello, world!")
	if err := editor.Read(r); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := os.CreateTemp("", "bed-test-editor-read-write-quit")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go ui.Emit(event.Event{Type: event.WriteQuit, Arg: f.Name()})
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "Hello, world!"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorWritePartial(t *testing.T) {
	f, err := os.CreateTemp("", "bed-test-editor-write-partial")
	defer os.Remove(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	str := "Hello, world! こんにちは、世界！"
	n, err := f.WriteString(str)
	if expected := 41; n != expected {
		t.Errorf("WriteString should return %d but got %d", expected, n)
	}
	if err != nil {
		t.Errorf("err should be nil but got %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	for _, testCase := range []struct {
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
		fout, err := os.CreateTemp("", "bed-test-editor-write-partial")
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		defer os.Remove(fout.Name())
		fout.Close()
		go func(name string) {
			ui.Emit(event.Event{Type: event.StartCmdlineCommand})
			for _, c := range testCase.cmdRange + "w " + name {
				ui.Emit(event.Event{Type: event.Rune, Rune: c})
			}
			ui.Emit(event.Event{Type: event.ExecuteCmdline})
			ui.Emit(event.Event{Type: event.Quit, Bang: true})
		}(fout.Name())
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
		bs, err := os.ReadFile(fout.Name())
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if string(bs) != testCase.expected {
			t.Errorf("file contents should be %q but got %q", testCase.expected, string(bs))
		}
	}
}

func TestEditorWriteVisualSelection(t *testing.T) {
	f1, err := os.CreateTemp("", "bed-test-editor-write-visual-selection1")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	defer os.Remove(f1.Name())
	f2, err := os.CreateTemp("", "bed-test-editor-write-visual-selection2")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	defer os.Remove(f2.Name())
	str := "Hello, world!"
	n, err := f1.WriteString(str)
	if expected := 13; n != expected {
		t.Errorf("WriteString should return %d but got %d", expected, n)
	}
	if err != nil {
		t.Errorf("err should be nil but got %v", err)
	}
	if err := f1.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := f2.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f1.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, e := range []struct {
			typ   event.Type
			ch    rune
			count int64
		}{
			{event.CursorNext, 'w', 4},
			{event.StartVisual, 'v', 0},
			{event.CursorNext, 'w', 5},
			{event.StartCmdlineCommand, ':', 0},
			{event.Rune, 'w', 0},
			{event.Rune, ' ', 0},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch, Count: e.count})
		}
		for _, ch := range f2.Name() {
			ui.Emit(event.Event{Type: event.Rune, Rune: ch})
		}
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err, expected := editor.err, "6 (0x6) bytes written"; !strings.HasSuffix(err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(f2.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "o, wor"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorWriteUndo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := os.CreateTemp("", "bed-test-editor-write-undo")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if _, err := f.WriteString("abc"); err != nil {
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
			arg string
		}{
			{event.DeleteByte, 'x', ""},
			{event.Write, 'w', f.Name()},
			{event.Undo, 'u', ""},
			{event.WriteQuit, 'x', f.Name()},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch})
		}
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err, expected := editor.err, "2 (0x2) bytes written"; !strings.HasSuffix(err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "abc"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
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
			{event.StartCmdlineCommand, ':'},
			{event.Rune, 'q'},
			{event.Rune, 'u'},
			{event.Rune, 'i'},
			{event.Rune, 't'},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch})
		}
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.err; err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorReplace(t *testing.T) {
	f1, _ := os.CreateTemp("", "bed-test-editor-replace1")
	f2, _ := os.CreateTemp("", "bed-test-editor-replace2")
	defer os.Remove(f1.Name())
	defer os.Remove(f2.Name())
	str := "Hello, world!"
	_, _ = f1.WriteString(str)
	_ = f1.Close()
	_ = f2.Close()
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f1.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, e := range []struct {
			typ   event.Type
			ch    rune
			count int64
			arg   string
		}{
			{event.CursorNext, 'w', 2, ""},
			{event.StartReplace, 'R', 0, ""},
			{event.SwitchFocus, '\x00', 0, ""},
			{event.Rune, 'a', 0, ""},
			{event.Rune, 'b', 0, ""},
			{event.Rune, 'c', 0, ""},
			{event.CursorNext, 'w', 2, ""},
			{event.Rune, 'd', 0, ""},
			{event.Rune, 'e', 0, ""},
			{event.ExitInsert, '\x00', 0, ""},
			{event.CursorLeft, 'b', 5, ""},
			{event.StartReplaceByte, 'r', 0, ""},
			{event.SwitchFocus, '\x00', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '2', 0, ""},
			{event.CursorNext, 'w', 2, ""},
			{event.StartReplace, 'R', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '2', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '3', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '4', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '5', 0, ""},
			{event.Backspace, '\x00', 0, ""},
			{event.ExitInsert, '\x00', 0, ""},
			{event.CursorEnd, '\x00', 0, ""},
			{event.StartReplace, '\x00', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '6', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '7', 0, ""},
			{event.Rune, '8', 0, ""},
			{event.Backspace, '\x00', 0, ""},
			{event.ExitInsert, '\x00', 0, ""},
			{event.CursorHead, '\x00', 0, ""},
			{event.DeleteByte, '\x00', 0, ""},
			{event.Write, 'w', 0, f2.Name()},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch, Count: e.count, Arg: e.arg})
		}
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err, expected := editor.err, "13 (0xd) bytes written"; !strings.HasSuffix(err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, _ := os.ReadFile(f2.Name())
	if expected := "earcrsterldvw"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorCopyCutPaste(t *testing.T) {
	f1, _ := os.CreateTemp("", "bed-test-editor-copy-cut-paste1")
	f2, _ := os.CreateTemp("", "bed-test-editor-copy-cut-paste2")
	defer os.Remove(f1.Name())
	defer os.Remove(f2.Name())
	str := "Hello, world!"
	_, _ = f1.WriteString(str)
	_ = f1.Close()
	_ = f2.Close()
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f1.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, e := range []struct {
			typ   event.Type
			ch    rune
			count int64
			arg   string
		}{
			{event.CursorNext, 'w', 2, ""},
			{event.StartVisual, 'v', 0, ""},
			{event.CursorNext, 'w', 5, ""},
			{event.Copy, 'y', 0, ""},
			{event.CursorNext, 'w', 3, ""},
			{event.Paste, 'p', 0, ""},
			{event.CursorPrev, 'b', 2, ""},
			{event.StartVisual, 'v', 0, ""},
			{event.CursorPrev, 'b', 5, ""},
			{event.Cut, 'd', 0, ""},
			{event.CursorNext, 'w', 5, ""},
			{event.PastePrev, 'P', 0, ""},
			{event.Write, 'w', 0, f2.Name()},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch, Count: e.count, Arg: e.arg})
		}
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err, expected := editor.err, "19 (0x13) bytes written"; !strings.HasSuffix(err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, _ := os.ReadFile(f2.Name())
	if expected := "Hell w woo,llo,rld!"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorShowBinary(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := os.CreateTemp("", "bed-test-editor-show-binary")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	_, _ = f.WriteString("Hello, world!")
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	defer os.Remove(f.Name())
	if err := editor.Open(f.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.ShowBinary})
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if msg, expected := editor.err.Error(), "01001000"; msg != expected {
		t.Errorf("message should be %q but got: %q", expected, msg)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorShowDecimal(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f, err := os.CreateTemp("", "bed-test-editor-show-decimal")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	_, _ = f.WriteString("Hello, world!")
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	defer os.Remove(f.Name())
	if err := editor.Open(f.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.ShowDecimal})
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if msg, expected := editor.err.Error(), "72"; msg != expected {
		t.Errorf("message should be %q but got: %q", expected, msg)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorShift(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f1, err := os.CreateTemp("", "bed-test-editor-shift-1")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	f2, err := os.CreateTemp("", "bed-test-editor-shift-2")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	defer os.Remove(f1.Name())
	defer os.Remove(f2.Name())
	_, _ = f1.WriteString("Hello, world!")
	if err := f1.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := f2.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f1.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		for _, e := range []struct {
			typ   event.Type
			ch    rune
			count int64
		}{
			{event.ShiftLeft, '<', 1},
			{event.CursorNext, 'w', 7},
			{event.ShiftRight, '>', 3},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch, Count: e.count})
		}
		ui.Emit(event.Event{Type: event.Write, Arg: f2.Name()})
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err, expected := editor.err, "13 (0xd) bytes written"; !strings.HasSuffix(err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(f2.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "\x90ello, \x0eorld!"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorCmdlineQuitErr(t *testing.T) {
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
			{event.StartCmdlineCommand, ':'},
			{event.Rune, 'c'},
			{event.Rune, 'q'},
			{event.Rune, ' '},
			{event.Rune, '4'},
			{event.Rune, '2'},
		} {
			ui.Emit(event.Event{Type: e.typ, Rune: e.ch})
		}
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
	}()

	if err, expected := editor.Run(), (&quitErr{42}); !reflect.DeepEqual(expected, err) {
		t.Errorf("err should be %v but got: %v", expected, err)
	}
	if err := editor.err; err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorChdir(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.Pwd})
		ui.Emit(event.Event{Type: event.Chdir, Arg: "../"})
		ui.Emit(event.Event{Type: event.Chdir, Arg: "-"})
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.err; err == nil || err.Error() != dir {
		t.Errorf("err should end with %q but got: %v", dir, err)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}
