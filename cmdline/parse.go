package cmdline

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/itchyny/bed/event"
)

func parse(cmdline []rune) (command, *event.Range, string, string, error) {
	i, l := 0, len(cmdline)
	for i < l && (unicode.IsSpace(cmdline[i]) || cmdline[i] == ':') {
		i++
	}
	if i == l {
		return command{}, nil, "", "", nil
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
	cmdName := string(cmdline[i:j])
	for _, cmd := range commands {
		if len(cmdName) == 0 || cmdName[0] != cmd.name[0] {
			continue
		}
		for _, c := range expand(cmd.name) {
			if cmdName == c {
				return cmd, r, string(cmdline[:k]), strings.TrimSpace(string(cmdline[k:])), nil
			}
		}
	}
	if len(strings.Fields(string(cmdline[k:]))) == 0 && r != nil {
		return command{"goto", event.CursorGoto}, r, string(cmdline[:k]), "", nil
	}
	return command{}, nil, "", "", fmt.Errorf("unknown command: %s", string(cmdline))
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
