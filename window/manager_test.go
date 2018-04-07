package window

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/itchyny/bed/event"
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
		t.Errorf("Bytes should starts with %q but got %q", "\x00", string(ws.Bytes))
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
		t.Errorf("Bytes should starts with %q but got %q", str, string(ws.Bytes))
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
		t.Errorf("Bytes should starts with %q but got %q", str, string(ws.Bytes))
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
