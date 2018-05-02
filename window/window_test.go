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

	if s.Name != "test" {
		t.Errorf("state.Name should be %q but got %q", "test", s.Name)
	}

	if s.Width != width {
		t.Errorf("state.Width should be %d but got %d", width, s.Width)
	}

	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	if s.Size != 13 {
		t.Errorf("s.Size should be %d but got %d", 13, s.Size)
	}

	if s.Length != 13 {
		t.Errorf("s.Length should be %d but got %d", 13, s.Length)
	}

	if s.Pending != false {
		t.Errorf("s.Pending should be %v but got %v", false, s.Pending)
	}

	if s.PendingByte != '\x00' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x00', s.PendingByte)
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

	if s.Name != "test" {
		t.Errorf("state.Name should be %q but got %q", "test", s.Name)
	}

	if s.Width != width {
		t.Errorf("state.Width should be %d but got %d", width, s.Width)
	}

	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	if s.Size != 0 {
		t.Errorf("s.Size should be %d but got %d", 0, s.Size)
	}

	if s.Length != 0 {
		t.Errorf("s.Length should be %d but got %d", 0, s.Length)
	}

	if s.Pending != false {
		t.Errorf("s.Pending should be %v but got %v", false, s.Pending)
	}

	if s.PendingByte != '\x00' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x00', s.PendingByte)
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

	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
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
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	window.cursorDown(0)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width) {
		t.Errorf("s.Cursor should be %d but got %d", width, s.Cursor)
	}

	window.cursorDown(1)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width)*2 {
		t.Errorf("s.Cursor should be %d but got %d", width*2, s.Cursor)
	}

	window.cursorUp(0)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width) {
		t.Errorf("s.Cursor should be %d but got %d", width, s.Cursor)
	}

	window.cursorDown(10)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width)*11 {
		t.Errorf("s.Cursor should be %d but got %d", width*11, s.Cursor)
	}
	if s.Offset != int64(width)*2 {
		t.Errorf("s.Offset should be %d but got %d", width*2, s.Offset)
	}
	if !strings.HasPrefix(string(s.Bytes), " world!") {
		t.Errorf("s.Bytes should start with %q but got %q", " world!", string(s.Bytes))
	}

	window.cursorRight(mode.Normal, 3)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width)*11+3 {
		t.Errorf("s.Cursor should be %d but got %d", width*11+3, s.Cursor)
	}

	window.cursorRight(mode.Normal, 20)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width)*12-1 {
		t.Errorf("s.Cursor should be %d but got %d", width*12-1, s.Cursor)
	}

	window.cursorLeft(3)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width)*12-4 {
		t.Errorf("s.Cursor should be %d but got %d", width*12-4, s.Cursor)
	}

	window.cursorLeft(20)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width)*11 {
		t.Errorf("s.Cursor should be %d but got %d", width*11, s.Cursor)
	}

	window.cursorPrev(154)
	s, _ = window.state(width, height)
	if s.Cursor != 22 {
		t.Errorf("s.Cursor should be %d but got %d", 22, s.Cursor)
	}
	if s.Offset != int64(width) {
		t.Errorf("s.Offset should be %d but got %d", width, s.Offset)
	}

	window.cursorNext(mode.Normal, 200)
	s, _ = window.state(width, height)
	if s.Cursor != 222 {
		t.Errorf("s.Cursor should be %d but got %d", 222, s.Cursor)
	}
	if s.Offset != int64(width)*4 {
		t.Errorf("s.Offset should be %d but got %d", width*4, s.Offset)
	}

	window.cursorNext(mode.Normal, 2000)
	s, _ = window.state(width, height)
	if s.Cursor != 1299 {
		t.Errorf("s.Cursor should be %d but got %d", 1299, s.Cursor)
	}
	if s.Offset != int64(width)*72 {
		t.Errorf("s.Offset should be %d but got %d", width*72, s.Offset)
	}

	window.cursorHead(1)
	s, _ = window.state(width, height)
	if s.Cursor != 1296 {
		t.Errorf("s.Cursor should be %d but got %d", 1296, s.Cursor)
	}
	if s.Offset != int64(width)*72 {
		t.Errorf("s.Offset should be %d but got %d", width*72, s.Offset)
	}

	window.cursorEnd(1)
	s, _ = window.state(width, height)
	if s.Cursor != 1299 {
		t.Errorf("s.Cursor should be %d but got %d", 1299, s.Cursor)
	}
	if s.Offset != int64(width)*72 {
		t.Errorf("s.Offset should be %d but got %d", width*72, s.Offset)
	}

	window.cursorUp(20)
	window.cursorEnd(1)
	s, _ = window.state(width, height)
	if s.Cursor != 991 {
		t.Errorf("s.Cursor should be %d but got %d", 991, s.Cursor)
	}
	if s.Offset != int64(width)*61 {
		t.Errorf("s.Offset should be %d but got %d", width*61, s.Offset)
	}

	window.cursorEnd(11)
	s, _ = window.state(width, height)
	if s.Cursor != 1151 {
		t.Errorf("s.Cursor should be %d but got %d", 1151, s.Cursor)
	}
	if s.Offset != int64(width)*62 {
		t.Errorf("s.Offset should be %d but got %d", width*62, s.Offset)
	}

	window.cursorDown(30)
	s, _ = window.state(width, height)
	if s.Cursor != 1299 {
		t.Errorf("s.Cursor should be %d but got %d", 1299, s.Cursor)
	}
	if s.Offset != int64(width)*72 {
		t.Errorf("s.Offset should be %d but got %d", width*72, s.Offset)
	}

	window.cursorPrev(2000)
	s, _ = window.state(width, height)
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.cursorDown(2000)
	s, _ = window.state(width, height)
	if s.Cursor != int64(width)*81 {
		t.Errorf("s.Cursor should be %d but got %d", width*81, s.Cursor)
	}
	if s.Offset != int64(width)*72 {
		t.Errorf("s.Offset should be %d but got %d", width*72, s.Offset)
	}

	window.cursorRight(mode.Normal, 1000)
	s, _ = window.state(width, height)
	if s.Cursor != 1299 {
		t.Errorf("s.Cursor should be %d but got %d", 1299, s.Cursor)
	}
	if s.Offset != int64(width)*72 {
		t.Errorf("s.Offset should be %d but got %d", width*72, s.Offset)
	}

	window.cursorUp(2000)
	s, _ = window.state(width, height)
	if s.Cursor != 3 {
		t.Errorf("s.Cursor should be %d but got %d", 3, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.windowTop(0)
	s, _ = window.state(width, height)
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.windowTop(7)
	s, _ = window.state(width, height)
	if s.Cursor != 96 {
		t.Errorf("s.Cursor should be %d but got %d", 96, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.windowTop(20)
	s, _ = window.state(width, height)
	if s.Cursor != 144 {
		t.Errorf("s.Cursor should be %d but got %d", 144, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.windowMiddle()
	s, _ = window.state(width, height)
	if s.Cursor != 64 {
		t.Errorf("s.Cursor should be %d but got %d", 64, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.windowBottom(0)
	s, _ = window.state(width, height)
	if s.Cursor != 144 {
		t.Errorf("s.Cursor should be %d but got %d", 144, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.windowBottom(7)
	s, _ = window.state(width, height)
	if s.Cursor != 48 {
		t.Errorf("s.Cursor should be %d but got %d", 48, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.windowBottom(20)
	s, _ = window.state(width, height)
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
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
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	window.pageDown()
	s, _ = window.state(width, height)
	if s.Cursor != 128 {
		t.Errorf("s.Cursor should be %d but got %d", 128, s.Cursor)
	}
	if s.Offset != 128 {
		t.Errorf("s.Offset should be %d but got %d", 128, s.Offset)
	}

	window.pageDownHalf()
	s, _ = window.state(width, height)
	if s.Cursor != 208 {
		t.Errorf("s.Cursor should be %d but got %d", 208, s.Cursor)
	}
	if s.Offset != 208 {
		t.Errorf("s.Offset should be %d but got %d", 208, s.Offset)
	}

	window.scrollDown(0)
	s, _ = window.state(width, height)
	if s.Cursor != 224 {
		t.Errorf("s.Cursor should be %d but got %d", 224, s.Cursor)
	}
	if s.Offset != 224 {
		t.Errorf("s.Offset should be %d but got %d", 224, s.Offset)
	}

	window.scrollUp(0)
	s, _ = window.state(width, height)
	if s.Cursor != 224 {
		t.Errorf("s.Cursor should be %d but got %d", 224, s.Cursor)
	}
	if s.Offset != 208 {
		t.Errorf("s.Offset should be %d but got %d", 208, s.Offset)
	}

	window.scrollDown(30)
	s, _ = window.state(width, height)
	if s.Cursor != 688 {
		t.Errorf("s.Cursor should be %d but got %d", 688, s.Cursor)
	}
	if s.Offset != 688 {
		t.Errorf("s.Offset should be %d but got %d", 688, s.Offset)
	}

	window.scrollUp(30)
	s, _ = window.state(width, height)
	if s.Cursor != 352 {
		t.Errorf("s.Cursor should be %d but got %d", 352, s.Cursor)
	}
	if s.Offset != 208 {
		t.Errorf("s.Offset should be %d but got %d", 208, s.Offset)
	}

	window.pageUpHalf()
	s, _ = window.state(width, height)
	if s.Cursor != 272 {
		t.Errorf("s.Cursor should be %d but got %d", 272, s.Cursor)
	}
	if s.Offset != 128 {
		t.Errorf("s.Offset should be %d but got %d", 128, s.Offset)
	}

	window.pageUp()
	s, _ = window.state(width, height)
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.pageEnd()
	s, _ = window.state(width, height)
	if s.Cursor != 1296 {
		t.Errorf("s.Cursor should be %d but got %d", 1296, s.Cursor)
	}
	if s.Offset != int64(width)*72 {
		t.Errorf("s.Offset should be %d but got %d", width*72, s.Offset)
	}

	window.pageTop()
	s, _ = window.state(width, height)
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}
	if s.Offset != 0 {
		t.Errorf("s.Offset should be %d but got %d", 0, s.Offset)
	}

	window.cursorNext(mode.Normal, 5)
	window.scrollTop(5)
	s, _ = window.state(width, height)
	if s.Cursor != 85 {
		t.Errorf("s.Cursor should be %d but got %d", 85, s.Cursor)
	}
	if s.Offset != 80 {
		t.Errorf("s.Offset should be %d but got %d", 80, s.Offset)
	}

	window.cursorDown(4)
	window.scrollTop(0)
	s, _ = window.state(width, height)
	if s.Cursor != 149 {
		t.Errorf("s.Cursor should be %d but got %d", 149, s.Cursor)
	}
	if s.Offset != 144 {
		t.Errorf("s.Offset should be %d but got %d", 144, s.Offset)
	}

	window.scrollTopHead(10)
	s, _ = window.state(width, height)
	if s.Cursor != 160 {
		t.Errorf("s.Cursor should be %d but got %d", 160, s.Cursor)
	}
	if s.Offset != 160 {
		t.Errorf("s.Offset should be %d but got %d", 160, s.Offset)
	}

	window.cursorNext(mode.Normal, 5)
	window.scrollMiddle(12)
	s, _ = window.state(width, height)
	if s.Cursor != 197 {
		t.Errorf("s.Cursor should be %d but got %d", 197, s.Cursor)
	}
	if s.Offset != 112 {
		t.Errorf("s.Offset should be %d but got %d", 112, s.Offset)
	}

	window.scrollMiddleHead(15)
	s, _ = window.state(width, height)
	if s.Cursor != 240 {
		t.Errorf("s.Cursor should be %d but got %d", 240, s.Cursor)
	}
	if s.Offset != 160 {
		t.Errorf("s.Offset should be %d but got %d", 160, s.Offset)
	}

	window.cursorNext(mode.Normal, 5)
	window.scrollBottom(12)
	s, _ = window.state(width, height)
	if s.Cursor != 197 {
		t.Errorf("s.Cursor should be %d but got %d", 197, s.Cursor)
	}
	if s.Offset != 48 {
		t.Errorf("s.Offset should be %d but got %d", 48, s.Offset)
	}

	window.cursorDown(8)
	window.scrollBottom(0)
	s, _ = window.state(width, height)
	if s.Cursor != 325 {
		t.Errorf("s.Cursor should be %d but got %d", 325, s.Cursor)
	}
	if s.Offset != 176 {
		t.Errorf("s.Offset should be %d but got %d", 176, s.Offset)
	}

	window.scrollBottomHead(10)
	s, _ = window.state(width, height)
	if s.Cursor != 160 {
		t.Errorf("s.Cursor should be %d but got %d", 160, s.Cursor)
	}
	if s.Offset != 16 {
		t.Errorf("s.Offset should be %d but got %d", 16, s.Offset)
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
	if !strings.HasPrefix(string(s.Bytes), "Hello, orld!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, orld!\x00", string(s.Bytes))
	}
	if s.Cursor != 7 {
		t.Errorf("s.Cursor should be %d but got %d", 7, s.Cursor)
	}

	window.deleteBytes(3)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, d!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, d!\x00", string(s.Bytes))
	}
	if s.Cursor != 7 {
		t.Errorf("s.Cursor should be %d but got %d", 7, s.Cursor)
	}

	window.deleteBytes(3)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, \x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, \x00", string(s.Bytes))
	}
	if s.Cursor != 6 {
		t.Errorf("s.Cursor should be %d but got %d", 6, s.Cursor)
	}

	window.deleteByte()
	window.deleteByte()
	window.deleteByte()
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hell\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hell\x00", string(s.Bytes))
	}
	if s.Cursor != 3 {
		t.Errorf("s.Cursor should be %d but got %d", 3, s.Cursor)
	}

	window.deleteBytes(0)
	window.deleteBytes(0)
	window.deleteBytes(0)
	window.deleteBytes(0)
	window.deleteBytes(0)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "\x00", string(s.Bytes))
	}
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}
	if s.Length != 0 {
		t.Errorf("s.Length should be %d but got %d", 0, s.Length)
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
	if !strings.HasPrefix(string(s.Bytes), "Hell, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, orld!\x00", string(s.Bytes))
	}
	if s.Cursor != 4 {
		t.Errorf("s.Cursor should be %d but got %d", 4, s.Cursor)
	}

	window.deletePrevBytes(3)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "H, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "H, world!\x00", string(s.Bytes))
	}
	if s.Cursor != 1 {
		t.Errorf("s.Cursor should be %d but got %d", 1, s.Cursor)
	}

	window.deletePrevBytes(3)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), ", world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", ", world!\x00", string(s.Bytes))
	}
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}
}

