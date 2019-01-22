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
		{
			name:     "search hex",
			str:      "\x13\x24\x35\x46\x57\x68",
			cursor:   0,
			pattern:  `\x35\x46\x57`,
			forward:  true,
			expected: 2,
		},
		{
			name:     "search nul",
			str:      "\x06\x07\x08\x00\x09\x10\x11",
			cursor:   0,
			pattern:  `\0`,
			forward:  true,
			expected: 3,
		},
		{
			name:     "search bell and bs",
			str:      "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x09\x0a",
			cursor:   0,
			pattern:  `\a\b\v`,
			forward:  true,
			expected: 7,
		},
		{
			name:     "search tab",
			str:      "\x06\x07\x08\x09\x10\x11",
			cursor:   0,
			pattern:  `\t`,
			forward:  true,
			expected: 3,
		},
		{
			name:     "search escape character",
			str:      `ab\cd\\e`,
			cursor:   0,
			pattern:  `\\\`,
			forward:  true,
			expected: 5,
		},
		{
			name:     "search unicode",
			str:      "\xe3\x81\x93\xe3\x82\x93\xe3\x81\xab\xe3\x81\xa1\xe3\x81\xaf",
			cursor:   0,
			pattern:  `\u3061\u306F`,
			forward:  true,
			expected: 9,
		},
		{
			name:     "search unicode in supplementary multilingual plane",
			str:      "\U0001F604\U0001F606\U0001F60E\U0001F60D\U0001F642",
			cursor:   0,
			pattern:  `\U0001F60E\U0001F60D`,
			forward:  true,
			expected: 8,
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
