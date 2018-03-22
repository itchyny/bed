package cmdline

import (
	"fmt"
	"strings"
	"unicode"

	. "github.com/itchyny/bed/common"
)

func parse(cmdline []rune) (command, string, error) {
	i, l := 0, len(cmdline)
	for i < l && (unicode.IsSpace(cmdline[i]) || cmdline[i] == ':') {
		i++
	}
	if i == l {
		return command{}, "", nil
	}
	j := i
	for j < l && !unicode.IsSpace(cmdline[j]) {
		j++
	}
	cmdName := string(cmdline[i:j])
	for _, cmd := range commands {
		if cmdName[0] != cmd.name[0] {
			continue
		}
		for _, c := range expand(cmd.name) {
			if cmdName == c {
				return cmd, strings.TrimSpace(string(cmdline[j:])), nil
			}
		}
	}
	if len(strings.Fields(string(cmdline[j:]))) == 0 {
		if cmdName == "$" {
			return command{cmdName, EventCursorGotoAbs}, cmdName, nil
		}
		relative, eventType := false, EventCursorGotoAbs
		for _, c := range cmdName {
			if !relative && (c == '-' || c == '+') {
				relative = true
				eventType = EventCursorGotoRel
			} else if !('0' <= c && c <= '9' || 'a' <= c && c <= 'f') {
				eventType = EventNop
				break
			}
		}
		if eventType != EventNop {
			return command{cmdName, EventType(eventType)}, cmdName, nil
		}
	}
	return command{}, "", fmt.Errorf("unknown command: %s", string(cmdline))
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
