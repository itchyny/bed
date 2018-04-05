package event

import "unicode"

var states = map[int]map[rune]struct {
	position Position
	state    int
}{
	0: {
		'\'': {position: nil, state: 2},
	},
	2: {
		'<': {position: VisualStart{}, state: 1},
		'>': {position: VisualEnd{}, state: 1},
	},
}

// ParsePos parses a Position.
func ParsePos(xs []rune, i int) (Position, int) {
	var state int
	var position Position
	for ; i < len(xs); i++ {
		if state <= 1 && unicode.IsSpace(xs[i]) {
			continue
		}
		if s, ok := states[state]; ok {
			if next, ok := s[xs[i]]; ok {
				state = next.state
				position = next.position
			} else {
				return position, i
			}
		} else {
			return position, i
		}
	}
	return position, i
}
