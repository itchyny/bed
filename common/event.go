package common

// Event represents the event emitted by UI.
type Event struct {
	Type    EventType
	Count   int64
	Rune    rune
	CmdName string
	Args    []string
	Error   error
	Mode    Mode
}

// EventType ...
type EventType int

// Event types
const (
	EventNop = iota

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

	EventSwitchFocus
	EventStartCmdline
	EventBackspaceCmdline
	EventDeleteCmdline
	EventDeleteWordCmdline
	EventClearToHeadCmdline
	EventClearCmdline
	EventExitCmdline
	EventExecuteCmdline
	EventEdit
	EventQuit
	EventWrite
	EventWriteQuit
	EventInfo
	EventError
)
