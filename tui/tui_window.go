package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"

	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/util"
)

type tuiWindow struct {
	region region
	screen tcell.Screen
}

func (ui *tuiWindow) getTextDrawer() *textDrawer {
	return &textDrawer{region: ui.region, screen: ui.screen}
}

func (ui *tuiWindow) setCursor(line int, offset int) {
	ui.screen.ShowCursor(ui.region.left+offset, ui.region.top+line)
}

func (ui *tuiWindow) offsetStyleWidth(state *WindowState) int {
	threshold := int64(0xfffff)
	for i := 0; i < 5; i++ {
		if state.Length <= threshold {
			return 6 + i*2
		}
		threshold = (threshold << 8) | 0xff
	}
	return 16
}

func (ui *tuiWindow) drawWindow(state *WindowState, active bool) {
	height, width := ui.region.height-2, state.Width
	bytes, styles := ui.bytesArray(height, width, state)
	cursorPos := int(state.Cursor - state.Offset)
	cursorLine := cursorPos / width
	offsetStyleWidth := ui.offsetStyleWidth(state)
	offsetStyle := " %0" + strconv.Itoa(offsetStyleWidth) + "x"
	d := ui.getTextDrawer()
	for i := 0; i < height; i++ {
		d.setTop(i + 1).setLeft(0).setOffset(0)
		d.setString(fmt.Sprintf(offsetStyle, state.Offset+int64(i*width)), tcell.StyleDefault.Bold(i == cursorLine))
		d.setLeft(offsetStyleWidth + 3)
		for j := 0; j < width; j++ {
			if styles[i][j] == math.MaxUint16 {
				d.setOffset(3*j).setString("   ", tcell.StyleDefault)
				d.setOffset(3*width+j+3).setString(" ", tcell.StyleDefault)
			} else {
				d.setOffset(3*j).setString(" ", styles[i][j]|tcell.StyleDefault)
				if i*width+j == cursorPos {
					styles[i][j] = styles[i][j].Reverse(active && !state.FocusText).Bold(
						!active || state.FocusText).Underline(!active || state.FocusText)
				}
				d.setOffset(3*j+1).setString(fmt.Sprintf("%02x", bytes[i][j]), styles[i][j])
				if i*width+j == cursorPos {
					styles[i][j] = styles[i][j].Reverse(active && state.FocusText).Bold(
						!active || !state.FocusText).Underline(!active || !state.FocusText)
				}
				d.setOffset(3*width+j+3).setString(string(prettyByte(bytes[i][j])), styles[i][j])
			}
		}
		d.setOffset(-2).setString(" | ", tcell.StyleDefault)
		d.setOffset(3*width).setString(" | ", tcell.StyleDefault)
		d.setOffset(4*width+3).setString(" ", tcell.StyleDefault)
	}
	i := int(state.Cursor % int64(width))
	if active {
		if state.FocusText {
			ui.setCursor(cursorLine+1, 3*width+i+6+offsetStyleWidth)
		} else if state.Pending {
			ui.setCursor(cursorLine+1, 3*i+5+offsetStyleWidth)
		} else {
			ui.setCursor(cursorLine+1, 3*i+4+offsetStyleWidth)
		}
	}
	ui.drawHeader(state, offsetStyleWidth)
	ui.drawScrollBar(state, height, 4*width+7+offsetStyleWidth)
	ui.drawFooter(state, offsetStyleWidth)
}

func (ui *tuiWindow) bytesArray(height, width int, state *WindowState) ([][]byte, [][]tcell.Style) {
	var k int
	if height <= 0 {
		return nil, nil
	}
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
				styles[i][j] = styles[i][j].Foreground(tcell.ColorDodgerBlue)
				if state.Mode == ModeReplace {
					k++
				}
				continue
			}
			bytes[i][j] = state.Bytes[k]
			if 0 < len(eis) && eis[0] <= int64(k)+state.Offset && int64(k)+state.Offset < eis[1] {
				styles[i][j] = styles[i][j].Foreground(tcell.ColorDodgerBlue)
			} else if 0 < len(eis) && eis[1] <= int64(k)+state.Offset {
				eis = eis[2:]
			}
			k++
		}
	}
	return bytes, styles
}

func (ui *tuiWindow) drawHeader(state *WindowState, offsetStyleWidth int) {
	style := tcell.StyleDefault.Underline(true)
	d := ui.getTextDrawer()
	d.setString(strings.Repeat(" ", 4*state.Width+8+offsetStyleWidth), style)
	d.setLeft(offsetStyleWidth)
	cursor := int(state.Cursor % int64(state.Width))
	for i := 0; i < state.Width; i++ {
		d.setOffset(3*i+4).setString(fmt.Sprintf("%2x", i), style.Bold(cursor == i))
	}
	d.setOffset(2).setString("|", style)
	d.setOffset(3*state.Width+4).setString("|", style)
}

func (ui *tuiWindow) drawScrollBar(state *WindowState, height int, left int) {
	stateSize := state.Size
	if state.Cursor+1 == state.Length && state.Cursor == state.Offset+int64(state.Size) {
		stateSize++
	}
	total := int64((stateSize + state.Width - 1) / state.Width)
	len := util.MaxInt64((state.Length+int64(state.Width)-1)/int64(state.Width), 1)
	size := util.MaxInt64(total*total/len, 1)
	pad := (total*total + len - len*size - 1) / util.MaxInt64(total-size+1, 1)
	top := (state.Offset / int64(state.Width) * total) / (len - pad)
	d := ui.getTextDrawer().setLeft(left)
	for i := 0; i < height; i++ {
		d.setTop(i + 1)
		if int(top) <= i && i < int(top+size) {
			d.setString("#", tcell.StyleDefault)
		} else {
			d.setString("|", tcell.StyleDefault)
		}
	}
}

func (ui *tuiWindow) drawFooter(state *WindowState, offsetStyleWidth int) {
	offsetStyle := "%0" + strconv.Itoa(offsetStyleWidth) + "x"
	j := int(state.Cursor - state.Offset)
	name := state.Name
	if name == "" {
		name = "[No name]"
	}
	line := fmt.Sprintf(" %s%s: "+offsetStyle+" / "+offsetStyle+
		" (%.2f%%) [0x%02x '%s']"+strings.Repeat(" ", ui.region.width),
		prettyMode(state.Mode), name, state.Cursor, state.Length,
		float64(state.Cursor*100)/float64(util.MaxInt64(state.Length, 1)),
		state.Bytes[j], prettyRune(state.Bytes[j]))
	ui.getTextDrawer().setTop(ui.region.height-1).setString(line, tcell.StyleDefault.Reverse(true))
}
