package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell"
	. "github.com/itchyny/bed/common"
)

func TestTuiHorizontalSplit(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(110, 20)
	width, height := screen.Size()

	layout := NewLayout(0).SplitBottom(1).Resize(0, 0, width, height-1)
	state := State{
		WindowStates: map[int]*WindowState{
			0: &WindowState{
				Name:   "test0",
				Width:  16,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte("Test window 0." + strings.Repeat("\x00", 110*10)),
				Size:   110 * 10,
				Length: 600,
				Mode:   ModeNormal,
			},
			1: &WindowState{
				Name:   "test1",
				Width:  16,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte("Test window 1." + strings.Repeat(" ", 110*10)),
				Size:   110 * 10,
				Length: 800,
				Mode:   ModeNormal,
			},
		},
		Layout: layout,
	}
	ui.Redraw(state)

	cells, _, _ := screen.GetContents()
	var runes []rune
	for i, cell := range cells {
		runes = append(runes, cell.Runes...)
		if (i+1)%width == 0 {
			runes = append(runes, '\n')
		}
	}
	got := string(runes)
	expectedStrs := []string{
		" 000000 | 54 65 73 74 20 77 69 6e 64 6f 77 20 30 2e 00 00 | Test window 0... #",
		" 000010 | 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 | ................ #",
		" test0: 000000 / 000258 (0.00%) [0x54 'T']                                    ",
		"        |  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f |                   ",
		" 000000 | 54 65 73 74 20 77 69 6e 64 6f 77 20 31 2e 20 20 | Test window 1.   #",
		" 000010 | 20 20 20 20 20 20 20 20 20 20 20 20 20 20 20 20 |                  #",
		" test1: 000000 / 000320 (0.00%) [0x54 'T']                                    ",
	}
	for _, expected := range expectedStrs {
		if !strings.Contains(got, expected) {
			t.Errorf("screen should contain %q but got %v", expected, got)
		}
	}

	x, y, visible := screen.GetCursor()
	if x != 10 || y != 10 {
		t.Errorf("cursor position should be (%d, %d) but got (%d, %d)", 10, 10, x, y)
	}
	if visible != true {
		t.Errorf("cursor should be visible but got %v", visible)
	}
}

func TestTuiVerticalSplit(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(110, 20)
	width, height := screen.Size()

	layout := NewLayout(0).SplitRight(1).Resize(0, 0, width, height-1)
	state := State{
		WindowStates: map[int]*WindowState{
			0: &WindowState{
				Name:   "test0",
				Width:  8,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte("Test window 0." + strings.Repeat("\x00", 55*19)),
				Size:   55 * 19,
				Length: 600,
				Mode:   ModeNormal,
			},
			1: &WindowState{
				Name:   "test1",
				Width:  8,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte("Test window 1." + strings.Repeat(" ", 54*19)),
				Size:   54 * 19,
				Length: 800,
				Mode:   ModeNormal,
			},
		},
		Layout: layout,
	}
	ui.Redraw(state)

	cells, _, _ := screen.GetContents()
	var runes []rune
	for i, cell := range cells {
		runes = append(runes, cell.Runes...)
		if (i+1)%width == 0 {
			runes = append(runes, '\n')
		}
	}
	got := string(runes)
	expectedStrs := []string{
		"        |  0  1  2  3  4  5  6  7 |                    |        |  0  1  2  3  4  5  6  7 |",
		" 000000 | 54 65 73 74 20 77 69 6e | Test win #         | 000000 | 54 65 73 74 20 77 69 6e | Test win #",
		" 000008 | 64 6f 77 20 30 2e 00 00 | dow 0... #         | 000008 | 64 6f 77 20 31 2e 20 20 | dow 1.   #",
		" 000010 | 00 00 00 00 00 00 00 00 | ........ #         | 000010 | 20 20 20 20 20 20 20 20 |          #",
		" test0: 000000 / 000258 (0.00%) [0x54 'T']             | test1: 000000 / 000320 (0.00%) [0x54 'T']",
	}
	for _, expected := range expectedStrs {
		if !strings.Contains(got, expected) {
			t.Errorf("screen should contain %q but got %v", expected, got)
		}
	}

	x, y, visible := screen.GetCursor()
	if x != 66 || y != 1 {
		t.Errorf("cursor position should be (%d, %d) but got (%d, %d)", 66, 1, x, y)
	}
	if visible != true {
		t.Errorf("cursor should be visible but got %v", visible)
	}
}

func TestTuiCmdline(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(20, 15)
	getCmdline := func() string {
		cells, _, _ := screen.GetContents()
		var runes []rune
		for _, cell := range cells[20*14:] {
			runes = append(runes, cell.Runes...)
		}
		return string(runes)
	}

	state := State{
		Mode:          ModeCmdline,
		Cmdline:       []rune("vnew test"),
		CmdlineCursor: 9,
	}
	ui.Redraw(state)

	got, expected := getCmdline(), ":vnew test "
	if !strings.HasPrefix(got, expected) {
		t.Errorf("cmdline should start with %q but got %q", expected, got)
	}

	state = State{
		Mode:          ModeNormal,
		Error:         errors.New("error"),
		Cmdline:       []rune("vnew test"),
		CmdlineCursor: 9,
	}
	ui.Redraw(state)

	got, expected = getCmdline(), "error "
	if !strings.HasPrefix(got, expected) {
		t.Errorf("cmdline should start with %q but got %q", expected, got)
	}
}
