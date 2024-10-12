package cmdline

import (
	"errors"
	"strings"
	"unicode"

	"github.com/itchyny/bed/event"
)

func parse(cmdline []rune) (command, *event.Range, string, bool, string, error) {
	i, l := 0, len(cmdline)
	for i < l && (unicode.IsSpace(cmdline[i]) || cmdline[i] == ':') {
		i++
	}
	if i == l {
		return command{}, nil, "", false, "", nil
	}
	r, i := event.ParseRange(cmdline, i)
	j := i
	for j < l && !unicode.IsSpace(cmdline[j]) {
		j++
	}
	k := j
	for k < l && unicode.IsSpace(cmdline[k]) {
		k++
	}
	cmdName, bang := strings.CutSuffix(string(cmdline[i:j]), "!")
	for _, cmd := range commands {
		if len(cmdName) == 0 || cmdName[0] != cmd.name[0] {
			continue
		}
		for _, c := range expand(cmd.name) {
			if cmdName == c {
				return cmd, r, string(cmdline[:k]), bang, strings.TrimSpace(string(cmdline[k:])), nil
			}
		}
	}
	if len(strings.Fields(string(cmdline[k:]))) == 0 && r != nil {
		return command{"goto", event.CursorGoto}, r, string(cmdline[:k]), bang, "", nil
	}
	return command{}, nil, "", false, "", errors.New("unknown command: " + string(cmdline))
}

func expand(name string) []string {
	var prefix, abbr string
	if i := strings.IndexRune(name, '['); i > 0 {
		prefix = name[:i]
		j := strings.IndexRune(name, ']')
		abbr = name[i+1 : j]
	}
	if len(abbr) == 0 {
		return []string{name}
	}
	cmds := make([]string, len(abbr)+1)
	for i := range len(abbr) + 1 {
		cmds[i] = prefix + abbr[:i]
	}
	return cmds
}
