package event

import "github.com/itchyny/bed/mode"

// Event represents the event emitted by UI.
type Event struct {
	Type    Type
	Count   int64
	Rune    rune
	CmdName string
	Arg     string
	Error   error
	Mode    mode.Mode
}

// Type ...
type Type int

// Event types
const (
	Nop = iota
	Redraw

	CursorUp
	CursorDown
	CursorLeft
	CursorRight
	CursorPrev
	CursorNext
	CursorHead
	CursorEnd
	CursorGotoAbs
	CursorGotoRel
	ScrollUp
	ScrollDown
	PageUp
	PageDown
	PageUpHalf
	PageDownHalf
	PageTop
	PageEnd
	JumpTo
	JumpBack

	DeleteByte
	DeletePrevByte
	Increment
	Decrement

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

	SwitchFocus
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
	Quit
	QuitAll
	Write
	WriteQuit
	Info
	Error
)
