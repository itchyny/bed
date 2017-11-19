package core

// UI defines the required user interface for the editor.
type UI interface {
	Init(ch chan<- Event) error
	Start(func(int, int) error) error
	SetLine(line int, str string) error
	Close() error
}
