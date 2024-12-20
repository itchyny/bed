package cmdline

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/itchyny/bed/event"
)

type completor struct {
	fs      fs
	env     env
	command bool
	target  string
	arg     string
	results []string
	index   int
}

func newCompletor(fs fs, env env) *completor {
	return &completor{fs: fs, env: env}
}

func (c *completor) complete(cmdline string, forward bool) string {
	cmd, r, _, name, prefix, arg, _ := parse(cmdline)
	if name == "" || c.command ||
		!hasSuffixFunc(prefix, unicode.IsSpace) && cmd.fullname != name {
		cmdline = c.completeCommand(cmdline, name, prefix, r, forward)
		if c.results != nil {
			return cmdline
		}
		prefix = cmdline
	}
	switch cmd.eventType {
	case event.Edit, event.New, event.Vnew, event.Write, event.WriteQuit:
		return c.completeFilepath(cmdline, prefix, arg, forward, false)
	case event.Chdir:
		return c.completeFilepath(cmdline, prefix, arg, forward, true)
	case event.Wincmd:
		return c.completeWincmd(cmdline, prefix, arg, forward)
	default:
		return cmdline
	}
}

func (c *completor) completeNext(prefix string, forward bool) string {
	if len(c.results) == 0 {
		return c.target
	}
	if forward {
		c.index = (c.index+2)%(len(c.results)+1) - 1
	} else {
		c.index = (c.index+len(c.results)+1)%(len(c.results)+1) - 1
	}
	if c.index < 0 {
		return c.target
	}
	if len(c.results) == 1 {
		defer c.clear()
	}
	return prefix + c.arg + c.results[c.index]
}

func (c *completor) completeCommand(
	cmdline, name, prefix string, r *event.Range, forward bool,
) string {
	prefix = prefix[:len(prefix)-len(name)]
	if c.results == nil {
		c.command, c.target, c.index = true, cmdline, -1
		c.arg, c.results = "", listCommandNames(name, r)
	}
	return c.completeNext(prefix, forward)
}

func listCommandNames(name string, r *event.Range) []string {
	var targets []string
	for _, cmd := range commands {
		if strings.HasPrefix(cmd.fullname, name) && cmd.rangeType.allows(r) {
			targets = append(targets, cmd.fullname)
		}
	}
	slices.Sort(targets)
	return targets
}

func (c *completor) completeFilepath(
	cmdline, prefix, arg string, forward, dirOnly bool,
) string {
	if !hasSuffixFunc(prefix, unicode.IsSpace) {
		prefix += " "
	}
	if c.results == nil {
		c.command, c.target, c.index = false, cmdline, -1
		c.arg, c.results = c.listFileNames(arg, dirOnly)
	}
	return c.completeNext(prefix, forward)
}

const separator = string(filepath.Separator)

func (c *completor) listFileNames(arg string, dirOnly bool) (string, []string) {
	var targets []string
	path, simplify := c.expandPath(arg)
	if strings.HasPrefix(arg, "$") && !strings.Contains(arg, separator) {
		base := strings.ToLower(arg[1:])
		for _, env := range c.env.List() {
			name, value, ok := strings.Cut(env, "=")
			if !ok {
				continue
			}
			if !strings.HasPrefix(strings.ToLower(name), base) {
				continue
			}
			if !filepath.IsAbs(value) {
				continue
			}
			fi, err := c.fs.Stat(value)
			if err != nil {
				continue
			}
			if fi.IsDir() {
				name += separator
			} else if dirOnly {
				continue
			}
			targets = append(targets, "$"+name)
		}
		slices.Sort(targets)
		return "", targets
	}
	if arg != "" && !strings.HasSuffix(arg, separator) &&
		(!strings.HasSuffix(arg, ".") || strings.HasSuffix(arg, "..")) {
		if stat, err := c.fs.Stat(path); err == nil && stat.IsDir() {
			return "", []string{arg + separator}
		}
	}
	if strings.HasSuffix(arg, separator) || strings.HasSuffix(arg, separator+".") {
		path += separator
	}
	dir, base := filepath.Dir(path), strings.ToLower(filepath.Base(path))
	if arg == "" {
		base = ""
	} else if strings.HasSuffix(path, separator) {
		if strings.HasSuffix(arg, separator+".") {
			base = "."
		} else {
			base = ""
		}
	}
	f, err := c.fs.Open(dir)
	if err != nil {
		return arg, nil
	}
	defer f.Close()
	fileInfos, err := f.Readdir(1024)
	if err != nil {
		return arg, nil
	}
	for _, fileInfo := range fileInfos {
		name := fileInfo.Name()
		if !strings.HasPrefix(strings.ToLower(name), base) {
			continue
		}
		isDir := fileInfo.IsDir()
		if !isDir && fileInfo.Mode()&os.ModeSymlink != 0 {
			fileInfo, err := c.fs.Stat(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			isDir = fileInfo.IsDir()
		}
		if isDir {
			name += separator
		} else if dirOnly {
			continue
		}
		targets = append(targets, name)
	}
	slices.SortFunc(targets, func(p, q string) int {
		ps, pd := p[len(p)-1] == filepath.Separator, p[0] == '.'
		qs, qd := q[len(q)-1] == filepath.Separator, q[0] == '.'
		switch {
		case ps && !qs:
			return 1
		case !ps && qs:
			return -1
		case pd && !qd:
			return 1
		case !pd && qd:
			return -1
		default:
			return strings.Compare(p, q)
		}
	})
	if simplify != nil {
		arg = simplify(dir) + separator
	} else if !strings.HasPrefix(arg, "."+separator) && dir == "." {
		arg = ""
	} else if arg = dir; !strings.HasSuffix(arg, separator) {
		arg += separator
	}
	return arg, targets
}

func (c *completor) expandPath(path string) (string, func(string) string) {
	switch {
	case strings.HasPrefix(path, "~"):
		if name, rest, _ := strings.Cut(path[1:], separator); name != "" {
			user, err := c.fs.GetUser(name)
			if err != nil {
				return path, nil
			}
			return filepath.Join(user.HomeDir, rest), func(path string) string {
				return filepath.Join("~"+user.Username, strings.TrimPrefix(path, user.HomeDir))
			}
		}
		homedir, err := c.fs.UserHomeDir()
		if err != nil {
			return path, nil
		}
		return filepath.Join(homedir, path[1:]), func(path string) string {
			return filepath.Join("~", strings.TrimPrefix(path, homedir))
		}
	case strings.HasPrefix(path, "$"):
		name, rest, _ := strings.Cut(path[1:], separator)
		value := strings.TrimRight(c.env.Get(name), separator)
		if value == "" {
			return path, nil
		}
		return filepath.Join(value, rest), func(path string) string {
			return filepath.Join("$"+name, strings.TrimPrefix(path, value))
		}
	default:
		return path, nil
	}
}

func (c *completor) completeWincmd(
	cmdline, prefix, arg string, forward bool,
) string {
	if !hasSuffixFunc(prefix, unicode.IsSpace) {
		prefix += " "
	}
	if c.results == nil {
		if arg != "" {
			return cmdline
		}
		c.command, c.target, c.arg, c.index = false, cmdline, "", -1
		c.results = strings.Split("nohjkltbpHJKL", "")
	}
	return c.completeNext(prefix, forward)
}

func (c *completor) clear() {
	c.command, c.target, c.arg = false, "", ""
	c.results, c.index = nil, 0
}

func hasSuffixFunc(s string, f func(rune) bool) bool {
	r, size := utf8.DecodeLastRuneInString(s)
	return size > 0 && f(r)
}
