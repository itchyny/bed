package event

import (
	"reflect"
	"testing"
)

func TestParseRange(t *testing.T) {
	testCases := []struct {
		target   string
		expected *Range
		rest     string
	}{
		{"", nil, ""},
		{"e", nil, "e"},
		{"   ", nil, ""},
		{"$", &Range{End{}, nil}, ""},
		{" $-72 , $-36 ", &Range{End{-72}, End{-36}}, ""},
		{"32", &Range{Absolute{32}, nil}, ""},
		{"+32", &Range{Relative{32}, nil}, ""},
		{"-32", &Range{Relative{-32}, nil}, ""},
		{"1024,4096", &Range{Absolute{1024}, Absolute{4096}}, ""},
		{"1+2+3+4+5+6+7+8+9,0xa+0xb+0xc+0xD+0xE+0xF", &Range{Absolute{45}, Absolute{75}}, ""},
		{"10d", &Range{Absolute{10}, nil}, "d"},
		{"0x12G", &Range{Absolute{0x12}, nil}, "G"},
		{"0x10fag", &Range{Absolute{0x10fa}, nil}, "g"},
		{".-100,.+100", &Range{Relative{-100}, Relative{100}}, ""},
		{"'", nil, "'"},
		{"' ", nil, "' "},
		{"'<", &Range{VisualStart{}, nil}, ""},
		{"'>", &Range{VisualEnd{}, nil}, ""},
		{" '<  ,  '>  write", &Range{VisualStart{}, VisualEnd{}}, "write"},
		{" '<+0x10 ,  '>-10w", &Range{VisualStart{0x10}, VisualEnd{-10}}, "w"},
	}
	for _, testCase := range testCases {
		got, rest := ParseRange(testCase.target)
		if !reflect.DeepEqual(got, testCase.expected) || rest != testCase.rest {
			t.Errorf("ParseRange(%q) should return\n\t%#v, %q\nbut got\n\t%#v, %q",
				testCase.target, testCase.expected, testCase.rest, got, rest)
		}
	}
}
