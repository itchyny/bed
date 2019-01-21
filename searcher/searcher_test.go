package searcher

import (
	"strings"
	"testing"
)

func TestSearcher(t *testing.T) {
	testCases := []struct {
		str      string
		cursor   int64
		pattern  string
		forward  bool
		expected int64
		err      error
	}{
		{
			str:      "abcde",
			cursor:   0,
			pattern:  "cd",
			forward:  true,
			expected: 2,
		},
		{
			str:     "abcde",
			cursor:  2,
			pattern: "cd",
			forward: true,
			err:     errNotFound("cd"),
		},
		{
			str:      "abcde",
			cursor:   4,
			pattern:  "bc",
			forward:  false,
			expected: 1,
		},
		{
			str:     "abcde",
			cursor:  0,
			pattern: "ba",
			forward: true,
			err:     errNotFound("ba"),
		},
		{
			str:      strings.Repeat(" ", 10*1024*1024+100) + "abcde",
			cursor:   102,
			pattern:  "bcd",
			forward:  true,
			expected: 10*1024*1024 + 101,
		},
		{
			str:     strings.Repeat(" ", 10*1024*1024+100) + "abcde",
			cursor:  102,
			pattern: "cba",
			forward: true,
			err:     errNotFound("cba"),
		},
		{
			str:      "abcde" + strings.Repeat(" ", 10*1024*1024),
			cursor:   10*1024*1024 + 2,
			pattern:  "bcd",
			forward:  false,
			expected: 1,
		},
		{
			str:     "abcde" + strings.Repeat(" ", 10*1024*1024),
			cursor:  10*1024*1024 + 2,
			pattern: "cba",
			forward: false,
			err:     errNotFound("cba"),
		},
	}
	for _, testCase := range testCases {
		s := NewSearcher(strings.NewReader(testCase.str))
		ch := s.Search(testCase.cursor, testCase.pattern, testCase.forward)
		select {
		case x := <-ch:
			switch x := x.(type) {
			case error:
				if testCase.err == nil {
					t.Error(x)
				} else if x != testCase.err {
					t.Errorf("Error should be %v but got %v", testCase.err, x)
				}
			case int64:
				if x != testCase.expected {
					t.Errorf("Search result should be %d but got %d", testCase.expected, x)
				}
			}
		}
	}
}
