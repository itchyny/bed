package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"

	"github.com/itchyny/bed/mathutil"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/state"
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

func (ui *tuiWindow) offsetStyleWidth(s *state.WindowState) int {
	threshold := int64(0xfffff)
	for i := 0; i < 5; i++ {
		if s.Length <= threshold {
			return 6 + i*2
		}
		threshold = (threshold << 8) | 0xff
	}
	return 16
}

func (ui *tuiWindow) drawWindow(s *state.WindowState, active bool) {
	height, width := ui.region.height-2, s.Width
	bytes, styles := ui.bytesArray(height, width, s)
	cursorPos := int(s.Cursor - s.Offset)
	cursorLine := cursorPos / width
	offsetStyleWidth := ui.offsetStyleWidth(s)
	offsetStyle := " %0" + strconv.Itoa(offsetStyleWidth) + "x"
	d := ui.getTextDrawer()
	for i := 0; i < height; i++ {
		d.setTop(i + 1).setLeft(0).setOffset(0)
		d.setString(fmt.Sprintf(offsetStyle, s.Offset+int64(i*width)), tcell.StyleDefault.Bold(i == cursorLine))
		d.setLeft(offsetStyleWidth + 3)
		for j := 0; j < width; j++ {
			if styles[i][j] == math.MaxUint16 {
				d.setOffset(3*j).setString("   ", tcell.StyleDefault)
				d.setOffset(3*width+j+3).setString(" ", tcell.StyleDefault)
			} else {
				d.setOffset(3*j).setString(" ", tcell.StyleDefault)
				if i*width+j == cursorPos {
					styles[i][j] = styles[i][j].Reverse(active && !s.FocusText).Bold(
						!active || s.FocusText).Underline(!active || s.FocusText)
				}
				d.setOffset(3*j+1).setString(fmt.Sprintf("%02x", bytes[i][j]), styles[i][j])
				if i*width+j == cursorPos {
					styles[i][j] = styles[i][j].Reverse(active && s.FocusText).Bold(
						!active || !s.FocusText).Underline(!active || !s.FocusText)
				}
				d.setOffset(3*width+j+3).setString(string(prettyByte(bytes[i][j])), styles[i][j])
			}
		}
		d.setOffset(-2).setString(" | ", tcell.StyleDefault)
		d.setOffset(3*width).setString(" | ", tcell.StyleDefault)
		d.setOffset(4*width+3).setString(" ", tcell.StyleDefault)
	}
	i := int(s.Cursor % int64(width))
	if active {
		if s.FocusText {
			ui.setCursor(cursorLine+1, 3*width+i+6+offsetStyleWidth)
		} else if s.Pending {
			ui.setCursor(cursorLine+1, 3*i+5+offsetStyleWidth)
		} else {
			ui.setCursor(cursorLine+1, 3*i+4+offsetStyleWidth)
		}
	}
	ui.drawHeader(s, offsetStyleWidth)
	ui.drawScrollBar(s, height, 4*width+7+offsetStyleWidth)
	ui.drawFooter(s, offsetStyleWidth)
}

func (ui *tuiWindow) bytesArray(height, width int, s *state.WindowState) ([][]byte, [][]tcell.Style) {
	var k int
	if height <= 0 {
		return nil, nil
	}
	eis := s.EditedIndices
	bytes := make([][]byte, height)
	styles := make([][]tcell.Style, height)
	for i := 0; i < height; i++ {
		bytes[i] = make([]byte, width)
		styles[i] = make([]tcell.Style, width)
		for j := 0; j < width; j++ {
			if k >= s.Size {
				styles[i][j] = tcell.Style(math.MaxUint16)
			}
			if s.Pending && i*width+j == int(s.Cursor-s.Offset) {
				bytes[i][j] = s.PendingByte
				styles[i][j] = styles[i][j].Foreground(tcell.ColorDodgerBlue)
				if s.Mode == mode.Replace {
					k++
				}
				continue
			}
			bytes[i][j] = s.Bytes[k]
			if 0 < len(eis) && eis[0] <= int64(k)+s.Offset && int64(k)+s.Offset < eis[1] {
				styles[i][j] = styles[i][j].Foreground(tcell.ColorDodgerBlue)
			} else if 0 < len(eis) && eis[1] <= int64(k)+s.Offset {
				eis = eis[2:]
			}
			if s.VisualStart >= 0 && s.Cursor < s.Length &&
				(s.VisualStart <= int64(k)+s.Offset && int64(k)+s.Offset <= s.Cursor ||
					s.Cursor <= int64(k)+s.Offset && int64(k)+s.Offset <= s.VisualStart) {
				styles[i][j] = styles[i][j].Underline(true)
			}
			k++
		}
	}
	return bytes, styles
}

func (ui *tuiWindow) drawHeader(s *state.WindowState, offsetStyleWidth int) {
	style := tcell.StyleDefault.Underline(true)
	d := ui.getTextDrawer()
	d.setString(strings.Repeat(" ", 4*s.Width+8+offsetStyleWidth), style)
	d.setLeft(offsetStyleWidth)
	cursor := int(s.Cursor % int64(s.Width))
	for i := 0; i < s.Width; i++ {
		d.setOffset(3*i+4).setString(fmt.Sprintf("%2x", i), style.Bold(cursor == i))
	}
	d.setOffset(2).setString("|", style)
	d.setOffset(3*s.Width+4).setString("|", style)
}

func (ui *tuiWindow) drawScrollBar(s *state.WindowState, height int, left int) {
	stateSize := s.Size
	if s.Cursor+1 == s.Length && s.Cursor == s.Offset+int64(s.Size) {
		stateSize++
	}
	total := int64((stateSize + s.Width - 1) / s.Width)
	len := mathutil.MaxInt64((s.Length+int64(s.Width)-1)/int64(s.Width), 1)
	size := mathutil.MaxInt64(total*total/len, 1)
	pad := (total*total + len - len*size - 1) / mathutil.MaxInt64(total-size+1, 1)
	top := (s.Offset / int64(s.Width) * total) / (len - pad)
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

func (ui *tuiWindow) drawFooter(s *state.WindowState, offsetStyleWidth int) {
	offsetStyle := "%0" + strconv.Itoa(offsetStyleWidth) + "x"
	j := int(s.Cursor - s.Offset)
	name := s.Name
	if name == "" {
		name = "[No name]"
	}
	line := fmt.Sprintf(" %s%s: "+offsetStyle+" / "+offsetStyle+
		" (%.2f%%) [0x%02x '%s']"+strings.Repeat(" ", ui.region.width),
		prettyMode(s.Mode), name, s.Cursor, s.Length,
		float64(s.Cursor*100)/float64(mathutil.MaxInt64(s.Length, 1)),
		s.Bytes[j], prettyRune(s.Bytes[j]))
	ui.getTextDrawer().setTop(ui.region.height-1).setString(line, tcell.StyleDefault.Reverse(true))
}
