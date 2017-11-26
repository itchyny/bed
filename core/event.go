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
)
