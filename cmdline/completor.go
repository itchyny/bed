package cmdline

import (
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
	if cmd.eventType != EventEdit && cmd.eventType != EventNew && cmd.eventType != EventVnew {
		c.results = nil
		c.index = 0
		return cmdline
	}
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
		c.target = prefix + samePrefix(c.results)
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
		fileInfos, err := f.Readdir(100)
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
		fileInfos, err := f.Readdir(300)
		if err != nil {
			return nil
		}
		lowerBase := strings.ToLower(base)
		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			if base != separator && !strings.HasPrefix(strings.ToLower(name), lowerBase) {
				continue
			}
			name = filepath.Join(dir, name)
			if prefix[0] == '~' {
				name = filepath.Join("~", strings.TrimLeft(name, homeDir))
			}
			if fileInfo.IsDir() {
				name += separator
			}
			targets = append(targets, name)
		}
	}
	sortFilePaths(targets)
	return targets
}

func sortFilePaths(paths []string) {
	sort.SliceStable(paths, func(i, j int) bool {
		return filePathSorter(paths[i]) < filePathSorter(paths[j])
	})
}

func filePathSorter(path string) string {
	var ret string
	if path[len(path)-1] == filepath.Separator {
		ret += "1"
	} else {
		ret += "0"
	}
	if strings.HasPrefix(filepath.Base(path), ".") {
		ret += "1"
	} else {
		ret += "0"
	}
	return ret + path
}

func samePrefix(results []string) string {
	var xs string
	for i, ys := range results {
		if i == 0 {
			xs = ys
		} else {
			yss := []rune(ys)
			for j, x := range xs {
				if j < len(yss) {
					if x != yss[j] {
						xs = string(yss[:j])
						break
					}
				} else {
					xs = string(yss)
					break
				}
			}
		}
	}
	return xs
}
