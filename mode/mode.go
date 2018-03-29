package mode

// Mode ...
type Mode int

// Modes
const (
	Normal Mode = iota
	Insert
	Replace
	Visual
	Cmdline
	Search
)
