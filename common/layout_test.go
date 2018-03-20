package common

import (
	"reflect"
	"testing"
)

func TestLayout(t *testing.T) {
	layout := NewLayout(0)

	layout = layout.SplitTop(1)
	layout = layout.SplitLeft(2)
	layout = layout.SplitBottom(3)
	layout = layout.SplitRight(4)

	var expected Layout
	expected = LayoutHorizontal{
		Top: LayoutVertical{
			Left: LayoutHorizontal{
				Top: LayoutWindow{Index: 2, Active: false},
				Bottom: LayoutVertical{
					Left:  LayoutWindow{Index: 3, Active: false},
					Right: LayoutWindow{Index: 4, Active: true},
				},
			},
			Right: LayoutWindow{Index: 1, Active: false},
		},
		Bottom: LayoutWindow{Index: 0, Active: false},
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
	}

	w, h := layout.Count()
	if w != 3 {
		t.Errorf("layout width be %d but got %d", 3, w)
	}
	if h != 3 {
		t.Errorf("layout height be %d but got %d", 3, h)
	}

	layout = layout.Resize(0, 0, 15, 15)

	expected = LayoutHorizontal{
		Top: LayoutVertical{
			Left: LayoutHorizontal{
				Top: LayoutWindow{Index: 2, Active: false, left: 0, top: 0, width: 10, height: 5},
				Bottom: LayoutVertical{
					Left:   LayoutWindow{Index: 3, Active: false, left: 0, top: 5, width: 5, height: 5},
					Right:  LayoutWindow{Index: 4, Active: true, left: 6, top: 5, width: 4, height: 5},
					left:   0,
					top:    5,
					width:  10,
					height: 5,
				},
				left:   0,
				top:    0,
				width:  10,
				height: 10,
			},
			Right:  LayoutWindow{Index: 1, Active: false, left: 11, top: 0, width: 4, height: 10},
			left:   0,
			top:    0,
			width:  15,
			height: 10,
		},
		Bottom: LayoutWindow{Index: 0, Active: false, left: 0, top: 10, width: 15, height: 5},
		left:   0,
		top:    0,
		width:  15,
		height: 15,
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
	}

	expectedWindow := LayoutWindow{Index: 1, Active: false, left: 11, top: 0, width: 4, height: 10}
	got := layout.Lookup(func(l LayoutWindow) bool { return l.Index == 1 })
	if !reflect.DeepEqual(got, expectedWindow) {
		t.Errorf("Lookup(Index == 1) should be %+v but got %+v", expectedWindow, got)
	}

	if got.LeftMargin() != 11 {
		t.Errorf("LeftMargin() should be %+v but got %+v", 11, got.LeftMargin())
	}
	if got.TopMargin() != 0 {
		t.Errorf("TopMargin() should be %+v but got %+v", 0, got.TopMargin())
	}
	if got.Width() != 4 {
		t.Errorf("Width() should be %+v but got %+v", 4, got.Width())
	}
	if got.Height() != 10 {
		t.Errorf("Height() should be %+v but got %+v", 10, got.Height())
	}

	expectedWindow = LayoutWindow{Index: 3, Active: false, left: 0, top: 5, width: 5, height: 5}
	got = layout.Lookup(func(l LayoutWindow) bool { return l.Index == 3 })
	if !reflect.DeepEqual(got, expectedWindow) {
		t.Errorf("Lookup(Index == 3) should be %+v but got %+v", expectedWindow, got)
	}

	if got.LeftMargin() != 0 {
		t.Errorf("LeftMargin() should be %+v but got %+v", 0, got.LeftMargin())
	}
	if got.TopMargin() != 5 {
		t.Errorf("TopMargin() should be %+v but got %+v", 5, got.TopMargin())
	}
	if got.Width() != 5 {
		t.Errorf("Width() should be %+v but got %+v", 5, got.Width())
	}
	if got.Height() != 5 {
		t.Errorf("Height() should be %+v but got %+v", 5, got.Height())
	}

	expectedWindow = LayoutWindow{Index: -1}
	got = layout.Lookup(func(l LayoutWindow) bool { return l.Index == 5 })
	if !reflect.DeepEqual(got, expectedWindow) {
		t.Errorf("Lookup(Index == 5) should be %+v but got %+v", expectedWindow, got)
	}

	expectedWindow = LayoutWindow{Index: 4, Active: true, left: 6, top: 5, width: 4, height: 5}
	got = layout.ActiveWindow()
	if !reflect.DeepEqual(got, expectedWindow) {
		t.Errorf("ActiveWindow() should be %+v but got %+v", expectedWindow, got)
	}

	expectedMap := map[int]LayoutWindow{
		0: LayoutWindow{Index: 0, Active: false, left: 0, top: 10, width: 15, height: 5},
		1: LayoutWindow{Index: 1, Active: false, left: 11, top: 0, width: 4, height: 10},
		2: LayoutWindow{Index: 2, Active: false, left: 0, top: 0, width: 10, height: 5},
		3: LayoutWindow{Index: 3, Active: false, left: 0, top: 5, width: 5, height: 5},
		4: LayoutWindow{Index: 4, Active: true, left: 6, top: 5, width: 4, height: 5},
	}

	if !reflect.DeepEqual(layout.Collect(), expectedMap) {
		t.Errorf("Collect should be %+v but got %+v", expectedMap, layout.Collect())
	}

	layout = layout.Close().Resize(0, 0, 15, 15)

	expected = LayoutHorizontal{
		Top: LayoutVertical{
			Left: LayoutHorizontal{
				Top:    LayoutWindow{Index: 2, Active: false, left: 0, top: 0, width: 7, height: 5},
				Bottom: LayoutWindow{Index: 3, Active: true, left: 0, top: 5, width: 7, height: 5},
				left:   0,
				top:    0,
				width:  7,
				height: 10,
			},
			Right:  LayoutWindow{Index: 1, Active: false, left: 8, top: 0, width: 7, height: 10},
			left:   0,
			top:    0,
			width:  15,
			height: 10,
		},
		Bottom: LayoutWindow{Index: 0, Active: false, left: 0, top: 10, width: 15, height: 5},
		left:   0,
		top:    0,
		width:  15,
		height: 15,
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
	}

	if layout.LeftMargin() != 0 {
		t.Errorf("LeftMargin() should be %+v but layout %+v", 0, layout.LeftMargin())
	}
	if layout.TopMargin() != 0 {
		t.Errorf("TopMargin() should be %+v but layout %+v", 0, layout.TopMargin())
	}
	if layout.Width() != 15 {
		t.Errorf("Width() should be %+v but layout %+v", 15, layout.Width())
	}
	if layout.Height() != 15 {
		t.Errorf("Height() should be %+v but layout %+v", 15, layout.Height())
	}

	expectedMap = map[int]LayoutWindow{
		0: LayoutWindow{Index: 0, Active: false, left: 0, top: 10, width: 15, height: 5},
		1: LayoutWindow{Index: 1, Active: false, left: 8, top: 0, width: 7, height: 10},
		2: LayoutWindow{Index: 2, Active: false, left: 0, top: 0, width: 7, height: 5},
		3: LayoutWindow{Index: 3, Active: true, left: 0, top: 5, width: 7, height: 5},
	}

	if !reflect.DeepEqual(layout.Collect(), expectedMap) {
		t.Errorf("Collect should be %+v but got %+v", expectedMap, layout.Collect())
	}

	w, h = layout.Count()
	if w != 2 {
		t.Errorf("layout width be %d but got %d", 3, w)
	}
	if h != 3 {
		t.Errorf("layout height be %d but got %d", 3, h)
	}

	layout = layout.Replace(5)

	expected = LayoutHorizontal{
		Top: LayoutVertical{
			Left: LayoutHorizontal{
				Top:    LayoutWindow{Index: 2, Active: false, left: 0, top: 0, width: 7, height: 5},
				Bottom: LayoutWindow{Index: 5, Active: true, left: 0, top: 5, width: 7, height: 5},
				left:   0,
				top:    0,
				width:  7,
				height: 10,
			},
			Right:  LayoutWindow{Index: 1, Active: false, left: 8, top: 0, width: 7, height: 10},
			left:   0,
			top:    0,
			width:  15,
			height: 10,
		},
		Bottom: LayoutWindow{Index: 0, Active: false, left: 0, top: 10, width: 15, height: 5},
		left:   0,
		top:    0,
		width:  15,
		height: 15,
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
	}

	layout = layout.Activate(1)

	expected = LayoutHorizontal{
		Top: LayoutVertical{
			Left: LayoutHorizontal{
				Top:    LayoutWindow{Index: 2, Active: false, left: 0, top: 0, width: 7, height: 5},
				Bottom: LayoutWindow{Index: 5, Active: false, left: 0, top: 5, width: 7, height: 5},
				left:   0,
				top:    0,
				width:  7,
				height: 10,
			},
			Right:  LayoutWindow{Index: 1, Active: true, left: 8, top: 0, width: 7, height: 10},
			left:   0,
			top:    0,
			width:  15,
			height: 10,
		},
		Bottom: LayoutWindow{Index: 0, Active: false, left: 0, top: 10, width: 15, height: 5},
		left:   0,
		top:    0,
		width:  15,
		height: 15,
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
	}

	layout = LayoutVertical{
		Left:  LayoutWindow{Index: 6, Active: false},
		Right: layout,
	}.SplitLeft(7).SplitTop(8).Resize(0, 0, 15, 10)

	expected = LayoutVertical{
		Left: LayoutWindow{Index: 6, Active: false, left: 0, top: 0, width: 3, height: 10},
		Right: LayoutHorizontal{
			Top: LayoutVertical{
				Left: LayoutHorizontal{
					Top:    LayoutWindow{Index: 2, Active: false, left: 4, top: 0, width: 3, height: 3},
					Bottom: LayoutWindow{Index: 5, Active: false, left: 4, top: 3, width: 3, height: 3},
					left:   4, top: 0, width: 3, height: 6,
				},
				Right: LayoutVertical{
					Left: LayoutHorizontal{
						Top:    LayoutWindow{Index: 8, Active: true, left: 8, top: 0, width: 3, height: 3},
						Bottom: LayoutWindow{Index: 7, Active: false, left: 8, top: 3, width: 3, height: 3},
						left:   8, top: 0, width: 3, height: 6,
					},
					Right: LayoutWindow{Index: 1, Active: false, left: 12, top: 0, width: 3, height: 6},
					left:  8, top: 0, width: 7, height: 6,
				},
				left: 4, top: 0, width: 11, height: 6,
			},
			Bottom: LayoutWindow{Index: 0, Active: false, left: 4, top: 6, width: 11, height: 4},
			left:   4, top: 0, width: 11, height: 10,
		},
		left: 0, top: 0, width: 15, height: 10,
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
	}

	if layout.LeftMargin() != 0 {
		t.Errorf("LeftMargin() should be %+v but layout %+v", 0, layout.LeftMargin())
	}
	if layout.TopMargin() != 0 {
		t.Errorf("TopMargin() should be %+v but layout %+v", 0, layout.TopMargin())
	}
	if layout.Width() != 15 {
		t.Errorf("Width() should be %+v but layout %+v", 15, layout.Width())
	}
	if layout.Height() != 10 {
		t.Errorf("Height() should be %+v but layout %+v", 10, layout.Height())
	}
}
