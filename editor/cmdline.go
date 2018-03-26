package editor

import "github.com/itchyny/bed/event"

// Cmdline defines the required cmdline interface for the editor.
type Cmdline interface {
	Init(chan<- event.Event, <-chan event.Event, chan<- struct{})
	Run()
	Get() ([]rune, int, []string, int)
}
