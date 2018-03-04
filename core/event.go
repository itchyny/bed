package core

// Event represents the event emitted by UI.
type Event struct {
	Type    EventType
	Count   int64
	Rune    rune
	CmdName string
	Args    []string
	Error   error
}

// EventType ...
type EventType int

// Event types
const (
	EventNop = iota
	EventQuit
	EventCursorUp
	EventCursorDown
	EventCursorLeft
	EventCursorRight
	EventCursorPrev
	EventCursorNext
	EventCursorHead
	EventCursorEnd
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
	EventInsert0
	EventInsert1
	EventInsert2
	EventInsert3
	EventInsert4
	EventInsert5
	EventInsert6
	EventInsert7
	EventInsert8
	EventInsert9
	EventInsertA
	EventInsertB
	EventInsertC
	EventInsertD
	EventInsertE
	EventInsertF
	EventBackspace
	EventDelete

	EventStartCmdline
	EventSpaceCmdline
	EventCursorLeftCmdline
	EventCursorRightCmdline
	EventCursorHeadCmdline
	EventCursorEndCmdline
	EventBackspaceCmdline
	EventDeleteCmdline
	EventDeleteWordCmdline
	EventClearToHeadCmdline
	EventClearCmdline
	EventExitCmdline
	EventRune
	EventExecuteCmdline
	EventWrite
	EventWriteQuit
	EventError
)
