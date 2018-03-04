package cmdline

import (
	"unicode"

	"github.com/itchyny/bed/core"
	"github.com/itchyny/bed/util"
)

// Cmdline implements core.Cmdline
type Cmdline struct {
	cmdline []rune
	cursor  int
	ch      chan<- core.Event
}

// NewCmdline creates a new Cmdline.
func NewCmdline() *Cmdline {
	return &Cmdline{}
}

// Init initializes the Cmdline.
func (c *Cmdline) Init(ch chan<- core.Event) error {
	c.ch = ch
	return nil
}

// CursorLeft moves the cursor left.
func (c *Cmdline) CursorLeft() {
	c.cursor = util.MaxInt(0, c.cursor-1)
}

// CursorRight moves the cursor right.
func (c *Cmdline) CursorRight() {
	c.cursor = util.MinInt(len(c.cmdline), c.cursor+1)
}

// CursorHead moves the cursor to the head.
func (c *Cmdline) CursorHead() {
	c.cursor = 0
}

// CursorEnd moves the cursor to the end.
func (c *Cmdline) CursorEnd() {
	c.cursor = len(c.cmdline)
}

// Backspace deletes one character left.
func (c *Cmdline) Backspace() {
	if c.cursor > 0 {
		c.cmdline = append(c.cmdline[:c.cursor-1], c.cmdline[c.cursor:]...)
		c.cursor--
	}
}

// Delete deletes one character right.
func (c *Cmdline) Delete() {
	if c.cursor < len(c.cmdline) {
		c.cmdline = append(c.cmdline[:c.cursor], c.cmdline[c.cursor+1:]...)
	}
}

// Clear the cmdline.
func (c *Cmdline) Clear() {
	c.cmdline = []rune{}
	c.cursor = 0
}

// ClearToHead delete all the characters left.
func (c *Cmdline) ClearToHead() {
	c.cmdline = c.cmdline[c.cursor:]
	c.cursor = 0
}

// Insert inserts one rune at the cursor.
func (c *Cmdline) Insert(ch rune) {
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
	cmdline := string(c.cmdline)
	if cmdline == "quit" {
		c.ch <- core.Event{Type: core.EventQuit}
	}
}
