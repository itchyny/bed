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
		{" $-72 , $-36 ", &Range{End{-72}, End{-36}}, 13},
		{"32", &Range{Absolute{32}, nil}, 2},
		{"1024,4096", &Range{Absolute{1024}, Absolute{4096}}, 9},
		{"1+2+3+4+5+6+7+8+9,0xa+0xb+0xc+0xd+0xe+0xf", &Range{Absolute{45}, Absolute{75}}, 41},
		{"'<", &Range{VisualStart{}, nil}, 2},
		{"'>", &Range{VisualEnd{}, nil}, 2},
		{" '<  ,  '>  write", &Range{VisualStart{}, VisualEnd{}}, 12},
		{" '<+0x10 ,  '>-10 ", &Range{VisualStart{0x10}, VisualEnd{-10}}, 18},
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
		{" $-0x72 ", End{-0x72}, 8},
		{"32", Absolute{32}, 2},
		{"1024,4096", Absolute{1024}, 4},
		{"1+2+3+4+5+6+7+8+9+0xa+0xb+0xc+0xd+0xe+0xf", Absolute{120}, 41},
		{"0xffff", Absolute{65535}, 6},
		{"+16777216", Relative{16777216}, 9},
		{"-0xabcdef", Relative{-0xabcdef}, 9},
		{"+10+20+30-40", Relative{20}, 12},
		{"'<", VisualStart{}, 2},
		{"'>", VisualEnd{}, 2},
		{" '<  ,  '> ", VisualStart{}, 5},
		{" '<+0x10 ,  '>-10 ", VisualStart{0x10}, 9},
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
