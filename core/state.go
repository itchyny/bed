package core

// State holds the state of the editor to display the user interface.
type State struct {
	Line   int64
	Width  int
	Cursor int64
	Bytes  []byte
	Size   int
}
