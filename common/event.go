package common

// Event represents the event emitted by UI.
type Event struct {
	Type    EventType
	Count   int64
	Rune    rune
	CmdName string
	Arg     string
	Error   error
	Mode    Mode
}

// EventType ...
type EventType int

// Event types
const (
	EventNop = iota
	EventRedraw

	EventCursorUp
	EventCursorDown
	EventCursorLeft
	EventCursorRight
	EventCursorPrev
	EventCursorNext
	EventCursorHead
	EventCursorEnd
	EventCursorGotoAbs
	EventCursorGotoRel
	EventScrollUp
	EventScrollDown
	EventPageUp
	EventPageDown
	EventPageUpHalf
	EventPageDownHalf
	EventPageTop
	EventPageEnd
	EventJumpTo
	EventJumpBack

	EventDeleteByte
	EventDeletePrevByte
	EventIncrement
	EventDecrement

	EventStartInsert
	EventStartInsertHead
	EventStartAppend
	EventStartAppendEnd
	EventStartReplaceByte
	EventStartReplace

	EventExitInsert
	EventBackspace
	EventDelete
	EventRune

	EventUndo
	EventRedo

	EventSwitchFocus
	EventStartCmdline
	EventBackspaceCmdline
	EventDeleteCmdline
	EventDeleteWordCmdline
	EventClearToHeadCmdline
	EventClearCmdline
	EventExitCmdline
	EventCompleteForwardCmdline
	EventCompleteBackCmdline
	EventExecuteCmdline

	EventEdit
	EventNew
	EventVnew
	EventWincmd
	EventFocusWindowUp
	EventFocusWindowDown
	EventFocusWindowLeft
	EventFocusWindowRight
	EventFocusWindowTopLeft
	EventFocusWindowBottomRight
	EventFocusWindowPrevious
	EventMoveWindowTop
	EventMoveWindowBottom
	EventMoveWindowLeft
	EventMoveWindowRight
	EventQuit
	EventQuitAll
	EventWrite
	EventWriteQuit
	EventInfo
	EventError
)
