package core

// State defines the state of the editor.
type State struct {
	Line   int
	Cursor int
	Bytes  []byte
	Size   int
}
