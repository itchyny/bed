package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell"

	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/key"
	"github.com/itchyny/bed/layout"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/state"
)

func mockKeyManager() map[mode.Mode]*key.Manager {
	kms := make(map[mode.Mode]*key.Manager)
	km := key.NewManager(true)
	km.Register(event.Quit, "Z", "Q")
	km.Register(event.CursorDown, "j")
	kms[mode.Normal] = km
	return kms
}

func getContents(screen tcell.SimulationScreen) string {
	width, _ := screen.Size()
	cells, _, _ := screen.GetContents()
	var runes []rune
	for i, cell := range cells {
		runes = append(runes, cell.Runes...)
		if (i+1)%width == 0 {
			runes = append(runes, '\n')
		}
	}
	return string(runes)
}

func shouldContain(t *testing.T, screen tcell.SimulationScreen, expected []string) {
	got := getContents(screen)
	for _, str := range expected {
		if !strings.Contains(got, str) {
			t.Errorf("screen should contain %q but got\n%v", str, got)
		}
	}
}

func TestTuiRun(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan event.Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(90, 20)
	go ui.Run(mockKeyManager())
	screen.InjectKey(tcell.KeyRune, 'Z', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'Q', tcell.ModNone)
	e := <-eventCh
	if e.Type != event.Rune {
		t.Errorf("pressing Z should emit event.Rune but got: %+v", e)
	}
	e = <-eventCh
	if e.Type != event.Quit {
		t.Errorf("pressing ZQ should emit event.Quit but got: %+v", e)
	}
	screen.InjectKey(tcell.KeyRune, '7', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, '0', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, '9', tcell.ModNone)
	screen.InjectKey(tcell.KeyRune, 'j', tcell.ModNone)
	e = <-eventCh
	e = <-eventCh
	e = <-eventCh
	e = <-eventCh
	if e.Type != event.CursorDown {
		t.Errorf("pressing 709j should emit event.CursorDown but got: %+v", e)
	}
	if e.Count != 709 {
		t.Errorf("pressing 709j should emit event with count %d but got: %+v", 709, e)
	}
	ui.Close()
}

func TestTuiEmpty(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan event.Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(90, 20)
	width, height := screen.Size()

	s := state.State{
		WindowStates: map[int]*state.WindowState{
			0: &state.WindowState{
				Name:   "",
				Width:  16,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte(strings.Repeat("\x00", 16*(height-1))),
				Size:   16 * (height - 1),
				Length: 0,
				Mode:   mode.Normal,
			},
		},
		Layout: layout.NewLayout(0).Resize(0, 0, width, height-1),
	}
	ui.Redraw(s)

	shouldContain(t, screen, []string{
		"        |  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f |                   ",
		" 000000 | 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 | ................ #",
		" 000010 | 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 | ................ #",
		" 000020 | 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 | ................ #",
		" 000100 | 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 | ................ #",
		" [No name]: 000000 / 000000 (0.00%) [0x00 '\\x00']                            ",
	})

	x, y, visible := screen.GetCursor()
	if x != 10 || y != 1 {
		t.Errorf("cursor position should be (%d, %d) but got (%d, %d)", 10, 1, x, y)
	}
	if visible != true {
		t.Errorf("cursor should be visible but got %v", visible)
	}
	ui.Close()
}

func TestTuiScrollBar(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan event.Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(90, 20)
	width, height := screen.Size()

	s := state.State{
		WindowStates: map[int]*state.WindowState{
			0: &state.WindowState{
				Name:   "",
				Width:  16,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte(strings.Repeat("a", 16*(height-1))),
				Size:   16 * (height - 1),
				Length: int64(16 * (height - 1) * 3),
				Mode:   mode.Normal,
			},
		},
		Layout: layout.NewLayout(0).Resize(0, 0, width, height-1),
	}
	ui.Redraw(s)

	shouldContain(t, screen, []string{
		"        |  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f |                    ",
		" 000000 | 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 | aaaaaaaaaaaaaaaa # ",
		" 000050 | 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 | aaaaaaaaaaaaaaaa # ",
		" 000060 | 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 | aaaaaaaaaaaaaaaa | ",
		" 000100 | 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 | aaaaaaaaaaaaaaaa | ",
		" [No name]: 000000 / 000390 (0.00%) [0x61 'a']                                 ",
	})

	x, y, visible := screen.GetCursor()
	if x != 10 || y != 1 {
		t.Errorf("cursor position should be (%d, %d) but got (%d, %d)", 10, 1, x, y)
	}
	if visible != true {
		t.Errorf("cursor should be visible but got %v", visible)
	}
	ui.Close()
}

