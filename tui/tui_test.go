package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell"
	. "github.com/itchyny/bed/common"
)

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
