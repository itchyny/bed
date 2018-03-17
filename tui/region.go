package tui

type region struct {
	top, left, height, width int
}

func (r region) splitHorizontally() []region {
	height := r.height / 2
	return []region{
		region{
			top:    r.top,
			left:   r.left,
			height: height,
			width:  r.width,
		},
		region{
			top:    r.top + height + 1,
			left:   r.left,
			height: r.height - height - 1,
			width:  r.width,
		},
	}
}

func (r region) splitVertically() []region {
	width := r.width / 2
	return []region{
		region{
			top:    r.top,
			left:   r.left,
			height: r.height,
			width:  width,
		},
		region{
			top:    r.top,
			left:   r.left + width + 1,
			height: r.height,
			width:  r.width - width - 1,
		},
	}
}
