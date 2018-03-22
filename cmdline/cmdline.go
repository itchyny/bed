package cmdline

import (
	"unicode"

	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/util"
)

// Cmdline implements editor.Cmdline
type Cmdline struct {
	cmdline           []rune
	cursor            int
	completor         *completor
	completionResults []string
	completionIndex   int
	eventCh           chan<- Event
	cmdlineCh         <-chan Event
	redrawCh          chan<- struct{}
}

// NewCmdline creates a new Cmdline.
func NewCmdline() *Cmdline {
	return &Cmdline{completor: newCompletor()}
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
		c.completor.clear()
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
		c.completor.clear()
		return
	}
	c.cmdline = []rune(c.completor.complete(string(c.cmdline), cmd, args, forward))
	c.cursor = len(c.cmdline)
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
	return c.cmdline, c.cursor, c.completor.results, c.completor.index
}
