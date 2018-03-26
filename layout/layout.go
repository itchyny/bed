package layout

import "github.com/itchyny/bed/mathutil"

// Layout represents the window layout.
type Layout interface {
	isLayout()
	Collect() map[int]Window
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
	Activate(int) Layout
	ActivateFirst() Layout
	ActiveWindow() Window
	Lookup(func(Window) bool) Window
	Close() Layout
}

// Window holds the window index and it is active or not.
type Window struct {
	Index  int
	Active bool
	left   int
	top    int
	width  int
	height int
}

// NewLayout creates a new Layout from a window index.
func NewLayout(index int) Layout {
	return Window{Index: index, Active: true}
}

func (l Window) isLayout() {}

// Collect returns all the Window.
func (l Window) Collect() map[int]Window {
	return map[int]Window{l.Index: l}
}

// Replace the active window with new window index.
func (l Window) Replace(index int) Layout {
	if l.Active {
		l.Index = index
	}
	return l
}

// Resize recalculates the position.
func (l Window) Resize(left, top, width, height int) Layout {
	l.left, l.top, l.width, l.height = left, top, width, height
	return l
}

// LeftMargin returns the left margin.
func (l Window) LeftMargin() int {
	return l.left
}

// TopMargin returns the top margin.
func (l Window) TopMargin() int {
	return l.top
}

// Width returns the width.
func (l Window) Width() int {
	return l.width
}

// Height returns the height.
func (l Window) Height() int {
	return l.height
}

// SplitTop splits the layout and opens a new window to the top.
func (l Window) SplitTop(index int) Layout {
	if !l.Active {
		return l
	}
	return Horizontal{
		Top:    Window{Index: index, Active: true},
		Bottom: Window{Index: l.Index, Active: false},
	}
}

// SplitBottom splits the layout and opens a new window to the bottom.
func (l Window) SplitBottom(index int) Layout {
	if !l.Active {
		return l
	}
	return Horizontal{
		Top:    Window{Index: l.Index, Active: false},
		Bottom: Window{Index: index, Active: true},
	}
}

// SplitLeft splits the layout and opens a new window to the left.
func (l Window) SplitLeft(index int) Layout {
	if !l.Active {
		return l
	}
	return Vertical{
		Left:  Window{Index: index, Active: true},
		Right: Window{Index: l.Index, Active: false},
	}
}

// SplitRight splits the layout and opens a new window to the right.
func (l Window) SplitRight(index int) Layout {
	if !l.Active {
		return l
	}
	return Vertical{
		Left:  Window{Index: l.Index, Active: false},
		Right: Window{Index: index, Active: true},
	}
}

// Count returns the width and height counts.
func (l Window) Count() (int, int) {
	return 1, 1
}

// Activate the specific window layout.
func (l Window) Activate(i int) Layout {
	l.Active = l.Index == i
	return l
}

// ActivateFirst the first layout.
func (l Window) ActivateFirst() Layout {
	l.Active = true
	return l
}

// ActiveWindow returns the active window.
func (l Window) ActiveWindow() Window {
	if l.Active {
		return l
	}
	return Window{Index: -1}
}

// Lookup search for the window meets the condition.
func (l Window) Lookup(cond func(Window) bool) Window {
	if cond(l) {
		return l
	}
	return Window{Index: -1}
}

// Close the active layout.
func (l Window) Close() Layout {
	if l.Active {
		panic("Active Window should not be closed")
	}
	return l
}

// Horizontal holds two layout horizontally.
type Horizontal struct {
	Top    Layout
	Bottom Layout
	left   int
	top    int
	width  int
	height int
}

func (l Horizontal) isLayout() {}

// Collect returns all the Window.
func (l Horizontal) Collect() map[int]Window {
	m := l.Top.Collect()
	for i, l := range l.Bottom.Collect() {
		m[i] = l
	}
	return m
}

// Replace the active window with new window index.
func (l Horizontal) Replace(index int) Layout {
	return Horizontal{
		Top:    l.Top.Replace(index),
		Bottom: l.Bottom.Replace(index),
		left:   l.left,
		top:    l.top,
		width:  l.width,
		height: l.height,
	}
}

