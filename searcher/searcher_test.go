package searcher

import (
	"strings"
	"testing"
)

func TestSearcher(t *testing.T) {
	testCases := []struct {
		name     string
		str      string
		cursor   int64
		pattern  string
		forward  bool
		expected int64
		err      error
	}{
		{
			name:     "search forward",
			str:      "abcde",
			cursor:   0,
			pattern:  "cd",
			forward:  true,
			expected: 2,
		},
		{
			name:    "search forward but not found",
			str:     "abcde",
			cursor:  2,
			pattern: "cd",
			forward: true,
			err:     errNotFound("cd"),
		},
		{
			name:     "search backward",
			str:      "abcde",
			cursor:   4,
			pattern:  "bc",
			forward:  false,
			expected: 1,
		},
		{
			name:    "search backward but not found",
			str:     "abcde",
			cursor:  0,
			pattern: "ba",
			forward: true,
			err:     errNotFound("ba"),
		},
		{
			name:     "search large target forward",
			str:      strings.Repeat(" ", 10*1024*1024+100) + "abcde",
			cursor:   102,
			pattern:  "bcd",
			forward:  true,
			expected: 10*1024*1024 + 101,
		},
		{
			name:    "search large target forward but not found",
			str:     strings.Repeat(" ", 10*1024*1024+100) + "abcde",
			cursor:  102,
			pattern: "cba",
			forward: true,
			err:     errNotFound("cba"),
		},
		{
			name:     "search large target backward",
			str:      "abcde" + strings.Repeat(" ", 10*1024*1024),
			cursor:   10*1024*1024 + 2,
			pattern:  "bcd",
			forward:  false,
			expected: 1,
		},
		{
			name:    "search large target backward but not found",
			str:     "abcde" + strings.Repeat(" ", 10*1024*1024),
			cursor:  10*1024*1024 + 2,
			pattern: "cba",
			forward: false,
			err:     errNotFound("cba"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
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
		})
	}
}
