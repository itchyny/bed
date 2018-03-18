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

	expected := LayoutHorizontal{
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

	if !reflect.DeepEqual(layout.Indices(), []int{2, 3, 4, 1, 0}) {
		t.Errorf("Indices should be %+v but got %+v", []int{0}, layout.Indices())
	}

	layout = layout.Resize(15, 15)

	expected = LayoutHorizontal{
		Top: LayoutVertical{
			Left: LayoutHorizontal{
				Top: LayoutWindow{Index: 2, Active: false, width: 10, height: 5},
				Bottom: LayoutVertical{
					Left:   LayoutWindow{Index: 3, Active: false, width: 5, height: 5},
					Right:  LayoutWindow{Index: 4, Active: true, width: 4, height: 5},
					width:  10,
					height: 5,
				},
				width:  10,
				height: 10,
			},
			Right:  LayoutWindow{Index: 1, Active: false, width: 4, height: 10},
			width:  15,
			height: 10,
		},
		Bottom: LayoutWindow{Index: 0, Active: false, width: 15, height: 5},
		width:  15,
		height: 15,
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
	}

	expectedWindow := LayoutWindow{Index: 1, Active: false, width: 4, height: 10}
	if !reflect.DeepEqual(layout.Lookup(1), expectedWindow) {
		t.Errorf("Lookup(1) should be %+v but got %+v", expectedWindow, layout.Lookup(1))
	}

	expectedWindow = LayoutWindow{Index: 3, Active: false, width: 5, height: 5}
	if !reflect.DeepEqual(layout.Lookup(3), expectedWindow) {
		t.Errorf("Lookup(3) should be %+v but got %+v", expectedWindow, layout.Lookup(3))
	}

	expectedWindow = LayoutWindow{Index: -1}
	if !reflect.DeepEqual(layout.Lookup(5), expectedWindow) {
		t.Errorf("Lookup(5) should be %+v but got %+v", expectedWindow, layout.Lookup(5))
	}

	layout = layout.Close().Resize(0, 0)

	expected = LayoutHorizontal{
		Top: LayoutVertical{
			Left: LayoutHorizontal{
				Top:    LayoutWindow{Index: 2, Active: false},
				Bottom: LayoutWindow{Index: 3, Active: true},
			},
			Right: LayoutWindow{Index: 1, Active: false},
		},
		Bottom: LayoutWindow{Index: 0, Active: false},
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
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
				Top:    LayoutWindow{Index: 2, Active: false},
				Bottom: LayoutWindow{Index: 5, Active: true},
			},
			Right: LayoutWindow{Index: 1, Active: false},
		},
		Bottom: LayoutWindow{Index: 0, Active: false},
	}

	if !reflect.DeepEqual(layout, expected) {
		t.Errorf("layout should be %+v but got %+v", expected, layout)
	}
}
