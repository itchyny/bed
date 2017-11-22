package core

// State defines the state of the editor.
type State struct {
	Line   int64
	Cursor int
	Bytes  []byte
	Size   int
}
