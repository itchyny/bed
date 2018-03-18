package common

import "github.com/itchyny/bed/util"

// Layout represents the window layout.
type Layout interface {
	isLayout()
	Collect() map[int]LayoutWindow
	Replace(int) Layout
	Resize(int, int, int, int) Layout
	LeftMargin() int
	TopMargin() int
	Width() int
	Height() int
	SplitTop(int) Layout
	SplitBottom(int) Layout
	SplitLeft(int) Layout
	SplitRight(int) Layout
	Count() (int, int)
	Activate() Layout
	ActiveWindow() LayoutWindow
	Lookup(func(LayoutWindow) bool) LayoutWindow
	Close() Layout
}

// LayoutWindow holds the window index and it is active or not.
type LayoutWindow struct {
	Index  int
	Active bool
	left   int
	top    int
	width  int
	height int
}

// NewLayout creates a new Layout from a window index.
func NewLayout(index int) Layout {
	return LayoutWindow{Index: index, Active: true}
}

func (l LayoutWindow) isLayout() {}

// Collect returns all the LayoutWindow.
func (l LayoutWindow) Collect() map[int]LayoutWindow {
	return map[int]LayoutWindow{l.Index: l}
}

// Replace the active window with new window index.
func (l LayoutWindow) Replace(index int) Layout {
	if l.Active {
		l.Index = index
	}
	return l
}

// Resize recalculates the position.
func (l LayoutWindow) Resize(left, top, width, height int) Layout {
	l.left, l.top, l.width, l.height = left, top, width, height
	return l
}

// LeftMargin returns the left margin.
func (l LayoutWindow) LeftMargin() int {
	return l.left
}

// TopMargin returns the top margin.
func (l LayoutWindow) TopMargin() int {
	return l.top
}

// Width returns the width.
func (l LayoutWindow) Width() int {
	return l.width
}

// Height returns the height.
func (l LayoutWindow) Height() int {
	return l.height
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

// ActiveWindow returns the active window.
func (l LayoutWindow) ActiveWindow() LayoutWindow {
	if l.Active {
		return l
	}
	return LayoutWindow{Index: -1}
}

// Lookup search for the window meets the condition.
func (l LayoutWindow) Lookup(cond func(LayoutWindow) bool) LayoutWindow {
	if cond(l) {
		return l
	}
	return LayoutWindow{Index: -1}
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
	left   int
	top    int
	width  int
	height int
}

func (l LayoutHorizontal) isLayout() {}

// Collect returns all the LayoutWindow.
func (l LayoutHorizontal) Collect() map[int]LayoutWindow {
	m := l.Top.Collect()
	for i, l := range l.Bottom.Collect() {
		m[i] = l
	}
	return m
}

// Replace the active window with new window index.
func (l LayoutHorizontal) Replace(index int) Layout {
	return LayoutHorizontal{
		Top:    l.Top.Replace(index),
		Bottom: l.Bottom.Replace(index),
		width:  l.width,
		height: l.height,
	}
}

// Resize recalculates the position.
func (l LayoutHorizontal) Resize(left, top, width, height int) Layout {
	_, h1 := l.Top.Count()
	_, h2 := l.Bottom.Count()
	topHeight := height * h1 / (h1 + h2)
	return LayoutHorizontal{
		Top:    l.Top.Resize(left, top, width, topHeight),
		Bottom: l.Bottom.Resize(left, top+topHeight, width, height-topHeight),
		left:   left,
		top:    top,
		width:  width,
		height: height,
	}
}

// LeftMargin returns the left margin.
func (l LayoutHorizontal) LeftMargin() int {
	return l.left
}

// TopMargin returns the top margin.
func (l LayoutHorizontal) TopMargin() int {
	return l.top
}

// Width returns the width.
func (l LayoutHorizontal) Width() int {
	return l.width
}

// Height returns the height.
func (l LayoutHorizontal) Height() int {
	return l.height
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

// ActiveWindow returns the active window.
func (l LayoutHorizontal) ActiveWindow() LayoutWindow {
	if layout := l.Top.ActiveWindow(); layout.Index >= 0 {
		return layout
	}
	return l.Bottom.ActiveWindow()
}

// Lookup search for the window meets the condition.
func (l LayoutHorizontal) Lookup(cond func(LayoutWindow) bool) LayoutWindow {
	if layout := l.Top.Lookup(cond); layout.Index >= 0 {
		return layout
	}
	return l.Bottom.Lookup(cond)
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
	Left   Layout
	Right  Layout
	left   int
	top    int
	width  int
	height int
}

func (l LayoutVertical) isLayout() {}

// Collect returns all the LayoutWindow.
func (l LayoutVertical) Collect() map[int]LayoutWindow {
	m := l.Left.Collect()
	for i, l := range l.Right.Collect() {
		m[i] = l
	}
	return m
}

// Replace the active window with new window index.
func (l LayoutVertical) Replace(index int) Layout {
	return LayoutVertical{
		Left:   l.Left.Replace(index),
		Right:  l.Right.Replace(index),
		width:  l.width,
		height: l.height,
	}
}

// Resize recalculates the position.
func (l LayoutVertical) Resize(left, top, width, height int) Layout {
	w1, _ := l.Left.Count()
	w2, _ := l.Right.Count()
	leftWidth := width * w1 / (w1 + w2)
	return LayoutVertical{
		Left: l.Left.Resize(left, top, leftWidth, height),
		Right: l.Right.Resize(
			util.MinInt(left+leftWidth+1, width), top,
			util.MaxInt(width-leftWidth-1, 0), height),
		left:   left,
		top:    top,
		width:  width,
		height: height,
	}
}

// LeftMargin returns the left margin.
func (l LayoutVertical) LeftMargin() int {
	return l.left
}

// TopMargin returns the top margin.
func (l LayoutVertical) TopMargin() int {
	return l.top
}

// Width returns the width.
func (l LayoutVertical) Width() int {
	return l.width
}

// Height returns the height.
func (l LayoutVertical) Height() int {
	return l.height
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

// ActiveWindow returns the active window.
func (l LayoutVertical) ActiveWindow() LayoutWindow {
	if layout := l.Left.ActiveWindow(); layout.Index >= 0 {
		return layout
	}
	return l.Right.ActiveWindow()
}

// Lookup search for the window meets the condition.
func (l LayoutVertical) Lookup(cond func(LayoutWindow) bool) LayoutWindow {
	if layout := l.Left.Lookup(cond); layout.Index >= 0 {
		return layout
	}
	return l.Right.Lookup(cond)
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
