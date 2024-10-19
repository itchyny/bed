package cmdline

import (
	"slices"
	"sync"
	"unicode"

	"github.com/itchyny/bed/event"
)

// Cmdline implements editor.Cmdline
type Cmdline struct {
	cmdline      []rune
	cursor       int
	completor    *completor
	typ          rune
	historyIndex int
	history      []string
	histories    map[bool][]string
	eventCh      chan<- event.Event
	cmdlineCh    <-chan event.Event
	redrawCh     chan<- struct{}
	mu           *sync.Mutex
}

// NewCmdline creates a new Cmdline.
func NewCmdline() *Cmdline {
	return &Cmdline{
		completor: newCompletor(&filesystem{}, &environment{}),
		histories: map[bool][]string{false: {}, true: {}},
		mu:        new(sync.Mutex),
	}
}

// Init initializes the Cmdline.
func (c *Cmdline) Init(eventCh chan<- event.Event, cmdlineCh <-chan event.Event, redrawCh chan<- struct{}) {
	c.eventCh, c.cmdlineCh, c.redrawCh = eventCh, cmdlineCh, redrawCh
}

// Run the cmdline.
func (c *Cmdline) Run() {
	for e := range c.cmdlineCh {
		c.mu.Lock()
		switch e.Type {
		case event.StartCmdlineCommand:
			c.start(':', e.Arg)
		case event.StartCmdlineSearchForward:
			c.start('/', "")
		case event.StartCmdlineSearchBackward:
			c.start('?', "")
		case event.ExitCmdline:
			c.clear()
		case event.CursorUp:
			c.cursorUp()
		case event.CursorDown:
			c.cursorDown()
		case event.CursorLeft:
			c.cursorLeft()
		case event.CursorRight:
			c.cursorRight()
		case event.CursorHead:
			c.cursorHead()
		case event.CursorEnd:
			c.cursorEnd()
		case event.BackspaceCmdline:
			c.backspace()
		case event.DeleteCmdline:
			c.deleteRune()
		case event.DeleteWordCmdline:
			c.deleteWord()
		case event.ClearToHeadCmdline:
			c.clearToHead()
		case event.ClearCmdline:
			c.clear()
		case event.Rune:
			c.insert(e.Rune)
		case event.CompleteForwardCmdline:
			c.complete(true)
			c.redrawCh <- struct{}{}
			c.mu.Unlock()
			continue
		case event.CompleteBackCmdline:
			c.complete(false)
			c.redrawCh <- struct{}{}
			c.mu.Unlock()
			continue
		case event.ExecuteCmdline:
			if c.execute() {
				c.mu.Unlock()
				continue
			}
		default:
			c.mu.Unlock()
			continue
		}
		c.completor.clear()
		c.mu.Unlock()
		c.redrawCh <- struct{}{}
	}
}

func (c *Cmdline) cursorUp() {
	if c.historyIndex--; c.historyIndex >= 0 {
		c.cmdline = []rune(c.history[c.historyIndex])
		c.cursor = len(c.cmdline)
	} else {
		c.clear()
		c.historyIndex = -1
	}
}

func (c *Cmdline) cursorDown() {
	if c.historyIndex++; c.historyIndex < len(c.history) {
		c.cmdline = []rune(c.history[c.historyIndex])
		c.cursor = len(c.cmdline)
	} else {
		c.clear()
		c.historyIndex = len(c.history)
	}
}

func (c *Cmdline) cursorLeft() {
	c.cursor = max(0, c.cursor-1)
}

func (c *Cmdline) cursorRight() {
	c.cursor = min(len(c.cmdline), c.cursor+1)
}

func (c *Cmdline) cursorHead() {
	c.cursor = 0
}

func (c *Cmdline) cursorEnd() {
	c.cursor = len(c.cmdline)
}

func (c *Cmdline) backspace() {
	if c.cursor > 0 {
		c.cmdline = slices.Delete(c.cmdline, c.cursor-1, c.cursor)
		c.cursor--
		return
	}
	if len(c.cmdline) == 0 {
		c.eventCh <- event.Event{Type: event.ExitCmdline}
	}
}

func (c *Cmdline) deleteRune() {
	if c.cursor < len(c.cmdline) {
		c.cmdline = slices.Delete(c.cmdline, c.cursor, c.cursor+1)
	}
}

func (c *Cmdline) deleteWord() {
	i := c.cursor
	for i > 0 && unicode.IsSpace(c.cmdline[i-1]) {
		i--
	}
	if i > 0 {
		isk := isKeyword(c.cmdline[i-1])
		for i > 0 && isKeyword(c.cmdline[i-1]) == isk && !unicode.IsSpace(c.cmdline[i-1]) {
			i--
		}
	}
	c.cmdline = slices.Delete(c.cmdline, i, c.cursor)
	c.cursor = i
}

func isKeyword(c rune) bool {
	return unicode.IsDigit(c) || unicode.IsLetter(c) || c == '_'
}

func (c *Cmdline) start(typ rune, arg string) {
	c.typ = typ
	c.cmdline = []rune(arg)
	c.cursor = len(c.cmdline)
	c.history = c.histories[typ == ':']
	c.historyIndex = len(c.history)
}

func (c *Cmdline) clear() {
	c.cmdline = []rune{}
	c.cursor = 0
}

func (c *Cmdline) clearToHead() {
	c.cmdline = slices.Delete(c.cmdline, 0, c.cursor)
	c.cursor = 0
}

func (c *Cmdline) insert(ch rune) {
	if unicode.IsPrint(ch) {
		c.cmdline = slices.Insert(c.cmdline, c.cursor, ch)
		c.cursor++
	}
}

func (c *Cmdline) complete(forward bool) {
	c.cmdline = []rune(c.completor.complete(string(c.cmdline), forward))
	c.cursor = len(c.cmdline)
}

func (c *Cmdline) execute() (finish bool) {
	defer c.saveHistory()
	switch c.typ {
	case ':':
		cmd, r, bang, _, _, arg, err := parse(string(c.cmdline))
		if err != nil {
			c.eventCh <- event.Event{Type: event.Error, Error: err}
		} else if cmd.name != "" {
			c.eventCh <- event.Event{Type: cmd.eventType, Range: r, CmdName: cmd.name, Bang: bang, Arg: arg}
			finish = cmd.eventType == event.QuitAll || cmd.eventType == event.QuitErr
		}
	case '/':
		c.eventCh <- event.Event{Type: event.ExecuteSearch, Arg: string(c.cmdline), Rune: '/'}
	case '?':
		c.eventCh <- event.Event{Type: event.ExecuteSearch, Arg: string(c.cmdline), Rune: '?'}
	default:
		panic("cmdline.Cmdline.execute: unreachable")
	}
	return
}

func (c *Cmdline) saveHistory() {
	cmdline := string(c.cmdline)
	if cmdline == "" {
		return
	}
	for i, h := range c.history {
		if h == cmdline {
			c.history = slices.Delete(c.history, i, i+1)
			break
		}
	}
	c.histories[c.typ == ':'] = append(c.history, cmdline)
}

// Get returns the current state of cmdline.
func (c *Cmdline) Get() ([]rune, int, []string, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cmdline, c.cursor, c.completor.results, c.completor.index
}
