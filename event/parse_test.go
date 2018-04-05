package event

import (
	"reflect"
	"testing"
)

func TestParseRange(t *testing.T) {
	testCases := []struct {
		target   string
		expected *Range
		index    int
	}{
		{" $ ", &Range{End{}, nil}, 3},
		{"32", &Range{Absolute{32}, nil}, 2},
		{"1024,4096", &Range{Absolute{1024}, Absolute{4096}}, 9},
		{"'<", &Range{VisualStart{}, nil}, 2},
		{"'>", &Range{VisualEnd{}, nil}, 2},
		{" '<  ,  '>  write", &Range{VisualStart{}, VisualEnd{}}, 12},
	}
	for _, testCase := range testCases {
		got, gotIndex := ParseRange([]rune(testCase.target), 0)
		if !reflect.DeepEqual(got, testCase.expected) {
			t.Errorf("ParseRange(%q) should return %+v but got %+v", testCase.target, testCase.expected, got)
		}
		if gotIndex != testCase.index {
			t.Errorf("ParseRange(%q) should return index %d but got %d", testCase.target, testCase.index, gotIndex)
		}
	}
}

func TestParsePos(t *testing.T) {
	testCases := []struct {
		target   string
		expected Position
		index    int
	}{
		{" $ ", End{}, 3},
		{"32", Absolute{32}, 2},
		{"1024,4096", Absolute{1024}, 4},
		{"0xffff", Absolute{65535}, 6},
		{"+16777216", Relative{16777216}, 9},
		{"-0xabcdef", Relative{-0xabcdef}, 9},
		{"'<", VisualStart{}, 2},
		{"'>", VisualEnd{}, 2},
		{" '<  ,  '> ", VisualStart{}, 5},
	}
	for _, testCase := range testCases {
		got, gotIndex := ParsePos([]rune(testCase.target), 0)
		if !reflect.DeepEqual(got, testCase.expected) {
			t.Errorf("ParsePos(%q) should return %#v but got %#v", testCase.target, testCase.expected, got)
		}
		if gotIndex != testCase.index {
			t.Errorf("ParsePos(%q) should return index %d but got %d", testCase.target, testCase.index, gotIndex)
		}
	}
}
