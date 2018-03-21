package window

import (
	"math"
	"reflect"
	"strings"
	"testing"

	. "github.com/itchyny/bed/common"
)

func TestWindowState(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, err := newWindow(r, "test", "test", make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(width, height)

	state, err := window.State()
	if err != nil {
		t.Fatal(err)
	}

	if state.Name != "test" {
		t.Errorf("state.Name should be %q but got %q", "test", state.Name)
	}

	if state.Width != width {
		t.Errorf("state.Width should be %d but got %d", width, state.Width)
	}

	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}

	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	if state.Size != 13 {
		t.Errorf("state.Size should be %d but got %d", 13, state.Size)
	}

	if state.Length != 13 {
		t.Errorf("state.Length should be %d but got %d", 13, state.Length)
	}

	if state.Pending != false {
		t.Errorf("state.Pending should be %v but got %v", false, state.Pending)
	}

	if state.PendingByte != '\x00' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x00', state.PendingByte)
	}

	if !reflect.DeepEqual(state.EditedIndices, []int64{}) {
		t.Errorf("state.EditedIndices should be empty but got %v", state.EditedIndices)
	}

	expected := []byte("Hello, world!" + strings.Repeat("\x00", height*width-13))
	if !reflect.DeepEqual(state.Bytes, expected) {
		t.Errorf("state.Bytes should be %q but got %q", expected, state.Bytes)
	}
}

func TestWindowEmptyState(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, err := newWindow(r, "test", "test", make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(width, height)

	state, err := window.State()
	if err != nil {
		t.Fatal(err)
	}

	if state.Name != "test" {
		t.Errorf("state.Name should be %q but got %q", "test", state.Name)
	}

	if state.Width != width {
		t.Errorf("state.Width should be %d but got %d", width, state.Width)
	}

	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}

	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	if state.Size != 0 {
		t.Errorf("state.Size should be %d but got %d", 0, state.Size)
	}

	if state.Length != 0 {
		t.Errorf("state.Length should be %d but got %d", 0, state.Length)
	}

	if state.Pending != false {
		t.Errorf("state.Pending should be %v but got %v", false, state.Pending)
	}

	if state.PendingByte != '\x00' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x00', state.PendingByte)
	}

	if !reflect.DeepEqual(state.EditedIndices, []int64{}) {
		t.Errorf("state.EditedIndices should be empty but got %v", state.EditedIndices)
	}

	expected := []byte(strings.Repeat("\x00", height*width))
	if !reflect.DeepEqual(state.Bytes, expected) {
		t.Errorf("state.Bytes should be %q but got %q", expected, state.Bytes)
	}

	window.scrollDown(0)
	state, err = window.State()
	if err != nil {
		t.Fatal(err)
	}

	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}
}

