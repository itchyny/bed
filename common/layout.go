package common

type Layout interface {
	isLayout()
	Indices() []int
}

type LayoutWindow struct {
	Index int
}

func (_ LayoutWindow) isLayout() {}

func (l LayoutWindow) Indices() []int {
	return []int{l.Index}
}

type LayoutHorizontal struct {
	Top    Layout
	Bottom Layout
}

func (_ LayoutHorizontal) isLayout() {}

func (l LayoutHorizontal) Indices() []int {
	return append(l.Top.Indices(), l.Bottom.Indices()...)
}

type LayoutVertical struct {
	Left  Layout
	Right Layout
}

func (_ LayoutVertical) isLayout() {}

func (l LayoutVertical) Indices() []int {
	return append(l.Left.Indices(), l.Right.Indices()...)
}
