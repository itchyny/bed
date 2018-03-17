package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"

	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/util"
)

type tuiWindow struct {
	region region
	screen tcell.Screen
}

func (ui *tuiWindow) setLine(line int, offset int, str string, style tcell.Style) {
	line += ui.region.top
	offset += ui.region.left
	for _, c := range str {
		ui.screen.SetContent(offset, line, c, nil, style)
		offset += runewidth.RuneWidth(c)
	}
}

func (ui *tuiWindow) setCursor(line int, offset int) {
	ui.screen.ShowCursor(ui.region.left+offset, ui.region.top+line)
}

func (ui *tuiWindow) drawWindow(state WindowState, active bool) {
	height, width := ui.region.height-2, state.Width
	bytes, styles := ui.bytesArray(height, width, state)
	cursorPos := int(state.Cursor - state.Offset)
	cursorLine := cursorPos / width
	for i := 0; i < height; i++ {
		style := tcell.StyleDefault.Underline(i == height-1)
		for j := 0; j < width; j++ {
			if styles[i][j] == math.MaxUint16 {
				ui.setLine(i+1, 3*j+11, "   ", style)
				ui.setLine(i+1, 3*width+j+14, " ", style)
			} else {
				ui.setLine(i+1, 3*j+11, " ", styles[i][j]|style)
				if i*width+j == cursorPos {
					styles[i][j] = styles[i][j].Reverse(!state.FocusText).Bold(state.FocusText).Underline(state.FocusText)
				}
				ui.setLine(i+1, 3*j+12, fmt.Sprintf("%02x", bytes[i][j]), styles[i][j]|style)
				if i*width+j == cursorPos {
					styles[i][j] = styles[i][j].Reverse(state.FocusText).Bold(!state.FocusText).Underline(!state.FocusText)
				}
				ui.setLine(i+1, 3*width+j+14, string(prettyByte(bytes[i][j])), styles[i][j]|style)
			}
		}
		ui.setLine(i+1, 0, fmt.Sprintf(" %08x", state.Offset+int64(i*width)), style.Bold(i == cursorLine))
		ui.setLine(i+1, 9, " | ", style)
		ui.setLine(i+1, 3*width+11, " | ", style)
		ui.setLine(i+1, 4*width+14, " ", style)
	}
	i := int(state.Cursor % int64(width))
	if active {
		if state.FocusText {
			ui.setCursor(cursorLine+1, 3*width+i+14)
		} else if state.Pending {
			ui.setCursor(cursorLine+1, 3*i+13)
		} else {
			ui.setCursor(cursorLine+1, 3*i+12)
		}
	}
	ui.drawHeader(state)
	ui.drawScrollBar(state, height, 4*width+15)
	ui.drawFooter(state)
}

func (ui *tuiWindow) bytesArray(height, width int, state WindowState) ([][]byte, [][]tcell.Style) {
	var k int
	eis := state.EditedIndices
	bytes := make([][]byte, height)
	styles := make([][]tcell.Style, height)
	for i := 0; i < height; i++ {
		bytes[i] = make([]byte, width)
		styles[i] = make([]tcell.Style, width)
		for j := 0; j < width; j++ {
			if k >= state.Size {
				styles[i][j] = tcell.Style(math.MaxUint16)
			}
			if state.Pending && i*width+j == int(state.Cursor-state.Offset) {
				bytes[i][j] = state.PendingByte
				styles[i][j] = styles[i][j].Foreground(tcell.ColorSkyblue)
				if state.Mode == ModeReplace {
					k++
				}
				continue
			}
			bytes[i][j] = state.Bytes[k]
			if 0 < len(eis) && eis[0] <= int64(k)+state.Offset && int64(k)+state.Offset < eis[1] {
				styles[i][j] = styles[i][j].Foreground(tcell.ColorSkyblue)
			} else if 0 < len(eis) && eis[1] <= int64(k)+state.Offset {
				eis = eis[2:]
			}
			k++
		}
	}
	return bytes, styles
}

func (ui *tuiWindow) drawHeader(state WindowState) {
	style := tcell.StyleDefault.Underline(true)
	ui.setLine(0, 0, strings.Repeat(" ", 4*state.Width+15), style)
	cursor := int(state.Cursor % int64(state.Width))
	for i := 0; i < state.Width; i++ {
		ui.setLine(0, 3*i+12, fmt.Sprintf("%2x", i), style.Bold(cursor == i))
	}
	ui.setLine(0, 10, "|", style)
	ui.setLine(0, 3*state.Width+12, "|", style)
}

func (ui *tuiWindow) drawScrollBar(state WindowState, height int, offset int) {
	stateSize := state.Size
	if state.Cursor+1 == state.Length && state.Cursor == state.Offset+int64(state.Size) {
		stateSize++
	}
	total := int64((stateSize + state.Width - 1) / state.Width)
	len := util.MaxInt64((state.Length+int64(state.Width)-1)/int64(state.Width), 1)
	size := util.MaxInt64(total*total/len, 1)
	pad := (total*total + len - len*size - 1) / util.MaxInt64(total-size+1, 1)
	top := (state.Offset / int64(state.Width) * total) / (len - pad)
	for i := 0; i < height; i++ {
		if int(top) <= i && i < int(top+size) {
			ui.setLine(i+1, offset, " ", tcell.StyleDefault.Reverse(true))
		} else if i < height-1 {
			ui.setLine(i+1, offset, "|", 0)
		} else {
			ui.setLine(i+1, offset, "|", tcell.StyleDefault.Underline(true))
		}
	}
}

func (ui *tuiWindow) drawFooter(state WindowState) {
	j := int(state.Cursor - state.Offset)
	name := state.Name
	if name == "" {
		name = "[No name]"
	}
	line := fmt.Sprintf("%s%s: %08x / %08x (%.2f%%) [0x%02x '%s']"+strings.Repeat(" ", ui.region.width),
		prettyMode(state.Mode), name, state.Cursor, state.Length, float64(state.Cursor*100)/float64(util.MaxInt64(state.Length, 1)),
		state.Bytes[j], prettyRune(state.Bytes[j]))
	ui.setLine(ui.region.height-1, 1, line, 0)
}