func TestWindowCursorMotions(t *testing.T) {
	r := strings.NewReader(strings.Repeat("Hello, world!", 100))
	width, height := 16, 10
	window, err := newWindow(r, "test", "test", make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(width, height)

	state, _ := window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	window.cursorDown(0)
	state, _ = window.State()
	if state.Cursor != int64(width) {
		t.Errorf("state.Cursor should be %d but got %d", width, state.Cursor)
	}

	window.cursorDown(1)
	state, _ = window.State()
	if state.Cursor != int64(width)*2 {
		t.Errorf("state.Cursor should be %d but got %d", width*2, state.Cursor)
	}

	window.cursorUp(0)
	state, _ = window.State()
	if state.Cursor != int64(width) {
		t.Errorf("state.Cursor should be %d but got %d", width, state.Cursor)
	}

	window.cursorDown(10)
	state, _ = window.State()
	if state.Cursor != int64(width)*11 {
		t.Errorf("state.Cursor should be %d but got %d", width*11, state.Cursor)
	}
	if state.Offset != int64(width)*2 {
		t.Errorf("state.Offset should be %d but got %d", width*2, state.Offset)
	}
	if !strings.HasPrefix(string(state.Bytes), " world!") {
		t.Errorf("state.Bytes should start with %q but got %q", " world!", string(state.Bytes))
	}

	window.cursorRight(ModeNormal, 3)
	state, _ = window.State()
	if state.Cursor != int64(width)*11+3 {
		t.Errorf("state.Cursor should be %d but got %d", width*11+3, state.Cursor)
	}

	window.cursorRight(ModeNormal, 20)
	state, _ = window.State()
	if state.Cursor != int64(width)*12-1 {
		t.Errorf("state.Cursor should be %d but got %d", width*12-1, state.Cursor)
	}

	window.cursorLeft(3)
	state, _ = window.State()
	if state.Cursor != int64(width)*12-4 {
		t.Errorf("state.Cursor should be %d but got %d", width*12-4, state.Cursor)
	}

	window.cursorLeft(20)
	state, _ = window.State()
	if state.Cursor != int64(width)*11 {
		t.Errorf("state.Cursor should be %d but got %d", width*11, state.Cursor)
	}

	window.cursorPrev(154)
	state, _ = window.State()
	if state.Cursor != 22 {
		t.Errorf("state.Cursor should be %d but got %d", 22, state.Cursor)
	}
	if state.Offset != int64(width) {
		t.Errorf("state.Offset should be %d but got %d", width, state.Offset)
	}

	window.cursorNext(ModeNormal, 200)
	state, _ = window.State()
	if state.Cursor != 222 {
		t.Errorf("state.Cursor should be %d but got %d", 222, state.Cursor)
	}
	if state.Offset != int64(width)*4 {
		t.Errorf("state.Offset should be %d but got %d", width*4, state.Offset)
	}

	window.cursorNext(ModeNormal, 2000)
	state, _ = window.State()
	if state.Cursor != 1299 {
		t.Errorf("state.Cursor should be %d but got %d", 1299, state.Cursor)
	}
	if state.Offset != int64(width)*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorHead(1)
	state, _ = window.State()
	if state.Cursor != 1296 {
		t.Errorf("state.Cursor should be %d but got %d", 1296, state.Cursor)
	}
	if state.Offset != int64(width)*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorEnd(1)
	state, _ = window.State()
	if state.Cursor != 1299 {
		t.Errorf("state.Cursor should be %d but got %d", 1299, state.Cursor)
	}
	if state.Offset != int64(width)*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorUp(20)
	window.cursorEnd(1)
	state, _ = window.State()
	if state.Cursor != 991 {
		t.Errorf("state.Cursor should be %d but got %d", 991, state.Cursor)
	}
	if state.Offset != int64(width)*61 {
		t.Errorf("state.Offset should be %d but got %d", width*61, state.Offset)
	}

	window.cursorEnd(11)
	state, _ = window.State()
	if state.Cursor != 1151 {
		t.Errorf("state.Cursor should be %d but got %d", 1151, state.Cursor)
	}
	if state.Offset != int64(width)*62 {
		t.Errorf("state.Offset should be %d but got %d", width*62, state.Offset)
	}

	window.cursorDown(30)
	state, _ = window.State()
	if state.Cursor != 1299 {
		t.Errorf("state.Cursor should be %d but got %d", 1299, state.Cursor)
	}
	if state.Offset != int64(width)*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorPrev(2000)
	state, _ = window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}
	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}

	window.cursorDown(2000)
	state, _ = window.State()
	if state.Cursor != int64(width)*81 {
		t.Errorf("state.Cursor should be %d but got %d", width*81, state.Cursor)
	}
	if state.Offset != int64(width)*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorRight(ModeNormal, 1000)
	state, _ = window.State()
	if state.Cursor != 1299 {
		t.Errorf("state.Cursor should be %d but got %d", 1299, state.Cursor)
	}
	if state.Offset != int64(width)*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.cursorUp(2000)
	state, _ = window.State()
	if state.Cursor != 3 {
		t.Errorf("state.Cursor should be %d but got %d", 3, state.Cursor)
	}
	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}
}

func TestParseGotoPos(t *testing.T) {
	for _, p := range []struct {
		pos   string
		count int64
	}{
		{"0", 0},
		{"10", 16},
		{"8000000000", 549755813888},
		{"123456789abcdefe", 1311768467463790334},
		{"-0", 0},
		{"-", -1},
		{"+", 1},
		{"+10", 16},
		{"-10", -16},
		{"+8000000000", 549755813888},
		{"-8000000000", -549755813888},
		{"-123456789abcdefe", -1311768467463790334},
		{"$", math.MaxInt64},
	} {
		if c := parseGotoPos(p.pos); c != p.count {
			t.Errorf("count should be %v but got: %v (pos: %q)", p.count, c, p.pos)
		}
	}
}

