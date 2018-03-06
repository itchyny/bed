package cmdline

import (
	"unicode"

	"github.com/itchyny/bed/core"
	"github.com/itchyny/bed/util"
)

// Cmdline implements core.Cmdline
type Cmdline struct {
	cmdline   []rune
	cursor    int
	eventCh   chan<- core.Event
	cmdlineCh <-chan core.Event
	redrawCh  chan<- struct{}
}

// NewCmdline creates a new Cmdline.
func NewCmdline() *Cmdline {
	return &Cmdline{}
}

// Init initializes the Cmdline.
func (c *Cmdline) Init(eventCh chan<- core.Event, cmdlineCh <-chan core.Event, redrawCh chan<- struct{}) error {
	c.eventCh, c.cmdlineCh, c.redrawCh = eventCh, cmdlineCh, redrawCh
	return nil
}

// Run the cmdline.
func (c *Cmdline) Run() {
	for e := range c.cmdlineCh {
		switch e.Type {
		case core.EventCursorLeft:
			c.cursorLeft()
		case core.EventCursorRight:
			c.cursorRight()
		case core.EventCursorHead:
			c.cursorHead()
		case core.EventCursorEnd:
			c.cursorEnd()
		case core.EventBackspaceCmdline:
			c.backspace()
		case core.EventDeleteCmdline:
			c.deleteRune()
		case core.EventDeleteWordCmdline:
			c.deleteWord()
		case core.EventClearToHeadCmdline:
			c.clearToHead()
		case core.EventClearCmdline:
			c.clear()
		case core.EventSpaceCmdline:
			c.insert(' ')
		case core.EventRune:
			c.insert(e.Rune)
		default:
			continue
		}
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

// Get returns the current state of cmdline.
func (c *Cmdline) Get() ([]rune, int) {
	return c.cmdline, c.cursor
}

// Execute invokes the command.
func (c *Cmdline) Execute() {
	cmd, args, err := parse(c.cmdline)
	if err != nil {
		c.eventCh <- core.Event{Type: core.EventError, Error: err}
		return
	}
	if cmd.name != "" {
		c.eventCh <- core.Event{Type: cmd.eventType, CmdName: cmd.name, Args: args}
	}
}
