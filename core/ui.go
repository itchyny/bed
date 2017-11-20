package core

// UI defines the required user interface for the editor.
type UI interface {
	Init(ch chan<- Event) error
	Start() error
	Height() int
	SetLine(line int, str string) error
	SetCursor(cursor *Position) error
	Close() error
}
