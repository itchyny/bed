package window

import (
	"io/ioutil"
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

func TestManagerOpenEmpty(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh := make(chan event.Event), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	wm.SetSize(110, 20)
	if err := wm.Open(""); err != nil {
		t.Errorf("err should be nil but got: %v", err)
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
	if ws.Size != 0 {
		t.Errorf("size should be %d but got %d", 0, ws.Size)
	}
	if ws.Length != int64(0) {
		t.Errorf("Length should be %d but got %d", int64(0), ws.Length)
	}
	if !strings.HasPrefix(string(ws.Bytes), "\x00") {
		t.Errorf("Bytes should start with %q but got %q", "\x00", string(ws.Bytes))
	}
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	wm.Close()
}

func TestManagerOpenStates(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh := make(chan event.Event), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	wm.SetSize(110, 20)
	f, err := ioutil.TempFile("", "bed-test-manager-open")
	str := "Hello, world! こんにちは、世界！"
	n, err := f.WriteString(str)
	if n != 41 {
		t.Errorf("WriteString should return %d but got %d", 41, n)
	}
	if err != nil {
		t.Errorf("err should be nil but got %v", err)
	}
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	defer os.Remove(f.Name())
	if err := wm.Open(f.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	windowStates, _, windowIndex, err := wm.State()
	ws := windowStates[0]
	if windowIndex != 0 {
		t.Errorf("window index should be %d but got %d", 0, windowIndex)
	}
	if ws.Name != filepath.Base(f.Name()) {
		t.Errorf("name should be %q but got %q", filepath.Base(f.Name()), ws.Name)
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
	eventCh, redrawCh := make(chan event.Event), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	go func() {
		for {
			select {
			case <-eventCh:
			case <-redrawCh:
			}
		}
	}()
	wm.SetSize(110, 20)
	f, _ := ioutil.TempFile("", "bed-test-manager-open")
	_ = f.Close()
	_ = os.Remove(f.Name())
	defer os.Remove(f.Name())
	if err := wm.Open(f.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	_, _, _, _ = wm.State()
	str := "Hello, world!"
	wm.Emit(event.Event{Type: event.StartInsert})
	wm.Emit(event.Event{Type: event.SwitchFocus})
	for _, c := range str {
		wm.Emit(event.Event{Type: event.Rune, Rune: c, Mode: mode.Insert})
	}
	wm.Emit(event.Event{Type: event.ExitInsert})
	wm.Emit(event.Event{Type: event.Write})
	windowStates, _, windowIndex, err := wm.State()
	ws := windowStates[0]
	if windowIndex != 0 {
		t.Errorf("window index should be %d but got %d", 0, windowIndex)
	}
	if ws.Name != filepath.Base(f.Name()) {
		t.Errorf("name should be %q but got %q", filepath.Base(f.Name()), ws.Name)
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
	bs, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	if string(bs) != str {
		t.Errorf("file contents should be %q but got %q", str, string(bs))
	}
	wm.Close()
}

func TestManagerOpenExpandBacktick(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh := make(chan event.Event), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	wm.SetSize(110, 20)
	cmd, filename := "`which ls`", "ls"
	if runtime.GOOS == "windows" {
		cmd, filename = "`where ping`", "PING.EXE"
	}
	if err := wm.Open(cmd); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	windowStates, _, _, err := wm.State()
	ws := windowStates[0]
	if ws.Name != filename {
		t.Errorf("name should be %q but got %q", filename, ws.Name)
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

func TestManagerAlternative(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh := make(chan event.Event), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	go func() {
		for {
			select {
			case <-eventCh:
			case <-redrawCh:
			}
		}
	}()
	wm.SetSize(110, 20)

	wm.Open("bed-test-manager-alternative-1")
	wm.Open("bed-test-manager-alternative-2")
	wm.Emit(event.Event{Type: event.Alternative})
	_, _, windowIndex, _ := wm.State()
	if expected := 0; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Open("bed-test-manager-alternative-3")
	_, _, windowIndex, _ = wm.State()
	if expected := 2; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative})
	_, _, windowIndex, _ = wm.State()
	if expected := 0; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative})
	_, _, windowIndex, _ = wm.State()
	if expected := 2; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Open("bed-test-manager-alternative-4")
	_, _, windowIndex, _ = wm.State()
	if expected := 3; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative, Count: 2})
	_, _, windowIndex, _ = wm.State()
	if expected := 1; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative, Count: 4})
	_, _, windowIndex, _ = wm.State()
	if expected := 3; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Alternative})
	_, _, windowIndex, _ = wm.State()
	if expected := 1; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Edit, Arg: "#2"})
	_, _, windowIndex, _ = wm.State()
	if expected := 1; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Edit, Arg: "#4"})
	_, _, windowIndex, _ = wm.State()
	if expected := 3; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Emit(event.Event{Type: event.Edit, Arg: "#"})
	_, _, windowIndex, _ = wm.State()
	if expected := 1; windowIndex != expected {
		t.Errorf("windowIndex should be %+v but got %+v", expected, windowIndex)
	}

	wm.Close()
}

func TestManagerWincmd(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh := make(chan event.Event), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	go func() {
		for {
			select {
			case <-eventCh:
			case <-redrawCh:
			}
		}
	}()
	wm.SetSize(110, 20)
	if err := wm.Open(""); err != nil {
		t.Errorf("err should be nil but got: %v", err)
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

	wm.Close()
}

func TestManagerCopyCutPaste(t *testing.T) {
	wm := NewManager()
	eventCh, redrawCh, waitCh := make(chan event.Event), make(chan struct{}), make(chan struct{})
	wm.Init(eventCh, redrawCh)
	f, err := ioutil.TempFile("", "bed-test-manager-copy-cut-paste")
	if err != nil {
		t.Errorf("err should be nil but got %v", err)
	}
	str := "Hello, world!"
	_, err = f.WriteString(str)
	if err != nil {
		t.Errorf("err should be nil but got %v", err)
	}
	if err := f.Close(); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	defer os.Remove(f.Name())
	wm.SetSize(110, 20)
	if err := wm.Open(f.Name()); err != nil {
		t.Errorf("err should be nil but got: %v", err)
	}
	_, _, _, _ = wm.State()
	go func() {
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
		if ev.Arg != "yanked" {
			t.Errorf("Arg should be %q but got: %q", "yanked", ev.Arg)
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
		if ev.Arg != "deleted" {
			t.Errorf("Arg should be %q but got: %q", "deleted", ev.Arg)
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
		expected := "Helrld!"
		if !strings.HasPrefix(string(ws.Bytes), expected) {
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
		expected = "Hefoobarfoobarfoobarlrld!"
		if !strings.HasPrefix(string(ws.Bytes), expected) {
			t.Errorf("Bytes should start with %q but got %q", expected, string(ws.Bytes))
		}
		close(waitCh)
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
