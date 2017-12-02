package core

// Event represents the event emitted by UI.
type Event struct {
	Type  EventType
	Count int64
}

// EventType ...
type EventType int

// Event types
const (
	Nop = iota
	Quit
	CursorUp
	CursorDown
	CursorLeft
	CursorRight
	CursorPrev
	CursorNext
	CursorHead
	CursorEnd
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

	StartInsert
	ExitInsert
	Insert0
	Insert1
	Insert2
	Insert3
	Insert4
	Insert5
	Insert6
	Insert7
	Insert8
	Insert9
	InsertA
	InsertB
	InsertC
	InsertD
	InsertE
	InsertF
)