func TestWindowIncrementDecrement(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.increment(0)
	s, _ := window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Iello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Iello, world\x00!", string(s.Bytes))
	}

	window.increment(1000)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "1ello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "1ello, world!\x00", string(s.Bytes))
	}

	window.increment(math.MaxInt64)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "0ello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "0ello, world!\x00", string(s.Bytes))
	}

	window.decrement(0)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "/ello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "/ello, world!\x00", string(s.Bytes))
	}

	window.decrement(1000)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Gello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Gello, world!\x00", string(s.Bytes))
	}

	window.decrement(math.MaxInt64)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, world!\x00", string(s.Bytes))
	}

	window.cursorNext(mode.Normal, 7)
	window.increment(1000)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, _orld!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, _orld!\x00", string(s.Bytes))
	}
}

func TestWindowIncrementDecrementEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	s, _ := window.state(width, height)
	if s.Size != 0 {
		t.Errorf("s.Size should be %d but got %d", 0, s.Size)
	}
	if s.Length != 0 {
		t.Errorf("s.Length should be %d but got %d", 0, s.Length)
	}

	window.increment(0)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "\x01\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "\x01\x00", string(s.Bytes))
	}
	if s.Size != 1 {
		t.Errorf("s.Size should be %d but got %d", 1, s.Size)
	}
	if s.Length != 1 {
		t.Errorf("s.Length should be %d but got %d", 1, s.Length)
	}

	window, _ = newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.decrement(0)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "\xff\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "\xff\x00", string(s.Bytes))
	}
	if s.Size != 1 {
		t.Errorf("s.Size should be %d but got %d", 1, s.Size)
	}
	if s.Length != 1 {
		t.Errorf("s.Length should be %d but got %d", 1, s.Length)
	}
}

