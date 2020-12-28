package window

import (
	"bytes"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/mode"
)

func TestWindowState(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, err := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(width, height)

	s, err := window.state(width, height)
	if err != nil {
		t.Fatal(err)
	}

	if expected := "test"; s.Name != expected {
		t.Errorf("state.Name should be %q but got %q", expected, s.Name)
	}

	if s.Width != width {
		t.Errorf("state.Width should be %d but got %d", width, s.Width)
	}

	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	if expected := 13; s.Size != expected {
		t.Errorf("s.Size should be %d but got %d", expected, s.Size)
	}

	if expected := int64(13); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}

	if expected := false; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}

	if expected := byte('\x00'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}

	if !reflect.DeepEqual(s.EditedIndices, []int64{}) {
		t.Errorf("state.EditedIndices should be empty but got %v", s.EditedIndices)
	}

	expected := []byte("Hello, world!" + strings.Repeat("\x00", height*width-13))
	if !reflect.DeepEqual(s.Bytes, expected) {
		t.Errorf("s.Bytes should be %q but got %q", expected, s.Bytes)
	}
}

func TestWindowEmptyState(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, err := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(width, height)

	s, err := window.state(width, height)
	if err != nil {
		t.Fatal(err)
	}

	if expected := "test"; s.Name != expected {
		t.Errorf("state.Name should be %q but got %q", expected, s.Name)
	}

	if s.Width != width {
		t.Errorf("state.Width should be %d but got %d", width, s.Width)
	}

	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	if expected := 0; s.Size != expected {
		t.Errorf("s.Size should be %d but got %d", expected, s.Size)
	}

	if expected := int64(0); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}

	if expected := false; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}

	if expected := byte('\x00'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}

	if !reflect.DeepEqual(s.EditedIndices, []int64{}) {
		t.Errorf("state.EditedIndices should be empty but got %v", s.EditedIndices)
	}

	expected := []byte(strings.Repeat("\x00", height*width))
	if !reflect.DeepEqual(s.Bytes, expected) {
		t.Errorf("s.Bytes should be %q but got %q", expected, s.Bytes)
	}

	window.scrollDown(0)
	s, err = window.state(width, height)
	if err != nil {
		t.Fatal(err)
	}

	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}
}

