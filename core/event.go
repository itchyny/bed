package core

// Event represents the event emitted by UI.
type Event int

const (
	ScrollDown = iota
	ScrollUp
	PageDown
	PageUp
	PageTop
	PageLast
)
