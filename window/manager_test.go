package window

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/itchyny/bed/buffer"
	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/layout"
	"github.com/itchyny/bed/mode"
)

func createTemp(dir, contents string) (*os.File, error) {
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		return nil, err
	}
	if _, err = f.WriteString(contents); err != nil {
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	return f, nil
}

func TestManagerOpenEmpty(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh, waitCh := make(chan event.Event), make(chan struct{}), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	go func() {
		defer func() {
			close(eventCh)
			close(redrawCh)
			close(waitCh)
		}()
		ev := <-eventCh
		if ev.Type != event.Error {
			t.Errorf("event type should be %d but got: %d", event.Error, ev.Type)
		}
		if expected := "no file name"; ev.Error.Error() != expected {
			t.Errorf("err should be %q but got: %v", expected, ev.Error)
		}
	}()
	wm.SetSize(110, 20)
	if err := wm.Open(""); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	windowStates, _, windowIndex, err := wm.State()
	ws := windowStates[0]
	if windowIndex != 0 {
		t.Errorf("window index should be %d but got %d", 0, windowIndex)
	}
	if expected := ""; ws.Name != expected {
		t.Errorf("name should be %q but got %q", expected, ws.Name)
	}
	if ws.Width != 16 {
		t.Errorf("width should be %d but got %d", 16, ws.Width)
	}
	if ws.Size != 0 {
		t.Errorf("size should be %d but got %d", 0, ws.Size)
	}
	if ws.Length != int64(0) {
		t.Errorf("Length should be %d but got %d", int64(0), ws.Length)
	}
	if expected := "\x00"; !strings.HasPrefix(string(ws.Bytes), expected) {
		t.Errorf("Bytes should start with %q but got %q", expected, string(ws.Bytes))
	}
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	wm.Emit(event.Event{Type: event.Write})
	<-waitCh
	wm.Close()
}

func TestManagerOpenStates(t *testing.T) {
	wm := NewManager()
	wm.Init(nil, nil)
	wm.SetSize(110, 20)
	str := "Hello, world! こんにちは、世界！"
	f, err := createTemp(t.TempDir(), str)
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := wm.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	windowStates, _, windowIndex, err := wm.State()
	ws := windowStates[0]
	if windowIndex != 0 {
		t.Errorf("window index should be %d but got %d", 0, windowIndex)
	}
	if expected := filepath.Base(f.Name()); ws.Name != expected {
		t.Errorf("name should be %q but got %q", expected, ws.Name)
	}
	if ws.Width != 16 {
		t.Errorf("width should be %d but got %d", 16, ws.Width)
	}
	if ws.Size != 41 {
		t.Errorf("size should be %d but got %d", 41, ws.Size)
	}
	if ws.Length != int64(41) {
		t.Errorf("Length should be %d but got %d", int64(41), ws.Length)
	}
	if !strings.HasPrefix(string(ws.Bytes), str) {
		t.Errorf("Bytes should start with %q but got %q", str, string(ws.Bytes))
	}
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	wm.Close()
}

func TestManagerOpenNonExistsWrite(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh, waitCh := make(chan event.Event), make(chan struct{}), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	go func() {
		defer func() {
			close(eventCh)
			close(redrawCh)
			close(waitCh)
		}()
		for range 16 {
			<-redrawCh
		}
		if ev := <-eventCh; ev.Type != event.QuitAll {
			t.Errorf("event type should be %d but got: %d", event.QuitAll, ev.Type)
		}
	}()
	wm.SetSize(110, 20)
	fname := filepath.Join(t.TempDir(), "test")
	if err := wm.Open(fname); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	_, _, _, _ = wm.State()
	str := "Hello, world!"
	wm.Emit(event.Event{Type: event.StartInsert})
	wm.Emit(event.Event{Type: event.SwitchFocus})
	for _, c := range str {
		wm.Emit(event.Event{Type: event.Rune, Rune: c, Mode: mode.Insert})
	}
	wm.Emit(event.Event{Type: event.ExitInsert})
	wm.Emit(event.Event{Type: event.WriteQuit})
	windowStates, _, windowIndex, err := wm.State()
	ws := windowStates[0]
	if windowIndex != 0 {
		t.Errorf("window index should be %d but got %d", 0, windowIndex)
	}
	if expected := filepath.Base(fname); ws.Name != expected {
		t.Errorf("name should be %q but got %q", expected, ws.Name)
	}
	if ws.Width != 16 {
		t.Errorf("width should be %d but got %d", 16, ws.Width)
	}
	if ws.Size != 13 {
		t.Errorf("size should be %d but got %d", 13, ws.Size)
	}
	if ws.Length != int64(13) {
		t.Errorf("Length should be %d but got %d", int64(13), ws.Length)
	}
	if !strings.HasPrefix(string(ws.Bytes), str) {
		t.Errorf("Bytes should start with %q but got %q", str, string(ws.Bytes))
	}
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	bs, err := os.ReadFile(fname)
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if string(bs) != str {
		t.Errorf("file contents should be %q but got %q", str, string(bs))
	}
	<-waitCh
	wm.Close()
}

