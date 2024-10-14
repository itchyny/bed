package cmdline

import (
	"errors"
	"strings"
	"unicode"

	"github.com/itchyny/bed/event"
)

func parse(src string) (cmd command, r *event.Range,
	bang bool, prefix string, arg string, err error) {
	arg = strings.TrimLeftFunc(src, func(r rune) bool {
		return unicode.IsSpace(r) || r == ':'
	})
	if arg == "" {
		return
	}
	r, arg = event.ParseRange(arg)
	name, arg := cutPrefixFunc(arg, func(r rune) bool {
		return !unicode.IsSpace(r)
	})
	name, bang = strings.CutSuffix(name, "!")
	arg = strings.TrimLeftFunc(arg, unicode.IsSpace)
	prefix = src[:len(src)-len(arg)]
	for _, cmd = range commands {
		if matchCommand(cmd.name, name) {
			return
		}
	}
	if arg == "" && r != nil {
		cmd = command{"goto", event.CursorGoto}
		return
	}
	err = errors.New("unknown command: " + name)
	return
}

func cutPrefixFunc(src string, f func(rune) bool) (string, string) {
	for i, r := range src {
		if !f(r) {
			return src[:i], src[i:]
		}
	}
	return src, ""
}

func matchCommand(cmd, name string) bool {
	prefix, rest, _ := strings.Cut(cmd, "[")
	abbr, _, _ := strings.Cut(rest, "]")
	return strings.HasPrefix(name, prefix) &&
		strings.HasPrefix(abbr, name[len(prefix):])
}
