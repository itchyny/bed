package mode

// Mode ...
type Mode int

// Modes
const (
	Normal = iota
	Insert
	Replace
	Cmdline
	Search
)
