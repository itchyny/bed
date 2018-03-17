package common

type Layout interface {
	isLayout()
	Indices() []int
	SplitTop(int) Layout
	SplitBottom(int) Layout
	SplitLeft(int) Layout
	SplitRight(int) Layout
}

type LayoutWindow struct {
	Index  int
	Active bool
}

func NewLayout(index int) Layout {
	return LayoutWindow{Index: index, Active: true}
}

func (_ LayoutWindow) isLayout() {}

func (l LayoutWindow) Indices() []int {
	return []int{l.Index}
}

func (l LayoutWindow) SplitTop(index int) Layout {
	if !l.Active {
		return l
	}
	return LayoutHorizontal{
		Top:    LayoutWindow{Index: index, Active: true},
		Bottom: LayoutWindow{Index: l.Index, Active: false},
	}
}

func (l LayoutWindow) SplitBottom(index int) Layout {
	if !l.Active {
		return l
	}
	return LayoutHorizontal{
		Top:    LayoutWindow{Index: l.Index, Active: false},
		Bottom: LayoutWindow{Index: index, Active: true},
	}
}

func (l LayoutWindow) SplitLeft(index int) Layout {
	if !l.Active {
		return l
	}
	return LayoutVertical{
		Left:  LayoutWindow{Index: index, Active: true},
		Right: LayoutWindow{Index: l.Index, Active: false},
	}
}

func (l LayoutWindow) SplitRight(index int) Layout {
	if !l.Active {
		return l
	}
	return LayoutVertical{
		Left:  LayoutWindow{Index: l.Index, Active: false},
		Right: LayoutWindow{Index: index, Active: true},
	}
}

type LayoutHorizontal struct {
	Top    Layout
	Bottom Layout
}

func (_ LayoutHorizontal) isLayout() {}

func (l LayoutHorizontal) Indices() []int {
	return append(l.Top.Indices(), l.Bottom.Indices()...)
}

func (l LayoutHorizontal) SplitTop(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.SplitTop(index),
		Bottom: l.Bottom.SplitTop(index),
	}
}

func (l LayoutHorizontal) SplitBottom(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.SplitBottom(index),
		Bottom: l.Bottom.SplitBottom(index),
	}
}

func (l LayoutHorizontal) SplitLeft(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.SplitLeft(index),
		Bottom: l.Bottom.SplitLeft(index),
	}
}

func (l LayoutHorizontal) SplitRight(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.SplitRight(index),
		Bottom: l.Bottom.SplitRight(index),
	}
}

type LayoutVertical struct {
	Left  Layout
	Right Layout
}

func (_ LayoutVertical) isLayout() {}

func (l LayoutVertical) Indices() []int {
	return append(l.Left.Indices(), l.Right.Indices()...)
}

func (l LayoutVertical) SplitTop(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.SplitTop(index),
		Right: l.Right.SplitTop(index),
	}
}

func (l LayoutVertical) SplitBottom(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.SplitBottom(index),
		Right: l.Right.SplitBottom(index),
	}
}

func (l LayoutVertical) SplitLeft(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.SplitLeft(index),
		Right: l.Right.SplitLeft(index),
	}
}

func (l LayoutVertical) SplitRight(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.SplitRight(index),
		Right: l.Right.SplitRight(index),
	}
}
