package core

// Event represents the event emitted by UI.
type Event int

// Events
const (
	ScrollUp = iota
	ScrollDown
	PageUp
	PageDown
	PageTop
	PageLast
)
