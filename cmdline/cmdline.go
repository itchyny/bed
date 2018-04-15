package cmdline

import (
	"sync"
	"unicode"

	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/mathutil"
)

// Cmdline implements editor.Cmdline
type Cmdline struct {
	cmdline   []rune
	cursor    int
	completor *completor
	typ       rune
	eventCh   chan<- event.Event
	cmdlineCh <-chan event.Event
	redrawCh  chan<- struct{}
	mu        *sync.Mutex
}

// NewCmdline creates a new Cmdline.
func NewCmdline() *Cmdline {
	return &Cmdline{
		completor: newCompletor(&filesystem{}),
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
			c.typ = ':'
			c.start(e.Arg)
		case event.StartCmdlineSearchForward:
			c.typ = '/'
			c.clear()
		case event.StartCmdlineSearchBackward:
			c.typ = '?'
			c.clear()
		case event.ExitCmdline:
			c.clear()
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
			c.execute()
		default:
			c.mu.Unlock()
			continue
		}
		c.completor.clear()
		c.mu.Unlock()
		c.redrawCh <- struct{}{}
	}
}

func (c *Cmdline) cursorLeft() {
	c.cursor = mathutil.MaxInt(0, c.cursor-1)
}

func (c *Cmdline) cursorRight() {
	c.cursor = mathutil.MinInt(len(c.cmdline), c.cursor+1)
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
		return
	}
	if len(c.cmdline) == 0 {
		c.eventCh <- event.Event{Type: event.ExitCmdline}
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
		for i > 0 && isKeyword(c.cmdline[i-1]) == isk && !unicode.IsSpace(c.cmdline[i-1]) {
			i--
		}
	}
	c.cmdline = append(c.cmdline[:i], c.cmdline[c.cursor:]...)
	c.cursor = i
}

func isKeyword(c rune) bool {
	return unicode.IsDigit(c) || unicode.IsLetter(c) || c == '_'
}

func (c *Cmdline) start(arg string) {
	c.cmdline = []rune(arg)
	c.cursor = len(c.cmdline)
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
	cmd, _, prefix, arg, err := parse(c.cmdline)
	if err != nil {
		c.completor.clear()
		return
	}
	c.cmdline = []rune(c.completor.complete(
		string(c.cmdline), cmd, prefix, arg, forward))
	c.cursor = len(c.cmdline)
}

func (c *Cmdline) execute() {
	switch c.typ {
	case ':':
		cmd, r, _, arg, err := parse(c.cmdline)
		if err != nil {
			c.eventCh <- event.Event{Type: event.Error, Error: err}
			return
		}
		if cmd.name != "" {
			c.eventCh <- event.Event{Type: cmd.eventType, Range: r, CmdName: cmd.name, Arg: arg}
		}
	case '/':
		c.eventCh <- event.Event{Type: event.ExecuteSearch, Arg: string(c.cmdline), Rune: '/'}
	case '?':
		c.eventCh <- event.Event{Type: event.ExecuteSearch, Arg: string(c.cmdline), Rune: '?'}
	default:
		panic("cmdline.Cmdline.execute: unreachable")
	}
}

// Get returns the current state of cmdline.
func (c *Cmdline) Get() ([]rune, int, []string, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cmdline, c.cursor, c.completor.results, c.completor.index
}