func TestWindowCursorMotions(t *testing.T) {
	r := strings.NewReader(strings.Repeat("Hello, world!", 100))
	width, height := 16, 10
	window, err := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(width, height)

	s, _ := window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorDown(0)
	s, _ = window.state(width, height)
	if expected := int64(width); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorDown(1)
	s, _ = window.state(width, height)
	if expected := int64(width) * 2; s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorUp(0)
	s, _ = window.state(width, height)
	if expected := int64(width); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorDown(10)
	s, _ = window.state(width, height)
	if expected := int64(width) * 11; s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 2; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}
	if expected := " world!"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}

	window.cursorRight(mode.Normal, 3)
	s, _ = window.state(width, height)
	if expected := int64(width)*11 + 3; s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorRight(mode.Normal, 20)
	s, _ = window.state(width, height)
	if expected := int64(width)*12 - 1; s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorLeft(3)
	s, _ = window.state(width, height)
	if expected := int64(width)*12 - 4; s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorLeft(20)
	s, _ = window.state(width, height)
	if expected := int64(width) * 11; s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorPrev(154)
	s, _ = window.state(width, height)
	if expected := int64(22); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if s.Offset != int64(width) {
		t.Errorf("s.Offset should be %d but got %d", width, s.Offset)
	}

	window.cursorNext(mode.Normal, 200)
	s, _ = window.state(width, height)
	if expected := int64(222); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 4; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorNext(mode.Normal, 2000)
	s, _ = window.state(width, height)
	if expected := int64(1299); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 72; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorHead(1)
	s, _ = window.state(width, height)
	if expected := int64(1296); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 72; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorEnd(1)
	s, _ = window.state(width, height)
	if expected := int64(1299); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 72; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorUp(20)
	window.cursorEnd(1)
	s, _ = window.state(width, height)
	if expected := int64(991); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 61; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorEnd(11)
	s, _ = window.state(width, height)
	if expected := int64(1151); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 62; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorDown(30)
	s, _ = window.state(width, height)
	if expected := int64(1299); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 72; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorPrev(2000)
	s, _ = window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorDown(2000)
	s, _ = window.state(width, height)
	if expected := int64(width) * 81; s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 72; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorRight(mode.Normal, 1000)
	s, _ = window.state(width, height)
	if expected := int64(1299); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 72; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorUp(2000)
	s, _ = window.state(width, height)
	if expected := int64(3); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.windowTop(0)
	s, _ = window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.windowTop(7)
	s, _ = window.state(width, height)
	if expected := int64(96); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.windowTop(20)
	s, _ = window.state(width, height)
	if expected := int64(144); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.windowMiddle()
	s, _ = window.state(width, height)
	if expected := int64(64); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.windowBottom(0)
	s, _ = window.state(width, height)
	if expected := int64(144); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.windowBottom(7)
	s, _ = window.state(width, height)
	if expected := int64(48); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.windowBottom(20)
	s, _ = window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorGotoPos(event.Absolute{Offset: 0}, "goto")
	s, _ = window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.Absolute{Offset: 50}, "goto")
	s, _ = window.state(width, height)
	if expected := int64(50); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.Absolute{Offset: 100}, "goto")
	s, _ = window.state(width, height)
	if expected := int64(100); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.Relative{Offset: -10}, "goto")
	s, _ = window.state(width, height)
	if expected := int64(90); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.Absolute{Offset: 30}, "%")
	s, _ = window.state(width, height)
	if expected := int64(390); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.Relative{Offset: 30}, "%")
	s, _ = window.state(width, height)
	if expected := int64(780); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.End{Offset: -30}, "%")
	s, _ = window.state(width, height)
	if expected := int64(909); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.Absolute{Offset: 30}, "go[to]")
	s, _ = window.state(width, height)
	if expected := int64(480); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.Relative{Offset: 30}, "go[to]")
	s, _ = window.state(width, height)
	if expected := int64(960); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorGotoPos(event.End{Offset: -30}, "go[to]")
	s, _ = window.state(width, height)
	if expected := int64(819); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowScreenMotions(t *testing.T) {
	r := strings.NewReader(strings.Repeat("Hello, world!", 100))
	width, height := 16, 10
	window, err := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(width, height)

	s, _ := window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.pageDown()
	s, _ = window.state(width, height)
	if expected := int64(128); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(128); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.pageDownHalf()
	s, _ = window.state(width, height)
	if expected := int64(208); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(208); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.scrollDown(0)
	s, _ = window.state(width, height)
	if expected := int64(224); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(224); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.scrollUp(0)
	s, _ = window.state(width, height)
	if expected := int64(224); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(208); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.scrollDown(30)
	s, _ = window.state(width, height)
	if expected := int64(688); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(688); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.scrollUp(30)
	s, _ = window.state(width, height)
	if expected := int64(352); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(208); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.pageUpHalf()
	s, _ = window.state(width, height)
	if expected := int64(272); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(128); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.pageUp()
	s, _ = window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.pageEnd()
	s, _ = window.state(width, height)
	if expected := int64(1296); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(width) * 72; s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.pageTop()
	s, _ = window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorNext(mode.Normal, 5)
	window.scrollTop(5)
	s, _ = window.state(width, height)
	if expected := int64(85); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(80); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorDown(4)
	window.scrollTop(0)
	s, _ = window.state(width, height)
	if expected := int64(149); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(144); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.scrollTopHead(10)
	s, _ = window.state(width, height)
	if expected := int64(160); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(160); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorNext(mode.Normal, 5)
	window.scrollMiddle(12)
	s, _ = window.state(width, height)
	if expected := int64(197); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(112); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.scrollMiddleHead(15)
	s, _ = window.state(width, height)
	if expected := int64(240); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(160); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorNext(mode.Normal, 5)
	window.scrollBottom(12)
	s, _ = window.state(width, height)
	if expected := int64(197); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(48); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.cursorDown(8)
	window.scrollBottom(0)
	s, _ = window.state(width, height)
	if expected := int64(325); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(176); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}

	window.scrollBottomHead(10)
	s, _ = window.state(width, height)
	if expected := int64(160); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(16); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}
}

func TestWindowDeleteBytes(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	eventCh := make(chan event.Event)
	go func() {
		for {
			<-eventCh
		}
	}()
	window, _ := newWindow(r, "test", "test", eventCh, make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 7)
	window.deleteBytes(0)
	s, _ := window.state(width, height)
	if expected := "Hello, orld!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(7); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.deleteBytes(3)
	s, _ = window.state(width, height)
	if expected := "Hello, d!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(7); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.deleteBytes(3)
	s, _ = window.state(width, height)
	if expected := "Hello, \x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(6); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.deleteByte()
	window.deleteByte()
	window.deleteByte()
	s, _ = window.state(width, height)
	if expected := "Hell\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(3); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.deleteBytes(0)
	window.deleteBytes(0)
	window.deleteBytes(0)
	window.deleteBytes(0)
	window.deleteBytes(0)
	s, _ = window.state(width, height)
	if expected := "\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	if expected := int64(0); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
}

