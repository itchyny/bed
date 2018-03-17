package common

import "github.com/itchyny/bed/util"

// Layout represents the window layout.
type Layout interface {
	isLayout()
	Indices() []int
	Replace(int) Layout
	SplitTop(int) Layout
	SplitBottom(int) Layout
	SplitLeft(int) Layout
	SplitRight(int) Layout
	Count() (int, int)
	Activate() Layout
	ActiveIndex() int
	Close() Layout
}

// LayoutWindow holds the window index and it is active or not.
type LayoutWindow struct {
	Index  int
	Active bool
}

// NewLayout creates a new Layout from a window index.
func NewLayout(index int) Layout {
	return LayoutWindow{Index: index, Active: true}
}

func (l LayoutWindow) isLayout() {}

// Indices returns all the window indeces.
func (l LayoutWindow) Indices() []int {
	return []int{l.Index}
}

// Replace the active window with new window index.
func (l LayoutWindow) Replace(index int) Layout {
	if l.Active {
		return NewLayout(index)
	}
	return l
}

// SplitTop splits the layout and opens a new window to the top.
func (l LayoutWindow) SplitTop(index int) Layout {
	if !l.Active {
		return l
	}
	return LayoutHorizontal{
		Top:    LayoutWindow{Index: index, Active: true},
		Bottom: LayoutWindow{Index: l.Index, Active: false},
	}
}

// SplitBottom splits the layout and opens a new window to the bottom.
func (l LayoutWindow) SplitBottom(index int) Layout {
	if !l.Active {
		return l
	}
	return LayoutHorizontal{
		Top:    LayoutWindow{Index: l.Index, Active: false},
		Bottom: LayoutWindow{Index: index, Active: true},
	}
}

// SplitLeft splits the layout and opens a new window to the left.
func (l LayoutWindow) SplitLeft(index int) Layout {
	if !l.Active {
		return l
	}
	return LayoutVertical{
		Left:  LayoutWindow{Index: index, Active: true},
		Right: LayoutWindow{Index: l.Index, Active: false},
	}
}

// SplitRight splits the layout and opens a new window to the right.
func (l LayoutWindow) SplitRight(index int) Layout {
	if !l.Active {
		return l
	}
	return LayoutVertical{
		Left:  LayoutWindow{Index: l.Index, Active: false},
		Right: LayoutWindow{Index: index, Active: true},
	}
}

// Count returns the width and height counts.
func (l LayoutWindow) Count() (int, int) {
	return 1, 1
}

// Activate the first layout.
func (l LayoutWindow) Activate() Layout {
	l.Active = true
	return l
}

// ActiveIndex returns the active window index.
func (l LayoutWindow) ActiveIndex() int {
	if l.Active {
		return l.Index
	}
	return -1
}

// Close the active layout.
func (l LayoutWindow) Close() Layout {
	if l.Active {
		panic("Active LayoutWindow should not be closed")
	}
	return l
}

// LayoutHorizontal holds two layout horizontally.
type LayoutHorizontal struct {
	Top    Layout
	Bottom Layout
}

func (l LayoutHorizontal) isLayout() {}

// Indices returns all the window indeces.
func (l LayoutHorizontal) Indices() []int {
	return append(l.Top.Indices(), l.Bottom.Indices()...)
}

// Replace the active window with new window index.
func (l LayoutHorizontal) Replace(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.Replace(index),
		Bottom: l.Bottom.Replace(index),
	}
}

// SplitTop splits the layout and opens a new window to the top.
func (l LayoutHorizontal) SplitTop(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.SplitTop(index),
		Bottom: l.Bottom.SplitTop(index),
	}
}

// SplitBottom splits the layout and opens a new window to the bottom.
func (l LayoutHorizontal) SplitBottom(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.SplitBottom(index),
		Bottom: l.Bottom.SplitBottom(index),
	}
}