func TestWindowScreenMotions(t *testing.T) {
	r := strings.NewReader(strings.Repeat("Hello, world!", 100))
	width, height := 16, 10
	window, err := newWindow(r, "test", "test", make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}
	window.setSize(width, height)

	state, _ := window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	window.pageDown()
	state, _ = window.State()
	if state.Cursor != 128 {
		t.Errorf("state.Cursor should be %d but got %d", 128, state.Cursor)
	}
	if state.Offset != 128 {
		t.Errorf("state.Offset should be %d but got %d", 128, state.Offset)
	}

	window.pageDownHalf()
	state, _ = window.State()
	if state.Cursor != 208 {
		t.Errorf("state.Cursor should be %d but got %d", 208, state.Cursor)
	}
	if state.Offset != 208 {
		t.Errorf("state.Offset should be %d but got %d", 208, state.Offset)
	}

	window.scrollDown(0)
	state, _ = window.State()
	if state.Cursor != 224 {
		t.Errorf("state.Cursor should be %d but got %d", 224, state.Cursor)
	}
	if state.Offset != 224 {
		t.Errorf("state.Offset should be %d but got %d", 224, state.Offset)
	}

	window.scrollUp(0)
	state, _ = window.State()
	if state.Cursor != 224 {
		t.Errorf("state.Cursor should be %d but got %d", 224, state.Cursor)
	}
	if state.Offset != 208 {
		t.Errorf("state.Offset should be %d but got %d", 208, state.Offset)
	}

	window.scrollDown(30)
	state, _ = window.State()
	if state.Cursor != 688 {
		t.Errorf("state.Cursor should be %d but got %d", 688, state.Cursor)
	}
	if state.Offset != 688 {
		t.Errorf("state.Offset should be %d but got %d", 688, state.Offset)
	}

	window.scrollUp(30)
	state, _ = window.State()
	if state.Cursor != 352 {
		t.Errorf("state.Cursor should be %d but got %d", 352, state.Cursor)
	}
	if state.Offset != 208 {
		t.Errorf("state.Offset should be %d but got %d", 208, state.Offset)
	}

	window.pageUpHalf()
	state, _ = window.State()
	if state.Cursor != 272 {
		t.Errorf("state.Cursor should be %d but got %d", 272, state.Cursor)
	}
	if state.Offset != 128 {
		t.Errorf("state.Offset should be %d but got %d", 128, state.Offset)
	}

	window.pageUp()
	state, _ = window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}
	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}

	window.pageEnd()
	state, _ = window.State()
	if state.Cursor != 1296 {
		t.Errorf("state.Cursor should be %d but got %d", 1296, state.Cursor)
	}
	if state.Offset != int64(width)*72 {
		t.Errorf("state.Offset should be %d but got %d", width*72, state.Offset)
	}

	window.pageTop()
	state, _ = window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}
	if state.Offset != 0 {
		t.Errorf("state.Offset should be %d but got %d", 0, state.Offset)
	}
}

func TestWindowDeleteBytes(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(ModeNormal, 7)
	window.deleteByte(0)
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, orld!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, orld!\x00", string(state.Bytes))
	}
	if state.Cursor != 7 {
		t.Errorf("state.Cursor should be %d but got %d", 7, state.Cursor)
	}

	window.deleteByte(3)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, d!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, d!\x00", string(state.Bytes))
	}
	if state.Cursor != 7 {
		t.Errorf("state.Cursor should be %d but got %d", 7, state.Cursor)
	}

	window.deleteByte(3)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, \x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, \x00", string(state.Bytes))
	}
	if state.Cursor != 6 {
		t.Errorf("state.Cursor should be %d but got %d", 6, state.Cursor)
	}

	window.deleteByte(0)
	window.deleteByte(0)
	window.deleteByte(0)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hell\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hell\x00", string(state.Bytes))
	}
	if state.Cursor != 3 {
		t.Errorf("state.Cursor should be %d but got %d", 3, state.Cursor)
	}

	window.deleteByte(0)
	window.deleteByte(0)
	window.deleteByte(0)
	window.deleteByte(0)
	window.deleteByte(0)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "\x00", string(state.Bytes))
	}
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}
	if state.Length != 0 {
		t.Errorf("state.Length should be %d but got %d", 0, state.Length)
	}
}

