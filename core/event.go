package core

// Event represents the event emitted by UI.
type Event int

// Events
const (
	CursorUp = iota
	CursorDown
	CursorLeft
	CursorRight
	CursorPrev
	CursorNext
	ScrollUp
	ScrollDown
	PageUp
	PageDown
	PageTop
	PageLast
)