// SplitLeft splits the layout and opens a new window to the left.
func (l LayoutHorizontal) SplitLeft(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.SplitLeft(index),
		Bottom: l.Bottom.SplitLeft(index),
	}
}

// SplitRight splits the layout and opens a new window to the right.
func (l LayoutHorizontal) SplitRight(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.SplitRight(index),
		Bottom: l.Bottom.SplitRight(index),
	}
}

// Count returns the width and height counts.
func (l LayoutHorizontal) Count() (int, int) {
	w1, h1 := l.Top.Count()
	w2, h2 := l.Bottom.Count()
	return util.MaxInt(w1, w2), h1 + h2
}

// Activate the first layout.
func (l LayoutHorizontal) Activate() Layout {
	return LayoutHorizontal{
		Top:    l.Top.Activate(),
		Bottom: l.Bottom,
	}
}

// ActiveIndex returns the active window index.
func (l LayoutHorizontal) ActiveIndex() int {
	if index := l.Top.ActiveIndex(); index >= 0 {
		return index
	}
	return l.Bottom.ActiveIndex()
}

// Close the active layout.
func (l LayoutHorizontal) Close() Layout {
	switch m := l.Top.(type) {
	case LayoutWindow:
		if m.Active {
			return l.Bottom.Activate()
		}
	}
	switch m := l.Bottom.(type) {
	case LayoutWindow:
		if m.Active {
			return l.Top.Activate()
		}
	}
	return LayoutHorizontal{
		Top:    l.Top.Close(),
		Bottom: l.Bottom.Close(),
	}
}

// LayoutVertical holds two layout vertically.
type LayoutVertical struct {
	Left  Layout
	Right Layout
}

func (l LayoutVertical) isLayout() {}

// Indices returns all the window indeces.
func (l LayoutVertical) Indices() []int {
	return append(l.Left.Indices(), l.Right.Indices()...)
}

// Replace the active window with new window index.
func (l LayoutVertical) Replace(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.Replace(index),
		Right: l.Right.Replace(index),
	}
}

// SplitTop splits the layout and opens a new window to the top.
func (l LayoutVertical) SplitTop(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.SplitTop(index),
		Right: l.Right.SplitTop(index),
	}
}

// SplitBottom splits the layout and opens a new window to the bottom.
func (l LayoutVertical) SplitBottom(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.SplitBottom(index),
		Right: l.Right.SplitBottom(index),
	}
}

// SplitLeft splits the layout and opens a new window to the left.
func (l LayoutVertical) SplitLeft(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.SplitLeft(index),
		Right: l.Right.SplitLeft(index),
	}
}

// SplitRight splits the layout and opens a new window to the right.
func (l LayoutVertical) SplitRight(index int) Layout {
	return LayoutVertical{
		Left:  l.Left.SplitRight(index),
		Right: l.Right.SplitRight(index),
	}
}

// Count returns the width and height counts.
func (l LayoutVertical) Count() (int, int) {
	w1, h1 := l.Left.Count()
	w2, h2 := l.Right.Count()
	return w1 + w2, util.MaxInt(h1, h2)
}

// Activate the first layout.
func (l LayoutVertical) Activate() Layout {
	return LayoutVertical{
		Left:  l.Left.Activate(),
		Right: l.Right,
	}
}

// ActiveIndex returns the active window index.
func (l LayoutVertical) ActiveIndex() int {
	if index := l.Left.ActiveIndex(); index >= 0 {
		return index
	}
	return l.Right.ActiveIndex()
}

// Close the active layout.
func (l LayoutVertical) Close() Layout {
	switch m := l.Left.(type) {
	case LayoutWindow:
		if m.Active {
			return l.Right.Activate()
		}
	}
	switch m := l.Right.(type) {
	case LayoutWindow:
		if m.Active {
			return l.Left.Activate()
		}
	}
	return LayoutVertical{
		Left:  l.Left.Close(),
		Right: l.Right.Close(),
	}
}
