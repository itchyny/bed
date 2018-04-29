package event

import (
	"github.com/itchyny/bed/buffer"
	"github.com/itchyny/bed/mode"
)

// Event represents the event emitted by UI.
type Event struct {
	Type    Type
	Range   *Range
	Count   int64
	Rune    rune
	CmdName string
	Arg     string
	Error   error
	Mode    mode.Mode
	Buffer  *buffer.Buffer
}

// Type ...
type Type int

// Event types
const (
	Nop Type = iota
	Redraw

	CursorUp
	CursorDown
	CursorLeft
	CursorRight
	CursorPrev
	CursorNext
	CursorHead
	CursorEnd
	CursorGoto
	ScrollUp
	ScrollDown
	ScrollTop
	ScrollTopHead
	ScrollMiddle
	ScrollMiddleHead
	ScrollBottom
	ScrollBottomHead
	PageUp
	PageDown
	PageUpHalf
	PageDownHalf
	PageTop
	PageEnd
	WindowTop
	WindowMiddle
	WindowBottom
	JumpTo
	JumpBack

	DeleteByte
	DeletePrevByte
	Increment
	Decrement
	SwitchFocus

	StartInsert
	StartInsertHead
	StartAppend
	StartAppendEnd
	StartReplaceByte
	StartReplace

	ExitInsert
	Backspace
	Delete
	Rune

	Undo
	Redo

	StartVisual
	SwitchVisualEnd
	ExitVisual

	Copy
	Cut
	Copied
	Paste
	PastePrev
	Pasted

	StartCmdlineCommand
	StartCmdlineSearchForward
	StartCmdlineSearchBackward
	BackspaceCmdline
	DeleteCmdline
	DeleteWordCmdline
	ClearToHeadCmdline
	ClearCmdline
	ExitCmdline
	CompleteForwardCmdline
	CompleteBackCmdline
	ExecuteCmdline
	ExecuteSearch
	NextSearch
	PreviousSearch

	Edit
	Enew
	New
	Vnew
	Wincmd
	FocusWindowUp
	FocusWindowDown
	FocusWindowLeft
	FocusWindowRight
	FocusWindowTopLeft
	FocusWindowBottomRight
	FocusWindowPrevious
	MoveWindowTop
	MoveWindowBottom
	MoveWindowLeft
	MoveWindowRight
	Suspend
	Quit
	QuitAll
	Write
	WriteQuit
	Info
	Error
)
