package tui

type region struct {
	top, left, height, width int
}

func (r region) valid() bool {
	return 0 <= r.top && 0 <= r.left && 0 < r.height && 0 < r.width
}
