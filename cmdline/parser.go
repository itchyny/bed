package cmdline

import (
	"fmt"
	"strings"
	"unicode"
)

func parse(cmdline []rune) (command, error) {
	i, l := 0, len(cmdline)
	for i < l && (unicode.IsSpace(cmdline[i]) || cmdline[i] == ':') {
		i++
	}
	xs := strings.Fields(string(cmdline[i:]))
	if len(xs) == 0 {
		return command{}, nil
	}
	for _, cmd := range commands {
		if xs[0][0] != cmd.name[0] {
			continue
		}
		for _, c := range expand(cmd.name) {
			if xs[0] == c {
				return cmd, nil
			}
		}
	}
	return command{}, fmt.Errorf("unknown command: %s", string(cmdline))
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