func TestWindowDeletePrevBytes(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(ModeNormal, 5)
	window.deletePrevByte(0)
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hell, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, orld!\x00", string(state.Bytes))
	}
	if state.Cursor != 4 {
		t.Errorf("state.Cursor should be %d but got %d", 4, state.Cursor)
	}

	window.deletePrevByte(3)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "H, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "H, world!\x00", string(state.Bytes))
	}
	if state.Cursor != 1 {
		t.Errorf("state.Cursor should be %d but got %d", 1, state.Cursor)
	}

	window.deletePrevByte(3)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), ", world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", ", world!\x00", string(state.Bytes))
	}
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}
}

func TestWindowIncrementDecrement(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.increment(0)
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), "Iello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Iello, world\x00!", string(state.Bytes))
	}

	window.increment(1000)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "1ello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "1ello, world!\x00", string(state.Bytes))
	}

	window.increment(math.MaxInt64)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "0ello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "0ello, world!\x00", string(state.Bytes))
	}

	window.decrement(0)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "/ello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "/ello, world!\x00", string(state.Bytes))
	}

	window.decrement(1000)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Gello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Gello, world!\x00", string(state.Bytes))
	}

	window.decrement(math.MaxInt64)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, world!\x00", string(state.Bytes))
	}

	window.cursorNext(ModeNormal, 7)
	window.increment(1000)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, _orld!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, _orld!\x00", string(state.Bytes))
	}
}

func TestWindowIncrementDecrementEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	state, _ := window.State()
	if state.Size != 0 {
		t.Errorf("state.Size should be %d but got %d", 0, state.Size)
	}
	if state.Length != 0 {
		t.Errorf("state.Length should be %d but got %d", 0, state.Length)
	}

	window.increment(0)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "\x01\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "\x01\x00", string(state.Bytes))
	}
	if state.Size != 1 {
		t.Errorf("state.Size should be %d but got %d", 1, state.Size)
	}
	if state.Length != 1 {
		t.Errorf("state.Length should be %d but got %d", 1, state.Length)
	}

	window, _ = newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.decrement(0)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "\xff\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "\xff\x00", string(state.Bytes))
	}
	if state.Size != 1 {
		t.Errorf("state.Size should be %d but got %d", 1, state.Size)
	}
	if state.Length != 1 {
		t.Errorf("state.Length should be %d but got %d", 1, state.Length)
	}
}

func TestWindowInsertByte(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(ModeNormal, 7)
	window.startInsert()
	state, _ := window.State()

	window.insertByte(ModeInsert, 0x04)
	state, _ = window.State()
	if state.Pending != true {
		t.Errorf("state.Pending should be %v but got %v", true, state.Pending)
	}
	if state.PendingByte != '\x40' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x40', state.PendingByte)
	}

	window.insertByte(ModeInsert, 0x0a)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, Jworld!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, Jworld!\x00", string(state.Bytes))
	}
	if state.Pending != false {
		t.Errorf("state.Pending should be %v but got %v", false, state.Pending)
	}
	if state.PendingByte != '\x00' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x00', state.PendingByte)
	}
	if state.Length != 14 {
		t.Errorf("state.Length should be %d but got %d", 14, state.Length)
	}
}

func TestWindowInsertEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.startInsert()
	window.insertByte(ModeInsert, 0x04)
	window.insertByte(ModeInsert, 0x0a)
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), "J\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "J\x00", string(state.Bytes))
	}
	if state.Pending != false {
		t.Errorf("state.Pending should be %v but got %v", false, state.Pending)
	}
	if state.PendingByte != '\x00' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x00', state.PendingByte)
	}
	if state.Length != 2 {
		t.Errorf("state.Length should be %d but got %d", 1, state.Length)
	}

	window.exitInsert()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "J\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "J\x00", string(state.Bytes))
	}
	if state.Length != 1 {
		t.Errorf("state.Length should be %d but got %d", 1, state.Length)
	}
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}
}

func TestWindowInsertHead(t *testing.T) {
	r := strings.NewReader(strings.Repeat("Hello, world!", 2))
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.pageEnd()
	window.startInsertHead()
	state, _ := window.State()
	if state.Cursor != 16 {
		t.Errorf("state.Cursor should be %d but got %d", 16, state.Cursor)
	}

	window.insertByte(ModeInsert, 0x03)
	window.insertByte(ModeInsert, 0x0a)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, world!Hel:lo, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, world!Hel:lo, world!\x00", string(state.Bytes))
	}
	if state.Pending != false {
		t.Errorf("state.Pending should be %v but got %v", false, state.Pending)
	}
	if state.PendingByte != '\x00' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x00', state.PendingByte)
	}
	if state.Length != 27 {
		t.Errorf("state.Length should be %d but got %d", 27, state.Length)
	}
	if state.Cursor != 17 {
		t.Errorf("state.Cursor should be %d but got %d", 17, state.Cursor)
	}
}