// Resize recalculates the position.
func (l Horizontal) Resize(left, top, width, height int) Layout {
	_, h1 := l.Top.Count()
	_, h2 := l.Bottom.Count()
	topHeight := height * h1 / (h1 + h2)
	return Horizontal{
		Top:    l.Top.Resize(left, top, width, topHeight),
		Bottom: l.Bottom.Resize(left, top+topHeight, width, height-topHeight),
		left:   left,
		top:    top,
		width:  width,
		height: height,
	}
}

// LeftMargin returns the left margin.
func (l Horizontal) LeftMargin() int {
	return l.left
}

// TopMargin returns the top margin.
func (l Horizontal) TopMargin() int {
	return l.top
}

// Width returns the width.
func (l Horizontal) Width() int {
	return l.width
}

// Height returns the height.
func (l Horizontal) Height() int {
	return l.height
}

// SplitTop splits the layout and opens a new window to the top.
func (l Horizontal) SplitTop(index int) Layout {
	return Horizontal{
		Top:    l.Top.SplitTop(index),
		Bottom: l.Bottom.SplitTop(index),
	}
}

// SplitBottom splits the layout and opens a new window to the bottom.
func (l Horizontal) SplitBottom(index int) Layout {
	return Horizontal{
		Top:    l.Top.SplitBottom(index),
		Bottom: l.Bottom.SplitBottom(index),
	}
}

// SplitLeft splits the layout and opens a new window to the left.
func (l Horizontal) SplitLeft(index int) Layout {
	return Horizontal{
		Top:    l.Top.SplitLeft(index),
		Bottom: l.Bottom.SplitLeft(index),
	}
}

// SplitRight splits the layout and opens a new window to the right.
func (l Horizontal) SplitRight(index int) Layout {
	return Horizontal{
		Top:    l.Top.SplitRight(index),
		Bottom: l.Bottom.SplitRight(index),
	}
}

// Count returns the width and height counts.
func (l Horizontal) Count() (int, int) {
	w1, h1 := l.Top.Count()
	w2, h2 := l.Bottom.Count()
	return mathutil.MaxInt(w1, w2), h1 + h2
}

// Activate the specific window layout.
func (l Horizontal) Activate(i int) Layout {
	return Horizontal{
		Top:    l.Top.Activate(i),
		Bottom: l.Bottom.Activate(i),
		left:   l.left,
		top:    l.top,
		width:  l.width,
		height: l.height,
	}
}

// ActivateFirst the first layout.
func (l Horizontal) ActivateFirst() Layout {
	return Horizontal{
		Top:    l.Top.ActivateFirst(),
		Bottom: l.Bottom,
		left:   l.left,
		top:    l.top,
		width:  l.width,
		height: l.height,
	}
}

// ActiveWindow returns the active window.
func (l Horizontal) ActiveWindow() Window {
	if layout := l.Top.ActiveWindow(); layout.Index >= 0 {
		return layout
	}
	return l.Bottom.ActiveWindow()
}

// Lookup search for the window meets the condition.
func (l Horizontal) Lookup(cond func(Window) bool) Window {
	if layout := l.Top.Lookup(cond); layout.Index >= 0 {
		return layout
	}
	return l.Bottom.Lookup(cond)
}

// Close the active layout.
func (l Horizontal) Close() Layout {
	switch m := l.Top.(type) {
	case Window:
		if m.Active {
			return l.Bottom.ActivateFirst()
		}
	}
	switch m := l.Bottom.(type) {
	case Window:
		if m.Active {
			return l.Top.ActivateFirst()
		}
	}
	return Horizontal{
		Top:    l.Top.Close(),
		Bottom: l.Bottom.Close(),
	}
}

// Vertical holds two layout vertically.
type Vertical struct {
	Left   Layout
	Right  Layout
	left   int
	top    int
	width  int
	height int
}

func (l Vertical) isLayout() {}

