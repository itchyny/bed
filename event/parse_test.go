package event

import (
	"reflect"
	"testing"
)

func TestParsePos(t *testing.T) {
	testCases := []struct {
		target   string
		expected Position
		index    int
	}{
		{"'<", VisualStart{}, 2},
		{"'>", VisualEnd{}, 2},
		{" '<  ,  '> ", VisualStart{}, 5},
	}
	for _, testCase := range testCases {
		got, gotIndex := ParsePos([]rune(testCase.target), 0)
		if !reflect.DeepEqual(got, testCase.expected) {
			t.Errorf("ParsePos(%q) should return %+v but got %+v", testCase.target, testCase.expected, got)
		}
		if gotIndex != testCase.index {
			t.Errorf("ParsePos(%q) should return index %d but got %d", testCase.target, testCase.index, gotIndex)
		}
	}
}