func TestWindowAppend(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(ModeNormal, 7)
	window.startAppend()
	state, _ := window.State()
	if state.Cursor != 8 {
		t.Errorf("state.Cursor should be %d but got %d", 8, state.Cursor)
	}

	window.insertByte(ModeInsert, 0x03)
	window.insertByte(ModeInsert, 0x0a)
	window.exitInsert()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, w:orld!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, w:orld!\x00", string(state.Bytes))
	}
	if state.Length != 14 {
		t.Errorf("state.Length should be %d but got %d", 14, state.Length)
	}
	if state.Cursor != 8 {
		t.Errorf("state.Cursor should be %d but got %d", 8, state.Cursor)
	}

	window.cursorNext(ModeNormal, 10)
	window.startAppend()
	window.insertByte(ModeInsert, 0x03)
	window.insertByte(ModeInsert, 0x0A)
	window.exitInsert()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, w:orld!:\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, w:orld!:\x00", string(state.Bytes))
	}
	if state.Length != 15 {
		t.Errorf("state.Length should be %d but got %d", 15, state.Length)
	}
	if state.Cursor != 14 {
		t.Errorf("state.Cursor should be %d but got %d", 14, state.Cursor)
	}
}

func TestWindowAppendEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.startAppend()
	window.exitInsert()
	state, _ := window.State()
	if state.Length != 0 {
		t.Errorf("state.Length should be %d but got %d", 0, state.Length)
	}
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	window.startAppend()
	window.insertByte(ModeInsert, 0x03)
	window.insertByte(ModeInsert, 0x0a)
	window.exitInsert()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), ":\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", ":\x00", string(state.Bytes))
	}
	if state.Length != 1 {
		t.Errorf("state.Length should be %d but got %d", 1, state.Length)
	}
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	window.startAppendEnd()
	window.insertByte(ModeInsert, 0x03)
	window.insertByte(ModeInsert, 0x0b)
	window.exitInsert()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), ":;\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", ":;\x00", string(state.Bytes))
	}
	if state.Length != 2 {
		t.Errorf("state.Length should be %d but got %d", 2, state.Length)
	}
	if state.Cursor != 1 {
		t.Errorf("state.Cursor should be %d but got %d", 1, state.Cursor)
	}
}

func TestWindowReplaceByte(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(ModeNormal, 7)
	window.startReplaceByte()
	state, _ := window.State()
	if state.Cursor != 7 {
		t.Errorf("state.Cursor should be %d but got %d", 7, state.Cursor)
	}

	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0a)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, :orld!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, :orld!\x00", string(state.Bytes))
	}
	if state.Length != 13 {
		t.Errorf("state.Length should be %d but got %d", 13, state.Length)
	}
	if state.Cursor != 7 {
		t.Errorf("state.Cursor should be %d but got %d", 7, state.Cursor)
	}
}

func TestWindowReplaceByteEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.startReplaceByte()
	state, _ := window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0a)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), ":\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", ":\x00", string(state.Bytes))
	}
	if state.Length != 1 {
		t.Errorf("state.Length should be %d but got %d", 1, state.Length)
	}
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}
}

func TestWindowReplace(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(ModeNormal, 10)
	window.startReplace()
	state, _ := window.State()
	if state.Cursor != 10 {
		t.Errorf("state.Cursor should be %d but got %d", 10, state.Cursor)
	}

	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0a)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, wor:d!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, wor:d!\x00", string(state.Bytes))
	}
	if state.Length != 13 {
		t.Errorf("state.Length should be %d but got %d", 13, state.Length)
	}
	if state.Cursor != 11 {
		t.Errorf("state.Cursor should be %d but got %d", 11, state.Cursor)
	}

	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0b)
	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0c)
	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0d)
	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0e)
	window.exitInsert()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, wor:;<=>\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, wor:;<=>\x00", string(state.Bytes))
	}
	if state.Length != 15 {
		t.Errorf("state.Length should be %d but got %d", 15, state.Length)
	}
	if state.Cursor != 14 {
		t.Errorf("state.Cursor should be %d but got %d", 14, state.Cursor)
	}
}

