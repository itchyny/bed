package core

// Cmdline defines the required cmdline interface for the editor.
type Cmdline interface {
	Init(eventCh chan<- Event, cmdlineCh <-chan Event) error
	Run()
	Get() ([]rune, int)
	Execute()
}
