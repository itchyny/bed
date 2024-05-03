package tui

import (
	"cmp"
	"fmt"

	"github.com/gdamore/tcell"

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
	for i := range 10 {
		if s.Length <= threshold {
			return 6 + i
		}
		threshold = (threshold << 4) | 0x0f
	}
	return 16
}

func (ui *tuiWindow) drawWindow(s *state.WindowState, active bool) {
	height, width := ui.region.height-2, s.Width
	cursorPos := int(s.Cursor - s.Offset)
	cursorLine := cursorPos / width
	offsetStyleWidth := ui.offsetStyleWidth(s)
	eis := s.EditedIndices
	for 0 < len(eis) && eis[1] <= s.Offset {
		eis = eis[2:]
	}
	editedColor := tcell.ColorLightSeaGreen
	d := ui.getTextDrawer()
	var k int
	for i := range height {
		d.addTop(1).setLeft(0).setOffset(0)
		d.setString(fmt.Sprintf(" %0*x", offsetStyleWidth, s.Offset+int64(i*width)), tcell.StyleDefault.Bold(i == cursorLine))
		d.setLeft(offsetStyleWidth + 3)
		for j := range width {
			b, style := byte(0), tcell.StyleDefault
			if s.Pending && i*width+j == cursorPos {
				b, style = s.PendingByte, tcell.StyleDefault.Foreground(editedColor)
				if s.Mode != mode.Replace {
					k--
				}
			} else if k >= s.Size {
				if k == cursorPos {
					d.setOffset(3*j+1).setByte(' ', tcell.StyleDefault.Underline(!active || s.FocusText))
					d.setOffset(3*width+j+3).setByte(' ', tcell.StyleDefault.Underline(!active || !s.FocusText))
				}
				k++
				continue
			} else {
				b = s.Bytes[k]
				pos := int64(k) + s.Offset
				if 0 < len(eis) && eis[0] <= pos && pos < eis[1] {
					style = tcell.StyleDefault.Foreground(editedColor)
				} else if 0 < len(eis) && eis[1] <= pos {
					eis = eis[2:]
				}
				if s.VisualStart >= 0 && s.Cursor < s.Length &&
					(s.VisualStart <= pos && pos <= s.Cursor ||
						s.Cursor <= pos && pos <= s.VisualStart) {
					style = style.Underline(true)
				}
			}
			style1, style2 := style, style
			if i*width+j == cursorPos {
				style1 = style1.Reverse(active && !s.FocusText).Bold(
					!active || s.FocusText).Underline(!active || s.FocusText)
				style2 = style2.Reverse(active && s.FocusText).Bold(
					!active || !s.FocusText).Underline(!active || !s.FocusText)
			}
			d.setOffset(3*j+1).setByte(hex[b>>4], style1)
			d.setOffset(3*j+2).setByte(hex[b&0x0f], style1)
			d.setOffset(3*width+j+3).setByte(prettyByte(b), style2)
			k++
		}
		d.setOffset(-2).setByte(' ', tcell.StyleDefault)
		d.setOffset(-1).setByte('|', tcell.StyleDefault)
		d.setOffset(0).setByte(' ', tcell.StyleDefault)
		d.addLeft(3*width).setByte(' ', tcell.StyleDefault)
		d.setOffset(1).setByte('|', tcell.StyleDefault)
		d.setOffset(2).setByte(' ', tcell.StyleDefault)
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

const hex = "0123456789abcdef"

func (ui *tuiWindow) drawHeader(s *state.WindowState, offsetStyleWidth int) {
	style := tcell.StyleDefault.Underline(true)
	d := ui.getTextDrawer().setLeft(-1)
	cursor := int(s.Cursor % int64(s.Width))
	for range offsetStyleWidth + 2 {
		d.addLeft(1).setByte(' ', style)
	}
	d.addLeft(1).setByte('|', style)
	for i := range s.Width {
		d.addLeft(1).setByte(' ', style)
		d.addLeft(1).setByte(" 123456789abcdef"[i>>4], style.Bold(cursor == i))
		d.addLeft(1).setByte(hex[i&0x0f], style.Bold(cursor == i))
	}
	d.addLeft(1).setByte(' ', style)
	d.addLeft(1).setByte('|', style)
	for range s.Width + 3 {
		d.addLeft(1).setByte(' ', style)
	}
}

func (ui *tuiWindow) drawScrollBar(s *state.WindowState, height int, left int) {
	stateSize := s.Size
	if s.Cursor+1 == s.Length && s.Cursor == s.Offset+int64(s.Size) {
		stateSize++
	}
	total := int64((stateSize + s.Width - 1) / s.Width)
	length := max((s.Length+int64(s.Width)-1)/int64(s.Width), 1)
	size := max(total*total/length, 1)
	pad := (total*total + length - length*size - 1) / max(total-size+1, 1)
	top := (s.Offset / int64(s.Width) * total) / (length - pad)
	d := ui.getTextDrawer().setLeft(left)
	for i := range height {
		var b byte
		if int(top) <= i && i < int(top+size) {
			b = '#'
		} else {
			b = '|'
		}
		d.addTop(1).setByte(b, tcell.StyleDefault)
	}
}

func (ui *tuiWindow) drawFooter(s *state.WindowState, offsetStyleWidth int) {
	var modified string
	if s.Modified {
		modified = " : +"
	}
	b := s.Bytes[int(s.Cursor-s.Offset)]
	left := fmt.Sprintf(" %s%s%s : 0x%02x : '%s'",
		prettyMode(s.Mode), cmp.Or(s.Name, "[No name]"), modified, b, prettyRune(b))
	right := fmt.Sprintf("%d/%d : 0x%0*x/0x%0*x : %.2f%% ",
		s.Cursor, s.Length, offsetStyleWidth, s.Cursor, offsetStyleWidth, s.Length,
		float64(s.Cursor*100)/float64(max(s.Length, 1)))
	line := fmt.Sprintf("%s  %*s", left, max(ui.region.width-len(left)-2, 0), right)
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
	switch b {
	case 0x07:
		return "\\a"
	case 0x08:
		return "\\b"
	case 0x09:
		return "\\t"
	case 0x0a:
		return "\\n"
	case 0x0b:
		return "\\v"
	case 0x0c:
		return "\\f"
	case 0x0d:
		return "\\r"
	case 0x27:
		return "\\'"
	default:
		if b < 0x20 {
			return fmt.Sprintf("\\x%02x", b)
		} else if b < 0x7f {
			return string(rune(b))
		} else {
			return fmt.Sprintf("\\u%04x", b)
		}
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