func TestManagerOpenExpandBacktick(t *testing.T) {
	wm := NewManager()
	wm.Init(nil, nil)
	wm.SetSize(110, 20)
	cmd, name := "`which ls`", "ls"
	if runtime.GOOS == "windows" {
		cmd, name = "`where ping`", "PING.EXE"
	}
	if err := wm.Open(cmd); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	windowStates, _, _, err := wm.State()
	ws := windowStates[0]
	if ws.Name != name {
		t.Errorf("name should be %q but got %q", name, ws.Name)
	}
	if ws.Width != 16 {
		t.Errorf("width should be %d but got %d", 16, ws.Width)
	}
	if ws.Size == 0 {
		t.Errorf("size should not be %d but got %d", 0, ws.Size)
	}
	if ws.Length == 0 {
		t.Errorf("length should not be %d but got %d", 0, ws.Length)
	}
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	wm.Close()
}

func TestManagerOpenExpandHomedir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	wm := NewManager()
	wm.Init(nil, nil)
	wm.SetSize(110, 20)
	str := "Hello, world!"
	f, err := createTemp(t.TempDir(), str)
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	home := os.Getenv("HOME")
	t.Cleanup(func() {
		if err := os.Setenv("HOME", home); err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
	})
	if err := os.Setenv("HOME", filepath.Dir(f.Name())); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	for i, prefix := range []string{"~/", "$HOME/"} {
		if err := wm.Open(prefix + filepath.Base(f.Name())); err != nil {
			t.Fatalf("err should be nil but got: %v", err)
		}
		windowStates, _, windowIndex, err := wm.State()
		ws := windowStates[i]
		if windowIndex != i {
			t.Errorf("window index should be %d but got %d", i, windowIndex)
		}
		if expected := filepath.Base(f.Name()); ws.Name != expected {
			t.Errorf("name should be %q but got %q", expected, ws.Name)
		}
		if !strings.HasPrefix(string(ws.Bytes), str) {
			t.Errorf("Bytes should start with %q but got %q", str, string(ws.Bytes))
		}
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
	}
	wm.Close()
}

