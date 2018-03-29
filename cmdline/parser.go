package cmdline

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/itchyny/bed/event"
)

func parse(cmdline []rune) (command, string, string, error) {
	i, l := 0, len(cmdline)
	for i < l && (unicode.IsSpace(cmdline[i]) || cmdline[i] == ':') {
		i++
	}
	if i == l {
		return command{}, "", "", nil
	}
	_, i = parseRange(cmdline, i)
	j := i
	for j < l && !unicode.IsSpace(cmdline[j]) {
		j++
	}
	k := j
	for k < l && unicode.IsSpace(cmdline[k]) {
		k++
	}
	cmdName := string(cmdline[i:j])
	for _, cmd := range commands {
		if cmdName[0] != cmd.name[0] {
			continue
		}
		for _, c := range expand(cmd.name) {
			if cmdName == c {
				return cmd, string(cmdline[:k]), strings.TrimSpace(string(cmdline[k:])), nil
			}
		}
	}
	if len(strings.Fields(string(cmdline[k:]))) == 0 {
		if cmdName == "$" {
			return command{cmdName, event.CursorGotoAbs}, string(cmdline[:k]), cmdName, nil
		}
		relative, eventType := false, event.CursorGotoAbs
		for _, c := range cmdName {
			if !relative && (c == '-' || c == '+') {
				relative = true
				eventType = event.CursorGotoRel
			} else if !('0' <= c && c <= '9' || 'a' <= c && c <= 'f') {
				eventType = event.Nop
				break
			}
		}
		if eventType != event.Nop {
			return command{cmdName, event.Type(eventType)}, string(cmdline[:k]), cmdName, nil
		}
	}
	return command{}, "", "", fmt.Errorf("unknown command: %s", string(cmdline))
}

func parseRange(cmdline []rune, i int) (*event.Range, int) {
	l := len(cmdline)
	from, i := parsePos(cmdline, i)
	if from == nil {
		return nil, i
	}
	if i >= l {
		return nil, i
	}
	for i < l && unicode.IsSpace(cmdline[i]) {
		i++
	}
	if i >= l || cmdline[i] != ',' {
		return nil, i
	}
	i++
	for i < l && unicode.IsSpace(cmdline[i]) {
		i++
	}
	to, i := parsePos(cmdline, i)
	if to == nil {
		return nil, i
	}
	return &event.Range{From: from, To: to}, i
}

func parsePos(cmdline []rune, i int) (event.Position, int) {
	l := len(cmdline)
	if i >= l {
		return nil, i
	}
	if cmdline[i] == '\'' {
		if i+1 < l {
			if cmdline[i+1] == '<' {
				return event.VisualStart{}, i + 2
			} else if cmdline[i+1] == '>' {
				return event.VisualEnd{}, i + 2
			}
		}
	}
	return nil, i
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
