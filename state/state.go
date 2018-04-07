package state

import (
	"github.com/itchyny/bed/layout"
	"github.com/itchyny/bed/mode"
)

// State holds the state of the editor to display the user interface.
type State struct {
	Mode              mode.Mode
	PrevMode          mode.Mode
	WindowStates      map[int]*WindowState
	Layout            layout.Layout
	Cmdline           []rune
	CmdlineCursor     int
	CompletionResults []string
	CompletionIndex   int
	SearchMode        rune
	Error             error
	ErrorType         int
}

// WindowState holds the state of one window.
type WindowState struct {
	Name          string
	Width         int
	Offset        int64
	Cursor        int64
	Bytes         []byte
	Size          int
	Length        int64
	Mode          mode.Mode
	Pending       bool
	PendingByte   byte
	VisualStart   int64
	EditedIndices []int64
	FocusText     bool
}

// Message types
const (
	MessageInfo = iota
	MessageError
)