func TestManagerOpenChdirWrite(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip on Windows")
	}
	wm := NewManager()
	eventCh, redrawCh, waitCh := make(chan event.Event), make(chan struct{}), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	f, err := createTemp(t.TempDir(), "Hello")
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	go func() {
		defer func() {
			close(eventCh)
			close(redrawCh)
			close(waitCh)
		}()
		ev := <-eventCh
		if ev.Type != event.Info {
			t.Errorf("event type should be %d but got: %d", event.Info, ev.Type)
		}
		dir, err := filepath.EvalSymlinks(filepath.Dir(f.Name()))
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if expected := dir; ev.Error.Error() != expected {
			t.Errorf("err should be %q but got: %v", expected, ev.Error)
		}
		ev = <-eventCh
		if ev.Type != event.Info {
			t.Errorf("event type should be %d but got: %d", event.Info, ev.Type)
		}
		if expected := filepath.Dir(dir); ev.Error.Error() != expected {
			t.Errorf("err should be %q but got: %v", expected, ev.Error)
		}
		for range 11 {
			<-redrawCh
		}
		ev = <-eventCh
		if ev.Type != event.Info {
			t.Errorf("event type should be %d but got: %d", event.Info, ev.Type)
		}
		if expected := "13 (0xd) bytes written"; !strings.HasSuffix(ev.Error.Error(), expected) {
			t.Errorf("err should be %q but got: %v", expected, ev.Error)
		}
	}()
	wm.SetSize(110, 20)
	wm.Emit(event.Event{Type: event.Chdir, Arg: filepath.Dir(f.Name())})
	if err := wm.Open(filepath.Base(f.Name())); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	_, _, windowIndex, _ := wm.State()
	if expected := 0; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}
	wm.Emit(event.Event{Type: event.Chdir, Arg: "../"})
	wm.Emit(event.Event{Type: event.StartAppendEnd})
	wm.Emit(event.Event{Type: event.SwitchFocus})
	for _, c := range ", world!" {
		wm.Emit(event.Event{Type: event.Rune, Rune: c, Mode: mode.Insert})
	}
	wm.Emit(event.Event{Type: event.ExitInsert})
	wm.Emit(event.Event{Type: event.Write})
	bs, err := os.ReadFile(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if expected := "Hello, world!"; string(bs) != expected {
		t.Errorf("file contents should be %q but got %q", expected, string(bs))
	}
	<-waitCh
	wm.Close()
}

func TestManagerOpenDirectory(t *testing.T) {
	wm := NewManager()
	wm.Init(nil, nil)
	wm.SetSize(110, 20)
	dir := t.TempDir()
	if err := wm.Open(dir); err != nil {
		if expected := dir + " is a directory"; err.Error() != expected {
			t.Errorf("err should be %q but got: %v", expected, err)
		}
	} else {
		t.Errorf("err should not be nil but got: %v", err)
	}
	wm.Close()
}

func TestManagerRead(t *testing.T) {
	wm := NewManager()
	wm.Init(nil, nil)
	wm.SetSize(110, 20)
	r := strings.NewReader("Hello, world!")
	if err := wm.Read(r); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	windowStates, _, windowIndex, err := wm.State()
	ws := windowStates[0]
	if windowIndex != 0 {
		t.Errorf("window index should be %d but got %d", 0, windowIndex)
	}
	if ws.Name != "" {
		t.Errorf("name should be %q but got %q", "", ws.Name)
	}
	if ws.Width != 16 {
		t.Errorf("width should be %d but got %d", 16, ws.Width)
	}
	if ws.Size != 13 {
		t.Errorf("size should be %d but got %d", 13, ws.Size)
	}
	if ws.Length != int64(13) {
		t.Errorf("Length should be %d but got %d", int64(13), ws.Length)
	}
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	wm.Close()
}

func TestManagerOnly(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh, waitCh := make(chan event.Event), make(chan struct{}), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	go func() {
		defer func() {
			close(eventCh)
			close(redrawCh)
			close(waitCh)
		}()
		for range 4 {
			<-eventCh
		}
	}()
	wm.SetSize(110, 20)
	if err := wm.Open(""); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	wm.Emit(event.Event{Type: event.Vnew})
	wm.Emit(event.Event{Type: event.Vnew})
	wm.Emit(event.Event{Type: event.FocusWindowRight})
	wm.Resize(110, 20)

	_, got, _, _ := wm.State()
	expected := layout.NewLayout(0).SplitLeft(1).SplitLeft(2).
		Activate(1).Resize(0, 0, 110, 20)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("layout should be %#v but got %#v", expected, got)
	}

	wm.Emit(event.Event{Type: event.Only})
	wm.Resize(110, 20)
	_, got, _, _ = wm.State()
	expected = layout.NewLayout(1).Resize(0, 0, 110, 20)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("layout should be %#v but got %#v", expected, got)
	}

	<-waitCh
	wm.Close()
}