func TestWindowDeletePrevBytes(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	eventCh := make(chan event.Event)
	go func() {
		for {
			<-eventCh
		}
	}()
	window, _ := newWindow(r, "test", "test", eventCh, make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 5)
	window.deletePrevBytes(0)
	s, _ := window.state(width, height)
	if expected := "Hell, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(4); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.deletePrevBytes(3)
	s, _ = window.state(width, height)
	if expected := "H, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(1); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.deletePrevBytes(3)
	s, _ = window.state(width, height)
	if expected := ", world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowIncrementDecrement(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.increment(0)
	s, _ := window.state(width, height)
	if expected := "Iello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}

	window.increment(1000)
	s, _ = window.state(width, height)
	if expected := "1ello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}

	window.increment(math.MaxInt64)
	s, _ = window.state(width, height)
	if expected := "0ello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}

	window.decrement(0)
	s, _ = window.state(width, height)
	if expected := "/ello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}

	window.decrement(1000)
	s, _ = window.state(width, height)
	if expected := "Gello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}

	window.decrement(math.MaxInt64)
	s, _ = window.state(width, height)
	if expected := "Hello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}

	window.cursorNext(mode.Normal, 7)
	window.increment(1000)
	s, _ = window.state(width, height)
	if expected := "Hello, _orld!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
}

func TestWindowIncrementDecrementEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	s, _ := window.state(width, height)
	if expected := 0; s.Size != expected {
		t.Errorf("s.Size should be %d but got %d", expected, s.Size)
	}
	if expected := int64(0); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}

	window.increment(0)
	s, _ = window.state(width, height)
	if expected := "\x01\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := 1; s.Size != expected {
		t.Errorf("s.Size should be %d but got %d", expected, s.Size)
	}
	if expected := int64(1); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}

	window, _ = newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.decrement(0)
	s, _ = window.state(width, height)
	if expected := "\xff\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := 1; s.Size != expected {
		t.Errorf("s.Size should be %d but got %d", expected, s.Size)
	}
	if expected := int64(1); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
}

func TestWindowInsertByte(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 1
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 7)
	window.startInsert()

	window.insertByte(mode.Insert, 0x04)
	s, _ := window.state(width, height)
	if expected := true; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}
	if expected := byte('\x40'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}

	window.insertByte(mode.Insert, 0x0a)
	s, _ = window.state(width, height)
	if expected := "Hello, Jworld!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := false; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}
	if expected := byte('\x00'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}
	if expected := int64(14); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}

	window.exitInsert()
	window.startAppendEnd()
	window.insertByte(mode.Insert, 0x04)
	window.insertByte(mode.Insert, 0x0b)
	window.insertByte(mode.Insert, 0x04)
	window.insertByte(mode.Insert, 0x0c)
	window.insertByte(mode.Insert, 0x04)
	window.insertByte(mode.Insert, 0x0d)
	s, _ = window.state(width, height)
	if expected := "M\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := false; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}
	if expected := byte('\x00'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}
	if expected := int64(18); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(16); s.Offset != expected {
		t.Errorf("s.Offset should be %d but got %d", expected, s.Offset)
	}
}

func TestWindowInsertEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startInsert()
	window.insertByte(mode.Insert, 0x04)
	window.insertByte(mode.Insert, 0x0a)
	s, _ := window.state(width, height)
	if expected := "J\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := false; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}
	if expected := byte('\x00'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}
	if expected := int64(2); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}

	window.exitInsert()
	s, _ = window.state(width, height)
	if expected := "J\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(1); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowInsertHead(t *testing.T) {
	r := strings.NewReader(strings.Repeat("Hello, world!", 2))
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.pageEnd()
	window.startInsertHead()
	s, _ := window.state(width, height)
	if expected := int64(16); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0a)
	s, _ = window.state(width, height)
	if expected := "Hello, world!Hel:lo, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := false; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}
	if expected := byte('\x00'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}
	if expected := int64(27); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(17); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowInsertHeadEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startInsertHead()
	s, _ := window.state(width, height)
	if expected := false; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}
	if expected := byte('\x00'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}
	if expected := int64(1); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.insertByte(mode.Insert, 0x04)
	window.insertByte(mode.Insert, 0x0a)
	window.exitInsert()
	s, _ = window.state(width, height)
	if expected := "J\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(1); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowAppend(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 7)
	window.startAppend()
	s, _ := window.state(width, height)
	if expected := int64(8); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0a)
	window.exitInsert()
	s, _ = window.state(width, height)
	if expected := "Hello, w:orld!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(14); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(8); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.cursorNext(mode.Normal, 10)
	window.startAppend()
	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0A)
	window.exitInsert()
	s, _ = window.state(width, height)
	if expected := "Hello, w:orld!:\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(15); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(14); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowAppendEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startAppend()
	window.exitInsert()
	s, _ := window.state(width, height)
	if expected := int64(0); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.startAppend()
	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0a)
	window.exitInsert()
	s, _ = window.state(width, height)
	if expected := ":\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(1); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.startAppendEnd()
	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0b)
	window.exitInsert()
	s, _ = window.state(width, height)
	if expected := ":;\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(2); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(1); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowReplaceByte(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 7)
	window.startReplaceByte()
	s, _ := window.state(width, height)
	if expected := int64(7); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0a)
	s, _ = window.state(width, height)
	if expected := "Hello, :orld!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(13); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(7); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowReplaceByteEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startReplaceByte()
	s, _ := window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0a)
	s, _ = window.state(width, height)
	if expected := ":\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(1); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowReplace(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 10)
	window.startReplace()
	s, _ := window.state(width, height)
	if expected := int64(10); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0a)
	s, _ = window.state(width, height)
	if expected := "Hello, wor:d!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(13); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(11); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0b)
	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0c)
	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0d)
	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0e)
	window.exitInsert()
	s, _ = window.state(width, height)
	if expected := "Hello, wor:;<=>\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(15); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(14); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowReplaceEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startReplace()
	s, _ := window.state(width, height)
	if expected := int64(0); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0a)
	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0b)
	window.exitInsert()
	s, _ = window.state(width, height)
	if expected := ":;\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(2); s.Length != expected {
		t.Errorf("s.Length should be %d but got %d", expected, s.Length)
	}
	if expected := int64(1); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowInsertByte2(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startInsert()
	window.insertByte(mode.Insert, 0x00)
	window.insertByte(mode.Insert, 0x01)
	window.insertByte(mode.Insert, 0x02)
	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x04)
	window.insertByte(mode.Insert, 0x05)
	window.insertByte(mode.Insert, 0x06)
	window.insertByte(mode.Insert, 0x07)
	window.insertByte(mode.Insert, 0x08)
	window.insertByte(mode.Insert, 0x09)
	window.insertByte(mode.Insert, 0x0a)
	window.insertByte(mode.Insert, 0x0b)
	window.insertByte(mode.Insert, 0x0c)
	window.insertByte(mode.Insert, 0x0d)
	window.insertByte(mode.Insert, 0x0e)
	window.insertByte(mode.Insert, 0x0f)
	window.exitInsert()
	s, _ := window.state(width, height)
	if expected := "\x01\x23\x45\x67\x89\xab\xcd\xef\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
}

func TestWindowBackspace(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 5)
	window.startInsert()
	window.backspace(mode.Insert)
	s, _ := window.state(width, height)
	if expected := "Hell, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	window.backspace(mode.Insert)
	window.backspace(mode.Insert)
	window.backspace(mode.Insert)
	window.backspace(mode.Insert)
	window.backspace(mode.Insert)
	s, _ = window.state(width, height)
	if expected := ", world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
}

func TestWindowBackspacePending(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 5)
	window.startInsert()
	window.insertByte(mode.Insert, 0x03)
	s, _ := window.state(width, height)
	if expected := true; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}
	if expected := byte('\x30'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}

	window.backspace(mode.Insert)
	s, _ = window.state(width, height)
	if expected := "Hello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := false; s.Pending != expected {
		t.Errorf("s.Pending should be %v but got %v", expected, s.Pending)
	}
	if expected := byte('\x00'); s.PendingByte != expected {
		t.Errorf("s.PendingByte should be %q but got %q", expected, s.PendingByte)
	}
}

func TestWindowEventRune(t *testing.T) {
	width, height := 16, 10
	redrawCh := make(chan struct{})
	window, _ := newWindow(strings.NewReader(""), "test", "test", make(chan event.Event), redrawCh)
	window.setSize(width, height)

	str := "48723fffab"
	go func() {
		window.emit(event.Event{Type: event.StartInsert})
		for _, r := range str {
			window.emit(event.Event{Type: event.Rune, Rune: r, Mode: mode.Insert})
		}
	}()
	<-redrawCh
	for range str {
		<-redrawCh
	}
	s, _ := window.state(width, height)
	if expected := "\x48\x72\x3f\xff\xab\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
}

