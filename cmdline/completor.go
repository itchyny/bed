package cmdline

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mitchellh/go-homedir"

	"github.com/itchyny/bed/event"
)

type completor struct {
	fs      fs
	target  string
	arg     string
	results []string
	index   int
}

func newCompletor(fs fs) *completor {
	return &completor{fs: fs}
}

func (c *completor) clear() {
	c.target = ""
	c.arg = ""
	c.results = nil
	c.index = 0
}

func (c *completor) complete(cmdline string, cmd command, prefix string, arg string, forward bool) string {
	switch cmd.eventType {
	case event.Edit, event.New, event.Vnew, event.Write:
		return c.completeFilepaths(cmdline, prefix, arg, forward)
	case event.Wincmd:
		return c.completeWincmd(cmdline, prefix, arg, forward)
	default:
		c.results = nil
		c.index = 0
		return cmdline
	}
}

func (c *completor) completeNext(prefix string, forward bool) string {
	if forward {
		c.index = (c.index+2)%(len(c.results)+1) - 1
	} else {
		c.index = (c.index+len(c.results)+1)%(len(c.results)+1) - 1
	}
	if c.index < 0 {
		return c.target
	}
	return prefix + c.arg + c.results[c.index]
}

func (c *completor) completeFilepaths(cmdline string, prefix string, arg string, forward bool) string {
	if !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	if len(c.results) > 0 {
		return c.completeNext(prefix, forward)
	}
	c.target = cmdline
	c.index = 0
	if len(arg) == 0 {
		c.arg, c.results = c.listFileNames("")
	} else {
		c.arg, c.results = c.listFileNames(arg)
	}
	if len(c.results) == 1 {
		cmdline := prefix + c.arg + c.results[0]
		c.results = nil
		return cmdline
	}
	if len(c.results) > 1 {
		if forward {
			c.index = 0
			return prefix + c.arg + c.results[0]
		}
		c.index = len(c.results) - 1
		return prefix + c.arg + c.results[len(c.results)-1]
	}
	return cmdline
}

func (c *completor) listFileNames(arg string) (string, []string) {
	var targets []string
	separator := string(filepath.Separator)
	if arg == "" {
		f, err := c.fs.Open(".")
		if err != nil {
			return arg, nil
		}
		defer f.Close()
		fileInfos, err := f.Readdir(500)
		if err != nil {
			return arg, nil
		}
		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			isDir := fileInfo.IsDir()
			if !isDir && fileInfo.Mode()&os.ModeSymlink != 0 {
				fileInfo, err = c.fs.Stat(name)
				if err != nil {
					return arg, nil
				}
				isDir = fileInfo.IsDir()
			}
			if isDir {
				name += separator
			}
			targets = append(targets, name)
		}
	} else {
		path, err := homedir.Expand(arg)
		if err != nil {
			return arg, nil
		}
		homeDir, err := homedir.Dir()
		if err != nil {
			return arg, nil
		}
		if strings.HasSuffix(arg, separator) ||
			arg[0] == '~' && strings.HasSuffix(arg, separator+".") {
			path += separator
		}
		if !strings.HasSuffix(arg, separator) && !strings.HasSuffix(arg, ".") {
			stat, err := c.fs.Stat(path)
			if err == nil && stat.IsDir() {
				return "", []string{arg + separator}
			}
		}
		dir, base := filepath.Dir(path), filepath.Base(path)
		if strings.HasSuffix(path, separator) {
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
		fileInfos, err := f.Readdir(500)
		if err != nil {
			return arg, nil
		}
		lowerBase := strings.ToLower(base)
		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			if base != separator && !strings.HasPrefix(strings.ToLower(name), lowerBase) {
				continue
			}
			isDir := fileInfo.IsDir()
			if !isDir && fileInfo.Mode()&os.ModeSymlink != 0 {
				fileInfo, err = c.fs.Stat(filepath.Join(dir, name))
				if err != nil {
					return arg, nil
				}
				isDir = fileInfo.IsDir()
			}
			if isDir {
				name += separator
			}
			targets = append(targets, name)
		}
		if !strings.HasSuffix(dir, separator) {
			dir += separator
		}
		if arg[0] == '~' {
			arg = filepath.Join("~", strings.TrimPrefix(dir, homeDir))
			if !strings.HasSuffix(arg, separator) {
				arg += separator
			}
		} else if arg[0] != '.' && dir[0] == '.' {
			arg = ""
		} else {
			arg = dir
		}
	}
	sortFilePaths(targets)
	return arg, targets
}

func sortFilePaths(paths []string) {
	for i, path := range paths {
		prefix := make([]rune, 2)
		if path[len(path)-1] == filepath.Separator {
			prefix[0] = '1'
		} else {
			prefix[0] = '0'
		}
		if strings.HasPrefix(filepath.Base(path), ".") {
			prefix[1] = '1'
		} else {
			prefix[1] = '0'
		}
		paths[i] = string(prefix) + path
	}
	sort.Strings(paths)
	for i, path := range paths {
		paths[i] = path[2:]
	}
}

func (c *completor) completeWincmd(cmdline string, prefix string, arg string, forward bool) string {
	if !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	if len(c.results) > 0 {
		return c.completeNext(prefix, forward)
	}
	if len(arg) > 0 {
		return cmdline
	}
	c.target = cmdline
	c.results = []string{"n", "h", "l", "k", "j", "H", "L", "K", "J", "t", "b", "p"}
	c.index = -1
	return cmdline
}