func TestWindowReplaceEmpty(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.startReplace()
	state, _ := window.State()
	if state.Cursor != 0 {
		t.Errorf("state.Cursor should be %d but got %d", 0, state.Cursor)
	}

	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0a)
	window.insertByte(ModeReplace, 0x03)
	window.insertByte(ModeReplace, 0x0b)
	window.exitInsert()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), ":;\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", ":;\x00", string(state.Bytes))
	}
	if state.Length != 2 {
		t.Errorf("state.Length should be %d but got %d", 2, state.Length)
	}
	if state.Cursor != 1 {
		t.Errorf("state.Cursor should be %d but got %d", 1, state.Cursor)
	}
}

func TestWindowInsertByte2(t *testing.T) {
	r := strings.NewReader("")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.startInsert()
	window.insertByte(ModeInsert, 0x00)
	window.insertByte(ModeInsert, 0x01)
	window.insertByte(ModeInsert, 0x02)
	window.insertByte(ModeInsert, 0x03)
	window.insertByte(ModeInsert, 0x04)
	window.insertByte(ModeInsert, 0x05)
	window.insertByte(ModeInsert, 0x06)
	window.insertByte(ModeInsert, 0x07)
	window.insertByte(ModeInsert, 0x08)
	window.insertByte(ModeInsert, 0x09)
	window.insertByte(ModeInsert, 0x0a)
	window.insertByte(ModeInsert, 0x0b)
	window.insertByte(ModeInsert, 0x0c)
	window.insertByte(ModeInsert, 0x0d)
	window.insertByte(ModeInsert, 0x0e)
	window.insertByte(ModeInsert, 0x0f)
	window.exitInsert()
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), "\x01\x23\x45\x67\x89\xab\xcd\xef\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "\x01\x23\x45\x67\x89\xab\xcd\xef\x00", string(state.Bytes))
	}
}

func TestWindowBackspace(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(ModeNormal, 5)
	window.startInsert()
	window.backspace()
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hell, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hell, world!\x00", string(state.Bytes))
	}
	window.backspace()
	window.backspace()
	window.backspace()
	window.backspace()
	window.backspace()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), ", world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", ", world!\x00", string(state.Bytes))
	}
}

func TestWindowBackspacePending(t *testing.T) {
	r := strings.NewReader("Hello, world!")
	width, height := 16, 10
	window, _ := newWindow(r, "test", "test", make(chan struct{}))
	window.setSize(width, height)

	window.cursorNext(ModeNormal, 5)
	window.startInsert()
	window.insertByte(ModeInsert, 0x03)
	state, _ := window.State()
	if state.Pending != true {
		t.Errorf("state.Pending should be %v but got %v", true, state.Pending)
	}
	if state.PendingByte != '\x30' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x30', state.PendingByte)
	}

	window.backspace()
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, world!\x00", string(state.Bytes))
	}
	if state.Pending != false {
		t.Errorf("state.Pending should be %v but got %v", false, state.Pending)
	}
	if state.PendingByte != '\x00' {
		t.Errorf("state.PendingByte should be %q but got %q", '\x00', state.PendingByte)
	}
}

func TestWindowEventRune(t *testing.T) {
	width, height := 16, 10
	redrawCh := make(chan struct{})
	window, _ := newWindow(strings.NewReader(""), "test", "test", redrawCh)
	window.setSize(width, height)

	str := "48723fffab"
	go func() {
		window.Run()
	}()
	go func() {
		window.eventCh <- Event{Type: EventStartInsert}
		for _, r := range str {
			window.eventCh <- Event{Type: EventRune, Rune: r, Mode: ModeInsert}
		}
	}()
	<-redrawCh
	for _ = range str {
		<-redrawCh
	}
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), "\x48\x72\x3f\xff\xab\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "\x48\x72\x3f\xff\xab\x00", string(state.Bytes))
	}
	window.Close()
}