func TestManagerAlternative(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh, waitCh := make(chan event.Event), make(chan struct{}), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	go func() {
		defer func() {
			close(eventCh)
			close(redrawCh)
			close(waitCh)
		}()
		for range 9 {
			<-eventCh
		}
	}()
	wm.SetSize(110, 20)

	if err := os.Chdir(os.TempDir()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := wm.Open("bed-test-manager-alternative-1"); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	if err := wm.Open("bed-test-manager-alternative-2"); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	wm.Emit(event.Event{Type: event.Alternative})
	_, _, windowIndex, _ := wm.State()
	if expected := 0; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	if err := wm.Open("bed-test-manager-alternative-3"); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	_, _, windowIndex, _ = wm.State()
	if expected := 2; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative})
	_, _, windowIndex, _ = wm.State()
	if expected := 0; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative})
	_, _, windowIndex, _ = wm.State()
	if expected := 2; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	if err := wm.Open("bed-test-manager-alternative-4"); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	_, _, windowIndex, _ = wm.State()
	if expected := 3; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative, Count: 2})
	_, _, windowIndex, _ = wm.State()
	if expected := 1; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative, Count: 4})
	_, _, windowIndex, _ = wm.State()
	if expected := 3; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative})
	_, _, windowIndex, _ = wm.State()
	if expected := 1; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Edit, Arg: "#2"})
	_, _, windowIndex, _ = wm.State()
	if expected := 1; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Edit, Arg: "#4"})
	_, _, windowIndex, _ = wm.State()
	if expected := 3; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Edit, Arg: "#"})
	_, _, windowIndex, _ = wm.State()
	if expected := 1; windowIndex != expected {
		t.Errorf("windowIndex should be %d but got %d", expected, windowIndex)
	}

	<-waitCh
	wm.Close()
}

func TestManagerWincmd(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh, waitCh := make(chan event.Event), make(chan struct{}), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	go func() {
		defer func() {
			close(eventCh)
			close(redrawCh)
			close(waitCh)
		}()
		for range 17 {
			<-eventCh
		}
	}()
	wm.SetSize(110, 20)
	if err := wm.Open(""); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	wm.Emit(event.Event{Type: event.Wincmd, Arg: "n"})
	wm.Emit(event.Event{Type: event.Wincmd, Arg: "n"})
	wm.Emit(event.Event{Type: event.Wincmd, Arg: "n"})
	wm.Emit(event.Event{Type: event.MoveWindowLeft})
	wm.Emit(event.Event{Type: event.FocusWindowRight})
	wm.Emit(event.Event{Type: event.FocusWindowBottomRight})
	wm.Emit(event.Event{Type: event.MoveWindowRight})
	wm.Emit(event.Event{Type: event.FocusWindowLeft})
	wm.Emit(event.Event{Type: event.MoveWindowTop})
	wm.Resize(110, 20)

	_, got, _, _ := wm.State()
	expected := layout.NewLayout(2).SplitBottom(0).SplitLeft(1).
		SplitLeft(3).Activate(2).Resize(0, 0, 110, 20)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("layout should be %#v but got %#v", expected, got)
	}

	wm.Emit(event.Event{Type: event.FocusWindowDown})
	wm.Emit(event.Event{Type: event.FocusWindowRight})
	wm.Emit(event.Event{Type: event.Quit})
	_, got, _, _ = wm.State()
	expected = layout.NewLayout(2).SplitBottom(0).SplitLeft(3).Resize(0, 0, 110, 20)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("layout should be %#v but got %#v", expected, got)
	}

	wm.Emit(event.Event{Type: event.Wincmd, Arg: "o"})
	_, got, _, _ = wm.State()
	expected = layout.NewLayout(3).Resize(0, 0, 110, 20)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("layout should be %#v but got %#v", expected, got)
	}

	wm.Emit(event.Event{Type: event.MoveWindowLeft})
	wm.Emit(event.Event{Type: event.MoveWindowRight})
	wm.Emit(event.Event{Type: event.MoveWindowTop})
	wm.Emit(event.Event{Type: event.MoveWindowBottom})
	wm.Resize(110, 20)
	_, got, _, _ = wm.State()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("layout should be %#v but got %#v", expected, got)
	}

	<-waitCh
	wm.Close()
}

