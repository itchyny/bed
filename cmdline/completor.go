package cmdline

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mitchellh/go-homedir"

	. "github.com/itchyny/bed/common"
)

type completor struct {
	fs      fs
	target  string
	results []string
	index   int
}

func newCompletor(fs fs) *completor {
	return &completor{fs: fs}
}

func (c *completor) clear() {
	c.target = ""
	c.results = nil
	c.index = 0
}

func (c *completor) complete(cmdline string, cmd command, prefix string, arg string, forward bool) string {
	switch cmd.eventType {
	case EventEdit, EventNew, EventVnew:
		return c.completeFilePaths(cmdline, prefix, arg, forward)
	default:
		c.results = nil
		c.index = 0
		return cmdline
	}
}

func (c *completor) completeFilePaths(cmdline string, prefix string, arg string, forward bool) string {
	if !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	if len(c.results) > 0 {
		if forward {
			c.index = (c.index+2)%(len(c.results)+1) - 1
		} else {
			c.index = (c.index+len(c.results)+1)%(len(c.results)+1) - 1
		}
		if c.index < 0 {
			return c.target
		}
		return prefix + c.results[c.index]
	}
	c.target = cmdline
	c.index = 0
	if len(arg) == 0 {
		c.results = c.listFileNames("")
	} else {
		c.results = c.listFileNames(arg)
	}
	if len(c.results) == 1 {
		cmdline := prefix + c.results[0]
		c.results = nil
		return cmdline
	}
	if len(c.results) > 1 {
		if forward {
			c.index = 0
			return prefix + c.results[0]
		}
		c.index = len(c.results) - 1
		return prefix + c.results[len(c.results)-1]
	}
	return cmdline
}

func (c *completor) listFileNames(prefix string) []string {
	var targets []string
	separator := string(filepath.Separator)
	if prefix == "" {
		f, err := c.fs.Open(".")
		if err != nil {
			return nil
		}
		defer f.Close()
		fileInfos, err := f.Readdir(500)
		if err != nil {
			return nil
		}
		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			if fileInfo.IsDir() {
				name += separator
			}
			targets = append(targets, name)
		}
	} else {
		path, err := homedir.Expand(prefix)
		if err != nil {
			return nil
		}
		homeDir, err := homedir.Dir()
		if err != nil {
			return nil
		}
		if len(prefix) > 1 && strings.HasSuffix(prefix, separator) ||
			prefix[0] == '~' && strings.HasSuffix(prefix, "/.") {
			path += separator
		}
		if !strings.HasSuffix(prefix, "/") && !strings.HasSuffix(prefix, ".") {
			stat, err := c.fs.Stat(path)
			if err == nil && stat.IsDir() {
				return []string{prefix + "/"}
			}
		}
		dir, base := filepath.Dir(path), filepath.Base(path)
		if strings.HasSuffix(path, separator) {
			if strings.HasSuffix(prefix, "/.") {
				base = "."
			} else {
				base = ""
			}
		}
		f, err := c.fs.Open(dir)
		if err != nil {
			return nil
		}
		defer f.Close()
		fileInfos, err := f.Readdir(500)
		if err != nil {
			return nil
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
					return nil
				}
				isDir = fileInfo.IsDir()
			}
			name = filepath.Join(dir, name)
			if prefix[0] == '~' {
				name = filepath.Join("~", strings.TrimPrefix(name, homeDir))
			}
			if isDir {
				name += separator
			}
			targets = append(targets, name)
		}
	}
	sortFilePaths(targets)
	return targets
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