// Collect returns all the Window.
func (l Vertical) Collect() map[int]Window {
	m := l.Left.Collect()
	for i, l := range l.Right.Collect() {
		m[i] = l
	}
	return m
}

// Replace the active window with new window index.
func (l Vertical) Replace(index int) Layout {
	return Vertical{
		Left:   l.Left.Replace(index),
		Right:  l.Right.Replace(index),
		left:   l.left,
		top:    l.top,
		width:  l.width,
		height: l.height,
	}
}

// Resize recalculates the position.
func (l Vertical) Resize(left, top, width, height int) Layout {
	w1, _ := l.Left.Count()
	w2, _ := l.Right.Count()
	leftWidth := width * w1 / (w1 + w2)
	return Vertical{
		Left: l.Left.Resize(left, top, leftWidth, height),
		Right: l.Right.Resize(
			mathutil.MinInt(left+leftWidth+1, left+width), top,
			mathutil.MaxInt(width-leftWidth-1, 0), height),
		left:   left,
		top:    top,
		width:  width,
		height: height,
	}
}

// LeftMargin returns the left margin.
func (l Vertical) LeftMargin() int {
	return l.left
}

// TopMargin returns the top margin.
func (l Vertical) TopMargin() int {
	return l.top
}

// Width returns the width.
func (l Vertical) Width() int {
	return l.width
}

// Height returns the height.
func (l Vertical) Height() int {
	return l.height
}

// SplitTop splits the layout and opens a new window to the top.
func (l Vertical) SplitTop(index int) Layout {
	return Vertical{
		Left:  l.Left.SplitTop(index),
		Right: l.Right.SplitTop(index),
	}
}

// SplitBottom splits the layout and opens a new window to the bottom.
func (l Vertical) SplitBottom(index int) Layout {
	return Vertical{
		Left:  l.Left.SplitBottom(index),
		Right: l.Right.SplitBottom(index),
	}
}

// SplitLeft splits the layout and opens a new window to the left.
func (l Vertical) SplitLeft(index int) Layout {
	return Vertical{
		Left:  l.Left.SplitLeft(index),
		Right: l.Right.SplitLeft(index),
	}
}

// SplitRight splits the layout and opens a new window to the right.
func (l Vertical) SplitRight(index int) Layout {
	return Vertical{
		Left:  l.Left.SplitRight(index),
		Right: l.Right.SplitRight(index),
	}
}

// Count returns the width and height counts.
func (l Vertical) Count() (int, int) {
	w1, h1 := l.Left.Count()
	w2, h2 := l.Right.Count()
	return w1 + w2, mathutil.MaxInt(h1, h2)
}

// Activate the specific window layout.
func (l Vertical) Activate(i int) Layout {
	return Vertical{
		Left:   l.Left.Activate(i),
		Right:  l.Right.Activate(i),
		left:   l.left,
		top:    l.top,
		width:  l.width,
		height: l.height,
	}
}

// ActivateFirst the first layout.
func (l Vertical) ActivateFirst() Layout {
	return Vertical{
		Left:   l.Left.ActivateFirst(),
		Right:  l.Right,
		left:   l.left,
		top:    l.top,
		width:  l.width,
		height: l.height,
	}
}

// ActiveWindow returns the active window.
func (l Vertical) ActiveWindow() Window {
	if layout := l.Left.ActiveWindow(); layout.Index >= 0 {
		return layout
	}
	return l.Right.ActiveWindow()
}

// Lookup search for the window meets the condition.
func (l Vertical) Lookup(cond func(Window) bool) Window {
	if layout := l.Left.Lookup(cond); layout.Index >= 0 {
		return layout
	}
	return l.Right.Lookup(cond)
}

// Close the active layout.
func (l Vertical) Close() Layout {
	switch m := l.Left.(type) {
	case Window:
		if m.Active {
			return l.Right.ActivateFirst()
		}
	}
	switch m := l.Right.(type) {
	case Window:
		if m.Active {
			return l.Left.ActivateFirst()
		}
	}
	return Vertical{
		Left:  l.Left.Close(),
		Right: l.Right.Close(),
	}
}
