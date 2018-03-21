package cmdline

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/mitchellh/go-homedir"

	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/util"
)

// Cmdline implements editor.Cmdline
type Cmdline struct {
	cmdline           []rune
	cursor            int
	completionTarget  string
	completionResults []string
	completionIndex   int
	eventCh           chan<- Event
	cmdlineCh         <-chan Event
	redrawCh          chan<- struct{}
}

// NewCmdline creates a new Cmdline.
func NewCmdline() *Cmdline {
	return &Cmdline{}
}

// Init initializes the Cmdline.
func (c *Cmdline) Init(eventCh chan<- Event, cmdlineCh <-chan Event, redrawCh chan<- struct{}) {
	c.eventCh, c.cmdlineCh, c.redrawCh = eventCh, cmdlineCh, redrawCh
}

// Run the cmdline.
func (c *Cmdline) Run() {
	for e := range c.cmdlineCh {
		switch e.Type {
		case EventStartCmdline:
			c.clear()
		case EventExitCmdline:
			// do nothing here but redraw
		case EventCursorLeft:
			c.cursorLeft()
		case EventCursorRight:
			c.cursorRight()
		case EventCursorHead:
			c.cursorHead()
		case EventCursorEnd:
			c.cursorEnd()
		case EventBackspaceCmdline:
			c.backspace()
		case EventDeleteCmdline:
			c.deleteRune()
		case EventDeleteWordCmdline:
			c.deleteWord()
		case EventClearToHeadCmdline:
			c.clearToHead()
		case EventClearCmdline:
			c.clear()
		case EventRune:
			c.insert(e.Rune)
		case EventCompleteForwardCmdline:
			c.complete(true)
			c.redrawCh <- struct{}{}
			continue
		case EventCompleteBackCmdline:
			c.complete(false)
			c.redrawCh <- struct{}{}
			continue
		case EventExecuteCmdline:
			c.execute()
		default:
			continue
		}
		c.completionResults = nil
		c.completionIndex = 0
		c.redrawCh <- struct{}{}
	}
}

func (c *Cmdline) cursorLeft() {
	c.cursor = util.MaxInt(0, c.cursor-1)
}

func (c *Cmdline) cursorRight() {
	c.cursor = util.MinInt(len(c.cmdline), c.cursor+1)
}

func (c *Cmdline) cursorHead() {
	c.cursor = 0
}

func (c *Cmdline) cursorEnd() {
	c.cursor = len(c.cmdline)
}

func (c *Cmdline) backspace() {
	if c.cursor > 0 {
		c.cmdline = append(c.cmdline[:c.cursor-1], c.cmdline[c.cursor:]...)
		c.cursor--
	}
}

func (c *Cmdline) deleteRune() {
	if c.cursor < len(c.cmdline) {
		c.cmdline = append(c.cmdline[:c.cursor], c.cmdline[c.cursor+1:]...)
	}
}

func (c *Cmdline) deleteWord() {
	i := c.cursor
	for i > 0 && unicode.IsSpace(c.cmdline[i-1]) {
		i--
	}
	if i > 0 {
		isk := isKeyword(c.cmdline[i-1])
		for i > 0 && isKeyword(c.cmdline[i-1]) == isk {
			i--
		}
	}
	c.cmdline = append(c.cmdline[:i], c.cmdline[c.cursor:]...)
	c.cursor = i
}

func isKeyword(c rune) bool {
	return unicode.IsDigit(c) || unicode.IsLetter(c) || c == '_'
}

func (c *Cmdline) clear() {
	c.cmdline = []rune{}
	c.cursor = 0
}

func (c *Cmdline) clearToHead() {
	c.cmdline = c.cmdline[c.cursor:]
	c.cursor = 0
}

func (c *Cmdline) insert(ch rune) {
	if unicode.IsPrint(ch) {
		c.cmdline = append(c.cmdline, '\x00')
		copy(c.cmdline[c.cursor+1:], c.cmdline[c.cursor:])
		c.cmdline[c.cursor] = ch
		c.cursor++
	}
}

func (c *Cmdline) complete(forward bool) {
	cmd, args, err := parse(c.cmdline)
	if err != nil {
		c.completionResults = nil
		c.completionIndex = 0
		return
	}
	if cmd.eventType != EventEdit && cmd.eventType != EventNew && cmd.eventType != EventVnew {
		c.completionResults = nil
		c.completionIndex = 0
		return
	}
	cmdName := strings.Fields(string(c.cmdline))[0] // todo: parse should return argument position
	if len(c.completionResults) > 0 {
		if forward {
			c.completionIndex = (c.completionIndex+2)%(len(c.completionResults)+1) - 1
		} else {
			c.completionIndex = (c.completionIndex+len(c.completionResults)+1)%(len(c.completionResults)+1) - 1
		}
		if c.completionIndex < 0 {
			c.cmdline = []rune(c.completionTarget)
		} else {
			c.cmdline = []rune(cmdName + " " + c.completionResults[c.completionIndex])
		}
		c.cursor = len(c.cmdline)
		return
	}
	c.completionTarget = string(c.cmdline)
	c.completionIndex = 0
	if len(args) == 0 {
		c.completionResults = listFileNames("")
	} else if len(args) == 1 {
		c.completionResults = listFileNames(args[0])
	}
	if len(c.completionResults) == 1 {
		c.cmdline = []rune(cmdName + " " + c.completionResults[0])
		c.cursor = len(c.cmdline)
		c.completionResults = nil
	} else if len(c.completionResults) > 1 {
		if forward {
			c.cmdline = []rune(cmdName + " " + c.completionResults[0])
			c.completionIndex = 0
			c.cursor = len(c.cmdline)
		} else {
			c.cmdline = []rune(cmdName + " " + c.completionResults[len(c.completionResults)-1])
			c.completionIndex = len(c.completionResults) - 1
			c.cursor = len(c.cmdline)
		}
		c.completionTarget = cmdName + " " + samePrefix(c.completionResults)
	}
}

func listFileNames(prefix string) []string {
	var targets []string
	separator := string(filepath.Separator)
	if prefix == "" {
		f, err := os.Open(".")
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
			stat, err := os.Stat(path)
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
		f, err := os.Open(dir)
		if err != nil {
			return nil
		}
		defer f.Close()
		fileInfos, err := f.Readdir(300)
		if err != nil {
			return nil
		}
		for _, fileInfo := range fileInfos {
			name := fileInfo.Name()
			if base != separator && !strings.HasPrefix(name, base) {
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
	sort.SliceStable(targets, func(i, j int) bool {
		score := func(path string) string {
			var ret string
			if strings.HasSuffix(path, separator) {
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
		return score(targets[i]) < score(targets[j])
	})
	return targets
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

func (c *Cmdline) execute() {
	cmd, args, err := parse(c.cmdline)
	if err != nil {
		c.eventCh <- Event{Type: EventError, Error: err}
		return
	}
	if cmd.name != "" {
		c.eventCh <- Event{Type: cmd.eventType, CmdName: cmd.name, Args: args}
	}
}

// Get returns the current state of cmdline.
func (c *Cmdline) Get() ([]rune, int, []string, int) {
	return c.cmdline, c.cursor, c.completionResults, c.completionIndex
}
