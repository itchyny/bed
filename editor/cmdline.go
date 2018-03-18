package editor

import . "github.com/itchyny/bed/common"

// Cmdline defines the required cmdline interface for the editor.
type Cmdline interface {
	Init(chan<- Event, <-chan Event, chan<- struct{})
	Run()
	Get() ([]rune, int)
}