func TestWindowInsertByte(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 1
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(mode.Normal, 7)
	window.startInsert()
	s, _ := window.state(width, height)

	window.insertByte(mode.Insert, 0x04)
	s, _ = window.state(width, height)
	if s.Pending != true {
		t.Errorf("s.Pending should be %v but got %v", true, s.Pending)
	}
	if s.PendingByte != '\x40' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x40', s.PendingByte)
	}

	window.insertByte(mode.Insert, 0x0a)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, Jworld!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, Jworld!\x00", string(s.Bytes))
	}
	if s.Pending != false {
		t.Errorf("s.Pending should be %v but got %v", false, s.Pending)
	}
	if s.PendingByte != '\x00' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x00', s.PendingByte)
	}
	if s.Length != 14 {
		t.Errorf("s.Length should be %d but got %d", 14, s.Length)
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
	if !strings.HasPrefix(string(s.Bytes), "M\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "M\x00", string(s.Bytes))
	}
	if s.Pending != false {
		t.Errorf("s.Pending should be %v but got %v", false, s.Pending)
	}
	if s.PendingByte != '\x00' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x00', s.PendingByte)
	}
	if s.Length != 18 {
		t.Errorf("s.Length should be %d but got %d", 18, s.Length)
	}
	if s.Offset != 16 {
		t.Errorf("s.Offset should be %d but got %d", 16, s.Offset)
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
	if !strings.HasPrefix(string(s.Bytes), "J\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "J\x00", string(s.Bytes))
	}
	if s.Pending != false {
		t.Errorf("s.Pending should be %v but got %v", false, s.Pending)
	}
	if s.PendingByte != '\x00' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x00', s.PendingByte)
	}
	if s.Length != 2 {
		t.Errorf("s.Length should be %d but got %d", 1, s.Length)
	}

	window.exitInsert()
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "J\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "J\x00", string(s.Bytes))
	}
	if s.Length != 1 {
		t.Errorf("s.Length should be %d but got %d", 1, s.Length)
	}
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
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
	if s.Cursor != 16 {
		t.Errorf("s.Cursor should be %d but got %d", 16, s.Cursor)
	}

	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0a)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, world!Hel:lo, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, world!Hel:lo, world!\x00", string(s.Bytes))
	}
	if s.Pending != false {
		t.Errorf("s.Pending should be %v but got %v", false, s.Pending)
	}
	if s.PendingByte != '\x00' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x00', s.PendingByte)
	}
	if s.Length != 27 {
		t.Errorf("s.Length should be %d but got %d", 27, s.Length)
	}
	if s.Cursor != 17 {
		t.Errorf("s.Cursor should be %d but got %d", 17, s.Cursor)
	}
}

func TestWindowInsertHeadEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startInsertHead()
	s, _ := window.state(width, height)
	if s.Pending != false {
		t.Errorf("s.Pending should be %v but got %v", false, s.Pending)
	}
	if s.PendingByte != '\x00' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x00', s.PendingByte)
	}
	if s.Length != 1 {
		t.Errorf("s.Length should be %d but got %d", 1, s.Length)
	}
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	window.insertByte(mode.Insert, 0x04)
	window.insertByte(mode.Insert, 0x0a)
	window.exitInsert()
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "J\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "J\x00", string(s.Bytes))
	}
	if s.Length != 1 {
		t.Errorf("s.Length should be %d but got %d", 1, s.Length)
	}
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
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
	if s.Cursor != 8 {
		t.Errorf("s.Cursor should be %d but got %d", 8, s.Cursor)
	}

	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0a)
	window.exitInsert()
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, w:orld!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, w:orld!\x00", string(s.Bytes))
	}
	if s.Length != 14 {
		t.Errorf("s.Length should be %d but got %d", 14, s.Length)
	}
	if s.Cursor != 8 {
		t.Errorf("s.Cursor should be %d but got %d", 8, s.Cursor)
	}

	window.cursorNext(mode.Normal, 10)
	window.startAppend()
	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0A)
	window.exitInsert()
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, w:orld!:\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, w:orld!:\x00", string(s.Bytes))
	}
	if s.Length != 15 {
		t.Errorf("s.Length should be %d but got %d", 15, s.Length)
	}
	if s.Cursor != 14 {
		t.Errorf("s.Cursor should be %d but got %d", 14, s.Cursor)
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
	if s.Length != 0 {
		t.Errorf("s.Length should be %d but got %d", 0, s.Length)
	}
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	window.startAppend()
	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0a)
	window.exitInsert()
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), ":\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", ":\x00", string(s.Bytes))
	}
	if s.Length != 1 {
		t.Errorf("s.Length should be %d but got %d", 1, s.Length)
	}
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	window.startAppendEnd()
	window.insertByte(mode.Insert, 0x03)
	window.insertByte(mode.Insert, 0x0b)
	window.exitInsert()
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), ":;\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", ":;\x00", string(s.Bytes))
	}
	if s.Length != 2 {
		t.Errorf("s.Length should be %d but got %d", 2, s.Length)
	}
	if s.Cursor != 1 {
		t.Errorf("s.Cursor should be %d but got %d", 1, s.Cursor)
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
	if s.Cursor != 7 {
		t.Errorf("s.Cursor should be %d but got %d", 7, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0a)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, :orld!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, :orld!\x00", string(s.Bytes))
	}
	if s.Length != 13 {
		t.Errorf("s.Length should be %d but got %d", 13, s.Length)
	}
	if s.Cursor != 7 {
		t.Errorf("s.Cursor should be %d but got %d", 7, s.Cursor)
	}
}

func TestWindowReplaceByteEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startReplaceByte()
	s, _ := window.state(width, height)
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0a)
	s, _ = window.state(width, height)
	if strings.HasPrefix(string(s.Bytes), ":\x00") {
		t.Errorf("s.Bytes should not start with %q but got %q", ":\x00", string(s.Bytes))
	}
	if s.Length != 0 {
		t.Errorf("s.Length should be %d but got %d", 0, s.Length)
	}
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
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
	if s.Cursor != 10 {
		t.Errorf("s.Cursor should be %d but got %d", 10, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0a)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, wor:d!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, wor:d!\x00", string(s.Bytes))
	}
	if s.Length != 13 {
		t.Errorf("s.Length should be %d but got %d", 13, s.Length)
	}
	if s.Cursor != 11 {
		t.Errorf("s.Cursor should be %d but got %d", 11, s.Cursor)
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
	if !strings.HasPrefix(string(s.Bytes), "Hello, wor:;<=>\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, wor:;<=>\x00", string(s.Bytes))
	}
	if s.Length != 15 {
		t.Errorf("s.Length should be %d but got %d", 15, s.Length)
	}
	if s.Cursor != 14 {
		t.Errorf("s.Cursor should be %d but got %d", 14, s.Cursor)
	}
}

func TestWindowReplaceEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan event.Event), make(chan struct{}))
	window.setSize(width, height)

	window.startReplace()
	s, _ := window.state(width, height)
	if s.Cursor != 0 {
		t.Errorf("s.Cursor should be %d but got %d", 0, s.Cursor)
	}

	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0a)
	window.insertByte(mode.Replace, 0x03)
	window.insertByte(mode.Replace, 0x0b)
	window.exitInsert()
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), ":;\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", ":;\x00", string(s.Bytes))
	}
	if s.Length != 2 {
		t.Errorf("s.Length should be %d but got %d", 2, s.Length)
	}
	if s.Cursor != 1 {
		t.Errorf("s.Cursor should be %d but got %d", 1, s.Cursor)
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
	if !strings.HasPrefix(string(s.Bytes), "\x01\x23\x45\x67\x89\xab\xcd\xef\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "\x01\x23\x45\x67\x89\xab\xcd\xef\x00", string(s.Bytes))
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
	if !strings.HasPrefix(string(s.Bytes), "Hell, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hell, world!\x00", string(s.Bytes))
	}
	window.backspace(mode.Insert)
	window.backspace(mode.Insert)
	window.backspace(mode.Insert)
	window.backspace(mode.Insert)
	window.backspace(mode.Insert)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), ", world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", ", world!\x00", string(s.Bytes))
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
	if s.Pending != true {
		t.Errorf("s.Pending should be %v but got %v", true, s.Pending)
	}
	if s.PendingByte != '\x30' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x30', s.PendingByte)
	}

	window.backspace(mode.Insert)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, world!\x00", string(s.Bytes))
	}
	if s.Pending != false {
		t.Errorf("s.Pending should be %v but got %v", false, s.Pending)
	}
	if s.PendingByte != '\x00' {
		t.Errorf("s.PendingByte should be %q but got %q", '\x00', s.PendingByte)
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
	if !strings.HasPrefix(string(s.Bytes), "\x48\x72\x3f\xff\xab\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "\x48\x72\x3f\xff\xab\x00", string(s.Bytes))
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
	if !strings.HasPrefix(string(s.Bytes), str+"\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", str+"\x00", string(s.Bytes))
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
	if !strings.HasPrefix(string(s.Bytes), "Hello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hello, world!\x00", string(s.Bytes))
	}
	if s.Cursor != 1 {
		t.Errorf("s.Cursor should be %d but got %d", 1, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(4)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hxyzello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hxyzello, world!\x00", string(s.Bytes))
	}
	if s.Cursor != 3 {
		t.Errorf("s.Cursor should be %d but got %d", 3, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(6)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hxyxzyzello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hxyxzyzello, world!\x00", string(s.Bytes))
	}
	if s.Cursor != 5 {
		t.Errorf("s.Cursor should be %d but got %d", 5, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(3)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hxywzello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hxywzello, world!\x00", string(s.Bytes))
	}
	if s.Cursor != 4 {
		t.Errorf("s.Cursor should be %d but got %d", 4, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(2)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hxyzello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hxyzello, world!\x00", string(s.Bytes))
	}
	if s.Cursor != 3 {
		t.Errorf("s.Cursor should be %d but got %d", 3, s.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(1)
	s, _ = window.state(width, height)
	if !strings.HasPrefix(string(s.Bytes), "Hxywzello, world!\x00") {
		t.Errorf("s.Bytes should start with %q but got %q", "Hxywzello, world!\x00", string(s.Bytes))
	}
	if s.Cursor != 4 {
		t.Errorf("s.Cursor should be %d but got %d", 4, s.Cursor)
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
		if n != int64(len(testCase.expected)) {
			t.Errorf("writeTo should return %d but got: %d", int64(len(testCase.expected)), n)
		}
		if err != nil {
			t.Errorf("err should be nil but got: %v", err)
		}
		if b.String() != testCase.expected {
			t.Errorf("window should write %q with range %+v but got %q", testCase.expected, testCase.r, b.String())
		}
	}
}