func TestWindowEventRuneText(t *testing.T) {
	width, height := 16, 10
	redrawCh := make(chan struct{})
	window, _ := newWindow(strings.NewReader(""), "test", "test", redrawCh)
	window.setSize(width, height)

	str := "Hello, World!\nこんにちは、世界！\n鰰は魚の一種"
	go func() {
		window.Run()
	}()
	go func() {
		window.eventCh <- Event{Type: EventSwitchFocus}
		window.eventCh <- Event{Type: EventStartInsert}
		for _, r := range str {
			window.eventCh <- Event{Type: EventRune, Rune: r, Mode: ModeInsert}
		}
	}()
	<-redrawCh
	<-redrawCh
	for _ = range str {
		<-redrawCh
	}
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), str+"\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", str+"\x00", string(state.Bytes))
	}
	window.Close()
}

func TestWindowEventUndoRedo(t *testing.T) {
	width, height := 16, 10
	redrawCh := make(chan struct{})
	window, _ := newWindow(strings.NewReader("Hello, world!"), "test", "test", redrawCh)
	window.setSize(width, height)
	waitCh := make(chan struct{})
	defer func() {
		close(waitCh)
		close(redrawCh)
		window.Close()
	}()

	waitRedraw := func(count int) {
		for i := 0; i < count; i++ {
			<-redrawCh
		}
	}
	go func() {
		window.Run()
	}()
	go func() {
		window.eventCh <- Event{Type: EventUndo}
		window.eventCh <- Event{Type: EventSwitchFocus}
		window.eventCh <- Event{Type: EventStartAppend, Mode: ModeInsert}

		<-waitCh
		window.eventCh <- Event{Type: EventRune, Rune: 'x', Mode: ModeInsert}
		window.eventCh <- Event{Type: EventRune, Rune: 'y', Mode: ModeInsert}
		window.eventCh <- Event{Type: EventRune, Rune: 'z', Mode: ModeInsert}
		window.eventCh <- Event{Type: EventExitInsert}

		<-waitCh
		window.eventCh <- Event{Type: EventStartInsert, Mode: ModeInsert}
		window.eventCh <- Event{Type: EventRune, Rune: 'x', Mode: ModeInsert}
		window.eventCh <- Event{Type: EventRune, Rune: 'y', Mode: ModeInsert}
		window.eventCh <- Event{Type: EventCursorLeft, Mode: ModeInsert}
		window.eventCh <- Event{Type: EventRune, Rune: 'z', Mode: ModeInsert}
		window.eventCh <- Event{Type: EventExitInsert}
		window.eventCh <- Event{Type: EventUndo}

		<-waitCh
		window.eventCh <- Event{Type: EventUndo}
		window.eventCh <- Event{Type: EventStartInsert, Mode: ModeInsert}
		window.eventCh <- Event{Type: EventRune, Rune: 'w', Mode: ModeInsert}

		<-waitCh
		window.eventCh <- Event{Type: EventExitInsert}
		window.eventCh <- Event{Type: EventUndo}

		<-waitCh
		window.eventCh <- Event{Type: EventRedo}
		window.eventCh <- Event{Type: EventRedo}
	}()

	waitRedraw(3)
	state, _ := window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hello, world!\x00", string(state.Bytes))
	}
	if state.Cursor != 1 {
		t.Errorf("state.Cursor should be %d but got %d", 1, state.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(4)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hxyzello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hxyzello, world!\x00", string(state.Bytes))
	}
	if state.Cursor != 3 {
		t.Errorf("state.Cursor should be %d but got %d", 3, state.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(7)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hxyxyzello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hxyxyzello, world!\x00", string(state.Bytes))
	}
	if state.Cursor != 5 {
		t.Errorf("state.Cursor should be %d but got %d", 5, state.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(3)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hxywzello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hxywzello, world!\x00", string(state.Bytes))
	}
	if state.Cursor != 4 {
		t.Errorf("state.Cursor should be %d but got %d", 4, state.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(2)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hxyzello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hxyzello, world!\x00", string(state.Bytes))
	}
	if state.Cursor != 3 {
		t.Errorf("state.Cursor should be %d but got %d", 3, state.Cursor)
	}
	waitCh <- struct{}{}

	waitRedraw(2)
	state, _ = window.State()
	if !strings.HasPrefix(string(state.Bytes), "Hxywzello, world!\x00") {
		t.Errorf("state.Bytes should start with %q but got %q", "Hxywzello, world!\x00", string(state.Bytes))
	}
	if state.Cursor != 4 {
		t.Errorf("state.Cursor should be %d but got %d", 4, state.Cursor)
	}
}
