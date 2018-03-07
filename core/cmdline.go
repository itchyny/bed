package core

// Cmdline defines the required cmdline interface for the editor.
type Cmdline interface {
	Init(chan<- Event, <-chan Event, chan<- struct{}) error
	Run()
	Get() ([]rune, int)
}
