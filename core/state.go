package core

// State defines the state of the editor.
type State struct {
	Line   int64
	Width  int
	Cursor int
	Bytes  []byte
	Size   int
}