func TestManagerCopyCutPaste(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh, waitCh := make(chan event.Event), make(chan struct{}), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	str := "Hello, world!"
	f, err := createTemp(t.TempDir(), str)
	if err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	wm.SetSize(110, 20)
	if err := wm.Open(f.Name()); err != nil {
		t.Fatalf("err should be nil but got: %v", err)
	}
	_, _, _, _ = wm.State()
	go func() {
		defer func() {
			close(eventCh)
			close(redrawCh)
			close(waitCh)
		}()
		<-redrawCh
		<-redrawCh
		<-redrawCh
		waitCh <- struct{}{}
		ev := <-eventCh
		if ev.Type != event.Copied {
			t.Errorf("event type should be %d but got: %d", event.Copied, ev.Type)
		}
		if ev.Buffer == nil {
			t.Errorf("Buffer should not be nil but got: %#v", ev)
		}
		if expected := "yanked"; ev.Arg != expected {
			t.Errorf("Arg should be %q but got: %q", expected, ev.Arg)
		}
		p := make([]byte, 20)
		_, _ = ev.Buffer.ReadAt(p, 0)
		if !strings.HasPrefix(string(p), "lo, worl") {
			t.Errorf("buffer string should be %q but got: %q", "", string(p))
		}
		waitCh <- struct{}{}
		<-redrawCh
		<-redrawCh
		waitCh <- struct{}{}
		ev = <-eventCh
		if ev.Type != event.Copied {
			t.Errorf("event type should be %d but got: %d", event.Copied, ev.Type)
		}
		if ev.Buffer == nil {
			t.Errorf("Buffer should not be nil but got: %#v", ev)
		}
		if expected := "deleted"; ev.Arg != expected {
			t.Errorf("Arg should be %q but got: %q", expected, ev.Arg)
		}
		p = make([]byte, 20)
		_, _ = ev.Buffer.ReadAt(p, 0)
		if !strings.HasPrefix(string(p), "lo, wo") {
			t.Errorf("buffer string should be %q but got: %q", "", string(p))
		}
		windowStates, _, _, _ := wm.State()
		ws := windowStates[0]
		if ws.Length != int64(7) {
			t.Errorf("Length should be %d but got %d", int64(7), ws.Length)
		}
		if expected := "Helrld!"; !strings.HasPrefix(string(ws.Bytes), expected) {
			t.Errorf("Bytes should start with %q but got %q", expected, string(ws.Bytes))
		}
		waitCh <- struct{}{}
		<-redrawCh
		waitCh <- struct{}{}
		ev = <-eventCh
		if ev.Type != event.Pasted {
			t.Errorf("event type should be %d but got: %d", event.Pasted, ev.Type)
		}
		if ev.Count != 18 {
			t.Errorf("Count should be %d but got: %d", 18, ev.Count)
		}
		windowStates, _, _, _ = wm.State()
		ws = windowStates[0]
		if ws.Length != int64(25) {
			t.Errorf("Length should be %d but got %d", int64(25), ws.Length)
		}
		if expected := "Hefoobarfoobarfoobarlrld!"; !strings.HasPrefix(string(ws.Bytes), expected) {
			t.Errorf("Bytes should start with %q but got %q", expected, string(ws.Bytes))
		}
	}()
	wm.Emit(event.Event{Type: event.CursorNext, Mode: mode.Normal, Count: 3})
	wm.Emit(event.Event{Type: event.StartVisual})
	wm.Emit(event.Event{Type: event.CursorNext, Mode: mode.Visual, Count: 7})
	<-waitCh
	wm.Emit(event.Event{Type: event.Copy})
	<-waitCh
	wm.Emit(event.Event{Type: event.StartVisual})
	wm.Emit(event.Event{Type: event.CursorNext, Mode: mode.Visual, Count: 5})
	<-waitCh
	wm.Emit(event.Event{Type: event.Cut})
	<-waitCh
	wm.Emit(event.Event{Type: event.CursorPrev, Mode: mode.Normal, Count: 2})
	<-waitCh
	wm.Emit(event.Event{Type: event.Paste, Buffer: buffer.NewBuffer(strings.NewReader("foobar")), Count: 3})
	<-waitCh
	wm.Close()
}
