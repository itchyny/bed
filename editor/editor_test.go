package editor

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/itchyny/bed/cmdline"
	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/key"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/state"
	"github.com/itchyny/bed/window"
)

type testUI struct {
	eventCh  chan<- event.Event
	initCh   chan struct{}
	redrawCh chan struct{}
}

func newTestUI() *testUI {
	return &testUI{
		initCh:   make(chan struct{}),
		redrawCh: make(chan struct{}),
	}
}

func (ui *testUI) Init(eventCh chan<- event.Event) error {
	ui.eventCh = eventCh
	go func() { defer close(ui.initCh); <-ui.redrawCh }()
	return nil
}

func (*testUI) Run(map[mode.Mode]*key.Manager) {}

func (*testUI) Size() (int, int) { return 90, 20 }

func (ui *testUI) Redraw(state.State) error {
	ui.redrawCh <- struct{}{}
	return nil
}

func (*testUI) Close() error { return nil }

func (ui *testUI) Emit(e event.Event) {
	<-ui.initCh
	ui.eventCh <- e
	switch e.Type {
	case event.ExecuteCmdline, event.NextSearch, event.PreviousSearch:
		<-ui.redrawCh
	}
	<-ui.redrawCh
}

func createTemp(dir, str string) (*os.File, error) {
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		return nil, err
	}
	if str != "" {
		if _, err = f.WriteString(str); err != nil {
			return nil, err
		}
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	if str == "" {
		if err = os.Remove(f.Name()); err != nil {
			return nil, err
		}
	}
	return f, nil
}

func TestEditorOpenEmptyWriteQuit(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.Increment, Count: 13})
		ui.Emit(event.Event{Type: event.Decrement, Count: 6})
		ui.Emit(event.Event{Type: event.Write, Arg: f.Name()})
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "1 (0x1) bytes written"; editor.err == nil ||
		!strings.HasSuffix(editor.err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, editor.err)
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
	if expected := "\x07"; string(bs) != expected {
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
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.StartInsert})
		ui.Emit(event.Event{Type: event.Rune, Rune: '4'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '8'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '0'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '0'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'f'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'a'})
		ui.Emit(event.Event{Type: event.ExitInsert})
		ui.Emit(event.Event{Type: event.CursorLeft})
		ui.Emit(event.Event{Type: event.Decrement})
		ui.Emit(event.Event{Type: event.StartInsertHead})
		ui.Emit(event.Event{Type: event.Rune, Rune: '1'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '2'})
		ui.Emit(event.Event{Type: event.ExitInsert})
		ui.Emit(event.Event{Type: event.CursorEnd})
		ui.Emit(event.Event{Type: event.Delete})
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
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.StartInsert})
		ui.Emit(event.Event{Type: event.Rune, Rune: '4'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '8'})
		ui.Emit(event.Event{Type: event.ExitInsert})
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
		t.Fatalf("err should be nil but got: %v", err)
	}
	r := strings.NewReader("Hello, world!")
	if err := editor.Read(r); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.WriteQuit, Arg: f.Name()})
	}()
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
	str := "Hello, world! こんにちは、世界！"
	f, err := createTemp(t.TempDir(), str)
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
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
			t.Fatalf("err should be nil but got: %v", err)
		}
		if err := editor.Open(f.Name()); err != nil {
			t.Fatalf("err should be nil but got: %v", err)
		}
		fout, err := createTemp(t.TempDir(), "")
		if err != nil {
			t.Fatalf("err should be nil but got: %v", err)
		}
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
		if expected := fmt.Sprintf("%[1]d (0x%[1]x) bytes written", testCase.count); editor.err == nil ||
			!strings.Contains(editor.err.Error(), expected) {
			t.Errorf("err should be contain %q but got: %v", expected, editor.err)
		}
		if editor.errtyp != state.MessageInfo {
			t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
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
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "Hello, world!")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.CursorNext, Count: 4})
		ui.Emit(event.Event{Type: event.StartVisual})
		ui.Emit(event.Event{Type: event.CursorNext, Count: 5})
		ui.Emit(event.Event{Type: event.StartCmdlineCommand})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'w'})
		ui.Emit(event.Event{Type: event.Rune, Rune: ' '})
		for _, ch := range f.Name() + ".out" {
			ui.Emit(event.Event{Type: event.Rune, Rune: ch})
		}
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "6 (0x6) bytes written"; editor.err == nil ||
		!strings.HasSuffix(editor.err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, editor.err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(f.Name() + ".out")
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
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "abc")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.DeleteByte})
		ui.Emit(event.Event{Type: event.Write, Arg: f.Name()})
		ui.Emit(event.Event{Type: event.Undo})
		ui.Emit(event.Event{Type: event.WriteQuit, Arg: f.Name()})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "2 (0x2) bytes written"; editor.err == nil ||
		!strings.HasSuffix(editor.err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, editor.err)
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
	if expected := "abc"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorSearch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "abcdefabcdefabcdef")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.StartCmdlineSearchForward})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'e'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'f'})
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
		ui.Emit(event.Event{Type: event.Nop}) // wait for redraw
		ui.Emit(event.Event{Type: event.DeleteByte})
		ui.Emit(event.Event{Type: event.PreviousSearch})
		ui.Emit(event.Event{Type: event.NextSearch})
		ui.Emit(event.Event{Type: event.DeleteByte})
		ui.Emit(event.Event{Type: event.StartCmdlineSearchBackward})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'b'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'c'})
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
		ui.Emit(event.Event{Type: event.Nop}) // wait for redraw
		ui.Emit(event.Event{Type: event.DeleteByte})
		ui.Emit(event.Event{Type: event.PreviousSearch})
		ui.Emit(event.Event{Type: event.DeleteByte})
		ui.Emit(event.Event{Type: event.Write})
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "14 (0xe) bytes written"; editor.err == nil ||
		!strings.HasSuffix(editor.err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, editor.err)
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
	if expected := "abcdfacdfacdef"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorCmdlineQuit(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.StartCmdlineCommand})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'q'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'u'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'i'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 't'})
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.err; err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorCmdlineQuitAll(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.StartCmdlineCommand})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'q'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'a'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'l'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'l'})
		ui.Emit(event.Event{Type: event.ExecuteCmdline})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := editor.err; err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorCmdlineQuitErr(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.StartCmdlineCommand})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'c'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'q'})
		ui.Emit(event.Event{Type: event.Rune, Rune: ' '})
		ui.Emit(event.Event{Type: event.Rune, Rune: '4'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '2'})
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