func TestWindowEventRuneText(t *testing.T) {
	width, height := 16, 10
	redrawCh := make(chan struct{})
	window, _ := newWindow(strings.NewReader(""), "test", "test", make(chan event.Event), redrawCh)
	window.setSize(width, height)

	str := "Hello, World!\nこんにちは、世界！\n鰰は魚の一種"
	go func() {
		window.emit(event.Event{Type: event.SwitchFocus})
		window.emit(event.Event{Type: event.StartInsert})
		for _, r := range str {
			window.emit(event.Event{Type: event.Rune, Rune: r, Mode: mode.Insert})
		}
	}()
	<-redrawCh
	<-redrawCh
	for range str {
		<-redrawCh
	}
	s, _ := window.state(width, height)
	if expected := str + "\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
}

func TestWindowEventUndoRedo(t *testing.T) {
	width, height := 16, 10
	redrawCh := make(chan struct{})
	window, _ := newWindow(strings.NewReader("Hello, world!"), "test", "test", make(chan event.Event), redrawCh)
	window.setSize(width, height)
	waitCh := make(chan struct{})
	defer func() {
		close(waitCh)
		close(redrawCh)
	}()

	waitRedraw := func(count int) {
		for i := 0; i < count; i++ {
			<-redrawCh
		}
	}
	go func() {
		window.emit(event.Event{Type: event.Undo})
		window.emit(event.Event{Type: event.SwitchFocus})
		window.emit(event.Event{Type: event.StartAppend, Mode: mode.Insert})

		<-waitCh
		window.emit(event.Event{Type: event.Rune, Rune: 'x', Mode: mode.Insert})
		window.emit(event.Event{Type: event.Rune, Rune: 'y', Mode: mode.Insert})
		window.emit(event.Event{Type: event.Rune, Rune: 'z', Mode: mode.Insert})
		window.emit(event.Event{Type: event.ExitInsert})

		<-waitCh
		window.emit(event.Event{Type: event.StartInsert, Mode: mode.Insert})
		window.emit(event.Event{Type: event.Rune, Rune: 'x', Mode: mode.Insert})
		window.emit(event.Event{Type: event.Rune, Rune: 'y', Mode: mode.Insert})
		window.emit(event.Event{Type: event.CursorLeft, Mode: mode.Insert})
		window.emit(event.Event{Type: event.Rune, Rune: 'z', Mode: mode.Insert})
		window.emit(event.Event{Type: event.ExitInsert})

		<-waitCh
		window.emit(event.Event{Type: event.Undo, Count: 2})
		window.emit(event.Event{Type: event.StartInsert, Mode: mode.Insert})
		window.emit(event.Event{Type: event.Rune, Rune: 'w', Mode: mode.Insert})

		<-waitCh
		window.emit(event.Event{Type: event.ExitInsert})
		window.emit(event.Event{Type: event.Undo})

		<-waitCh
		window.emit(event.Event{Type: event.Redo, Count: 2})
	}()

	waitRedraw(3)
	s, _ := window.state(width, height)
	if expected := "Hello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(1); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(4)
	s, _ = window.state(width, height)
	if expected := "Hxyzello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(3); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(6)
	s, _ = window.state(width, height)
	if expected := "Hxyxzyzello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(5); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(3)
	s, _ = window.state(width, height)
	if expected := "Hxywzello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(4); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(2)
	s, _ = window.state(width, height)
	if expected := "Hxyzello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(3); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(1)
	s, _ = window.state(width, height)
	if expected := "Hxywzello, world!\x00"; !strings.HasPrefix(string(s.Bytes), expected) {
		t.Errorf("s.Bytes should start with %q but got %q", expected, string(s.Bytes))
	}
	if expected := int64(4); s.Cursor != expected {
		t.Errorf("s.Cursor should be %d but got %d", expected, s.Cursor)
	}
}

func TestWindowWriteTo(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	window, err := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(20, 10)
	window.cursorNext(mode.Normal, 3)
	window.startVisual()
	window.cursorNext(mode.Normal, 7)
	for _, testCase := range []struct {
		r        *event.Range
		expected string
	}{
		{nil, "Hello, world!"},
		{&event.Range{From: event.VisualStart{}, To: event.VisualEnd{}}, "lo, worl"},
	} {
		b := new(bytes.Buffer)
		n, err := window.writeTo(testCase.r, b)
		if expected := int64(len(testCase.expected)); n != expected {
			t.Errorf("writeTo should return %d but got: %d", expected, n)
		}
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if b.String() != testCase.expected {
			t.Errorf("window should write %q with range %+v but got %q", testCase.expected, testCase.r, b.String())
		}
	}
}
