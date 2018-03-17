package tui

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

type textDrawer struct {
	top, left, offset int
	region            region
	screen            tcell.Screen
}

func (d *textDrawer) setString(str string, style tcell.Style) {
	top := d.region.top + d.top
	left := d.region.left + d.left + d.offset
	right := d.region.left + d.region.width
	for _, c := range str {
		w := runewidth.RuneWidth(c)
		if left+w > right {
			break
		}
		if left+w == right && c != ' ' {
			if int(style)&int(tcell.AttrReverse) != 0 {
				d.screen.SetContent(left, top, ' ', nil, style)
			}
			break
		}
		d.screen.SetContent(left, top, c, nil, style)
		left += w
	}
}

func (d *textDrawer) setTop(top int) *textDrawer {
	d.top = top
	return d
}

func (d *textDrawer) setLeft(left int) *textDrawer {
	d.left = left
	return d
}

func (d *textDrawer) setOffset(offset int) *textDrawer {
	d.offset = offset
	return d
}