func TestEditorReplace(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "Hello, world!")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.CursorNext, Count: 2})
		ui.Emit(event.Event{Type: event.StartReplace})
		ui.Emit(event.Event{Type: event.SwitchFocus})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'a'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'b'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'c'})
		ui.Emit(event.Event{Type: event.CursorNext, Count: 2})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'd'})
		ui.Emit(event.Event{Type: event.Rune, Rune: 'e'})
		ui.Emit(event.Event{Type: event.ExitInsert})
		ui.Emit(event.Event{Type: event.CursorLeft, Count: 5})
		ui.Emit(event.Event{Type: event.StartReplaceByte})
		ui.Emit(event.Event{Type: event.SwitchFocus})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '2'})
		ui.Emit(event.Event{Type: event.CursorNext, Count: 2})
		ui.Emit(event.Event{Type: event.StartReplace})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '2'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '3'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '4'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '5'})
		ui.Emit(event.Event{Type: event.Backspace})
		ui.Emit(event.Event{Type: event.ExitInsert})
		ui.Emit(event.Event{Type: event.CursorEnd})
		ui.Emit(event.Event{Type: event.StartReplace})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '6'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '7'})
		ui.Emit(event.Event{Type: event.Rune, Rune: '8'})
		ui.Emit(event.Event{Type: event.Backspace})
		ui.Emit(event.Event{Type: event.ExitInsert})
		ui.Emit(event.Event{Type: event.CursorHead})
		ui.Emit(event.Event{Type: event.DeleteByte})
		ui.Emit(event.Event{Type: event.Write, Arg: f.Name() + ".out"})
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "13 (0xd) bytes written"; editor.err == nil ||
		!strings.HasSuffix(editor.err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, editor.err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(f.Name() + ".out")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "earcrsterldvw"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorCopyCutPaste(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "Hello, world!")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.CursorNext, Count: 2})
		ui.Emit(event.Event{Type: event.StartVisual})
		ui.Emit(event.Event{Type: event.CursorNext, Count: 5})
		ui.Emit(event.Event{Type: event.Copy})
		ui.Emit(event.Event{Type: event.CursorNext, Count: 3})
		ui.Emit(event.Event{Type: event.Paste})
		ui.Emit(event.Event{Type: event.CursorPrev, Count: 2})
		ui.Emit(event.Event{Type: event.StartVisual})
		ui.Emit(event.Event{Type: event.CursorPrev, Count: 5})
		ui.Emit(event.Event{Type: event.Cut})
		ui.Emit(event.Event{Type: event.CursorNext, Count: 5})
		ui.Emit(event.Event{Type: event.PastePrev})
		ui.Emit(event.Event{Type: event.Write, Arg: f.Name() + ".out"})
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "19 (0x13) bytes written"; editor.err == nil ||
		!strings.HasSuffix(editor.err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, editor.err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(f.Name() + ".out")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "Hell w woo,llo,rld!"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorShowBinary(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "Hello, world!")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.ShowBinary})
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "01001000"; editor.err.Error() != expected {
		t.Errorf("err should be %q but got: %v", expected, editor.err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorShowDecimal(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "Hello, world!")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.ShowDecimal})
		ui.Emit(event.Event{Type: event.Quit})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "72"; editor.err.Error() != expected {
		t.Errorf("err should be %q but got: %v", expected, editor.err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}

func TestEditorShift(t *testing.T) {
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	f, err := createTemp(t.TempDir(), "Hello, world!")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		ui.Emit(event.Event{Type: event.ShiftLeft, Count: 1})
		ui.Emit(event.Event{Type: event.CursorNext, Count: 7})
		ui.Emit(event.Event{Type: event.ShiftRight, Count: 3})
		ui.Emit(event.Event{Type: event.Write, Arg: f.Name() + ".out"})
		ui.Emit(event.Event{Type: event.Quit, Bang: true})
	}()
	if err := editor.Run(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "13 (0xd) bytes written"; editor.err == nil ||
		!strings.HasSuffix(editor.err.Error(), expected) {
		t.Errorf("err should end with %q but got: %v", expected, editor.err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(f.Name() + ".out")
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "\x90ello, \x0eorld!"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
}

func TestEditorChdir(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	ui := newTestUI()
	editor := NewEditor(ui, window.NewManager(), cmdline.NewCmdline())
	if err := editor.Init(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := editor.OpenEmpty(); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
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
		t.Errorf("err should be %q but got: %v", dir, err)
	}
	if editor.errtyp != state.MessageInfo {
		t.Errorf("errtyp should be MessageInfo but got: %v", editor.errtyp)
	}
	if err := editor.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
}
