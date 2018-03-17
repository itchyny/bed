package tui

type region struct {
	top, left, height, width int
}

func (r region) splitHorizontally(h1, h2 int) []region {
	height := r.height * h1 / (h1 + h2)
	return []region{
		region{
			top:    r.top,
			left:   r.left,
			height: height,
			width:  r.width,
		},
		region{
			top:    r.top + height,
			left:   r.left,
			height: r.height - height,
			width:  r.width,
		},
	}
}

func (r region) splitVertically(w1, w2 int) []region {
	width := r.width * w1 / (w1 + w2)
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

func (r region) valid() bool {
	return 0 <= r.top && 0 <= r.left && 0 < r.height && 0 < r.width
}