func TestTuiHorizontalSplit(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan event.Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(110, 20)
	width, height := screen.Size()

	s := state.State{
		WindowStates: map[int]*state.WindowState{
			0: &state.WindowState{
				Name:   "test0",
				Width:  16,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte("Test window 0." + strings.Repeat("\x00", 110*10)),
				Size:   110 * 10,
				Length: 600,
				Mode:   mode.Normal,
			},
			1: &state.WindowState{
				Name:   "test1",
				Width:  16,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte("Test window 1." + strings.Repeat(" ", 110*10)),
				Size:   110 * 10,
				Length: 800,
				Mode:   mode.Normal,
			},
		},
		Layout: layout.NewLayout(0).SplitBottom(1).Resize(0, 0, width, height-1),
	}
	ui.Redraw(s)

	shouldContain(t, screen, []string{
		" 000000 | 54 65 73 74 20 77 69 6e 64 6f 77 20 30 2e 00 00 | Test window 0... #",
		" 000010 | 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 | ................ #",
		" test0: 000000 / 000258 (0.00%) [0x54 'T']                                    ",
		"        |  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f |                   ",
		" 000000 | 54 65 73 74 20 77 69 6e 64 6f 77 20 31 2e 20 20 | Test window 1.   #",
		" 000010 | 20 20 20 20 20 20 20 20 20 20 20 20 20 20 20 20 |                  #",
		" test1: 000000 / 000320 (0.00%) [0x54 'T']                                    ",
	})

	x, y, visible := screen.GetCursor()
	if x != 10 || y != 10 {
		t.Errorf("cursor position should be (%d, %d) but got (%d, %d)", 10, 10, x, y)
	}
	if visible != true {
		t.Errorf("cursor should be visible but got %v", visible)
	}
	ui.Close()
}

func TestTuiVerticalSplit(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan event.Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(110, 20)
	width, height := screen.Size()

	s := state.State{
		WindowStates: map[int]*state.WindowState{
			0: &state.WindowState{
				Name:   "test0",
				Width:  8,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte("Test window 0." + strings.Repeat("\x00", 55*19)),
				Size:   55 * 19,
				Length: 600,
				Mode:   mode.Normal,
			},
			1: &state.WindowState{
				Name:   "test1",
				Width:  8,
				Offset: 0,
				Cursor: 0,
				Bytes:  []byte("Test window 1." + strings.Repeat(" ", 54*19)),
				Size:   54 * 19,
				Length: 800,
				Mode:   mode.Normal,
			},
		},
		Layout: layout.NewLayout(0).SplitRight(1).Resize(0, 0, width, height-1),
	}
	ui.Redraw(s)

	shouldContain(t, screen, []string{
		"        |  0  1  2  3  4  5  6  7 |                    |        |  0  1  2  3  4  5  6  7 |",
		" 000000 | 54 65 73 74 20 77 69 6e | Test win #         | 000000 | 54 65 73 74 20 77 69 6e | Test win #",
		" 000008 | 64 6f 77 20 30 2e 00 00 | dow 0... #         | 000008 | 64 6f 77 20 31 2e 20 20 | dow 1.   #",
		" 000010 | 00 00 00 00 00 00 00 00 | ........ #         | 000010 | 20 20 20 20 20 20 20 20 |          #",
		" test0: 000000 / 000258 (0.00%) [0x54 'T']             | test1: 000000 / 000320 (0.00%) [0x54 'T']",
	})

	x, y, visible := screen.GetCursor()
	if x != 66 || y != 1 {
		t.Errorf("cursor position should be (%d, %d) but got (%d, %d)", 66, 1, x, y)
	}
	if visible != true {
		t.Errorf("cursor should be visible but got %v", visible)
	}
	ui.Close()
}

func TestTuiCmdline(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan event.Event)
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

	s := state.State{
		Mode:          mode.Cmdline,
		Cmdline:       []rune("vnew test"),
		CmdlineCursor: 9,
	}
	ui.Redraw(s)

	got, expected := getCmdline(), ":vnew test "
	if !strings.HasPrefix(got, expected) {
		t.Errorf("cmdline should start with %q but got %q", expected, got)
	}

	s = state.State{
		Mode:          mode.Normal,
		Error:         errors.New("error"),
		Cmdline:       []rune("vnew test"),
		CmdlineCursor: 9,
	}
	ui.Redraw(s)

	got, expected = getCmdline(), "error "
	if !strings.HasPrefix(got, expected) {
		t.Errorf("cmdline should start with %q but got %q", expected, got)
	}
	ui.Close()
}

func TestTuiCmdlineCompletionCandidates(t *testing.T) {
	ui := NewTui()
	eventCh := make(chan event.Event)
	screen := tcell.NewSimulationScreen("")
	if err := ui.initForTest(eventCh, screen); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(20, 15)

	s := state.State{
		Mode:              mode.Cmdline,
		Cmdline:           []rune("new test2"),
		CmdlineCursor:     9,
		CompletionResults: []string{"test1", "test2", "test3", "test9/", "/bin/ls"},
		CompletionIndex:   1,
	}
	ui.Redraw(s)

	shouldContain(t, screen, []string{
		" test1  test2  test3",
		":new test2",
	})

	s.CompletionIndex += 2
	s.Cmdline = []rune("new test9/")
	ui.Redraw(s)

	shouldContain(t, screen, []string{
		" test3  test9/  /bin",
		":new test9/",
	})
	ui.Close()
}
