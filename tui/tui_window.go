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
	for i := 0; i < 10; i++ {
		if s.Length <= threshold {
			return 6 + i
		}
		threshold = (threshold << 4) | 0x0f
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
			style := styles[i][j]
			if style == math.MaxInt16 {
				continue
			}
			if style == math.MaxInt16-1 {
				d.setOffset(3*j+1).setString(" ", style.Underline(!active || s.FocusText))
				d.setOffset(3*width+j+3).setString(" ", style.Underline(!active || !s.FocusText))
				continue
			}
			style1, style2 := style, style
			if i*width+j == cursorPos {
				style1 = style1.Reverse(active && !s.FocusText).Bold(
					!active || s.FocusText).Underline(!active || s.FocusText)
				style2 = style2.Reverse(active && s.FocusText).Bold(
					!active || !s.FocusText).Underline(!active || !s.FocusText)
			}
			d.setOffset(3*j+1).setString(fmt.Sprintf("%02x", bytes[i][j]), style1)
			d.setOffset(3*width+j+3).setString(string(prettyByte(bytes[i][j])), style2)
		}
		d.setOffset(-2).setString(" | ", tcell.StyleDefault)
		d.setOffset(3*width).setString(" | ", tcell.StyleDefault)
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
	for 0 < len(eis) && eis[1] <= s.Offset {
		eis = eis[2:]
	}
	bytes := make([][]byte, height)
	styles := make([][]tcell.Style, height)
	color := tcell.ColorLightSeaGreen
	cursorPos := int(s.Cursor - s.Offset)
	for i := 0; i < height; i++ {
		bytes[i] = make([]byte, width)
		styles[i] = make([]tcell.Style, width)
		for j := 0; j < width; j++ {
			if s.Pending && i*width+j == cursorPos {
				bytes[i][j] = s.PendingByte
				styles[i][j] = styles[i][j].Foreground(color)
				if s.Mode == mode.Replace {
					k++
				}
				continue
			}
			if k >= s.Size {
				if k == cursorPos {
					styles[i][j] = tcell.Style(math.MaxInt16 - 1)
				} else {
					styles[i][j] = tcell.Style(math.MaxInt16)
				}
				k++
				continue
			}
			bytes[i][j] = s.Bytes[k]
			pos := int64(k) + s.Offset
			if 0 < len(eis) && eis[0] <= pos && pos < eis[1] {
				styles[i][j] = styles[i][j].Foreground(color)
			} else if 0 < len(eis) && eis[1] <= pos {
				eis = eis[2:]
			}
			if s.VisualStart >= 0 && s.Cursor < s.Length &&
				(s.VisualStart <= pos && pos <= s.Cursor ||
					s.Cursor <= pos && pos <= s.VisualStart) {
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
	offsetStyle := "0x%0" + strconv.Itoa(offsetStyleWidth) + "x"
	j := int(s.Cursor - s.Offset)
	name := s.Name
	if name == "" {
		name = "[No name]"
	}
	left := fmt.Sprintf(" %s%s : 0x%02x : '%s'",
		prettyMode(s.Mode), name, s.Bytes[j], prettyRune(s.Bytes[j]))
	right := fmt.Sprintf("%d/%d : "+offsetStyle+"/"+offsetStyle+" : %.2f%% ",
		s.Cursor, s.Length, s.Cursor, s.Length,
		float64(s.Cursor*100)/float64(mathutil.MaxInt64(s.Length, 1)))
	line := left + strings.Repeat(
		" ", mathutil.MaxInt(2, ui.region.width-len(left)-len(right)),
	) + right
	ui.getTextDrawer().setTop(ui.region.height-1).setString(line, tcell.StyleDefault.Reverse(true))
}

func prettyByte(b byte) byte {
	switch {
	case 0x20 <= b && b < 0x7f:
		return b
	default:
		return 0x2e
	}
}

func prettyRune(b byte) string {
	switch {
	case b == 0x07:
		return "\\a"
	case b == 0x08:
		return "\\b"
	case b == 0x09:
		return "\\t"
	case b == 0x0a:
		return "\\n"
	case b == 0x0b:
		return "\\v"
	case b == 0x0c:
		return "\\f"
	case b == 0x0d:
		return "\\r"
	case b < 0x20:
		return fmt.Sprintf("\\x%02x", b)
	case b == 0x27:
		return "\\'"
	case b < 0x7f:
		return string(rune(b))
	default:
		return fmt.Sprintf("\\u%04x", b)
	}
}

func prettyMode(m mode.Mode) string {
	switch m {
	case mode.Insert:
		return "[INSERT] "
	case mode.Replace:
		return "[REPLACE] "
	case mode.Visual:
		return "[VISUAL] "
	default:
		return ""
	}
}
