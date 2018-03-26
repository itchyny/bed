package tui

import "github.com/itchyny/bed/layout"

type region struct {
	left, top, height, width int
}

func fromLayout(l layout.Layout) region {
	return region{
		left:   l.LeftMargin(),
		top:    l.TopMargin(),
		height: l.Height(),
		width:  l.Width(),
	}
}

func (r region) valid() bool {
	return 0 <= r.left && 0 <= r.top && 0 < r.height && 0 < r.width
}
