package cmdline

import (
	"fmt"
	"strings"
	"unicode"

	. "github.com/itchyny/bed/common"
)

func parse(cmdline []rune) (command, []string, error) {
	i, l := 0, len(cmdline)
	for i < l && (unicode.IsSpace(cmdline[i]) || cmdline[i] == ':') {
		i++
	}
	xs := strings.Fields(string(cmdline[i:]))
	if len(xs) == 0 {
		return command{}, nil, nil
	}
	for _, cmd := range commands {
		if xs[0][0] != cmd.name[0] {
			continue
		}
		for _, c := range expand(cmd.name) {
			if xs[0] == c {
				return cmd, xs[1:], nil
			}
		}
	}
	if len(xs) == 1 {
		if xs[0] == "$" {
			return command{xs[0], EventCursorGotoAbs}, xs, nil
		}
		relative, eventType := false, EventCursorGotoAbs
		for _, c := range xs[0] {
			if !relative && (c == '-' || c == '+') {
				relative = true
				eventType = EventCursorGotoRel
			} else if !('0' <= c && c <= '9' || 'a' <= c && c <= 'f') {
				eventType = EventNop
				break
			}
		}
		if eventType != EventNop {
			return command{xs[0], EventType(eventType)}, xs, nil
		}
	}
	return command{}, nil, fmt.Errorf("unknown command: %s", string(cmdline))
}

func expand(name string) []string {
	var prefix, abbr string
	if i := strings.IndexRune(name, '['); i > 0 {
		prefix = name[:i]
		abbr = name[i+1 : len(name)-1]
	}
	if len(abbr) == 0 {
		return []string{name}
	}
	cmds := make([]string, len(abbr)+1)
	for i := 0; i <= len(abbr); i++ {
		cmds[i] = prefix + abbr[:i]
	}
	return cmds
}
