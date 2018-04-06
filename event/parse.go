package event

import "unicode"

// ParseRange parses a Range.
func ParseRange(xs []rune, i int) (*Range, int) {
	from, i := ParsePos(xs, i)
	if from == nil {
		return nil, i
	}
	if i >= len(xs) || xs[i] != ',' {
		return &Range{From: from}, i
	}
	to, i := ParsePos(xs, i+1)
	return &Range{From: from, To: to}, i
}

var states = map[int]map[rune]struct {
	position Position
	state    int
}{
	0: {
		'$':  {position: End{}, state: 1},
		'.':  {position: Relative{}, state: 1},
		'\'': {position: nil, state: 2},
	},
	2: {
		'<': {position: VisualStart{}, state: 1},
		'>': {position: VisualEnd{}, state: 1},
	},
}

// ParsePos parses a Position.
//    +---- num.. ----+
//    +-- [-+]num.. --+   +---------------+
//    +------ $ ------+   |               |
// ---+------ . ------+---+-- [-+]num.. --+---
//    +-- ' -+- < -+--+
//           +- > -+
func ParsePos(xs []rune, i int) (Position, int) {
	var state int
	var position Position
	for ; i < len(xs); i++ {
		if state <= 1 && unicode.IsSpace(xs[i]) {
			continue
		}
		if state == 0 && '0' <= xs[i] && xs[i] <= '9' {
			var offset int64
			offset, i = parseNum(xs, i)
			if position == nil {
				position = Absolute{offset}
			}
			state = 1
			continue
		}
		if state <= 1 && (xs[i] == '+' || xs[i] == '-') {
			var offset int64
			sign := int64(1)
			if xs[i] == '-' {
				sign = -1
			}
			offset, i = parseNum(xs, i+1)
			offset *= sign
			if position == nil {
				position = Relative{offset}
			} else {
				position = position.addOffset(offset)
			}
			state = 1
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

func parseNum(xs []rune, i int) (int64, int) {
	var offset int64
	var hex int
	for ; i < len(xs); i++ {
		c := xs[i]
		if hex == 0 && c == '0' {
			hex = 1
		} else if hex == 1 && c == 'x' {
			hex = 2
		} else if '0' <= c && c <= '9' || hex == 2 && 'a' <= c && c <= 'f' {
			if hex == 2 {
				offset *= 0x10
			} else {
				hex = 3
				offset *= 10
			}
			if '0' <= c && c <= '9' {
				offset += int64(c - '0')
			} else {
				offset += int64(c - 'a' + 0x0a)
			}
		} else {
			return offset, i - 1
		}
	}
	return offset, i - 1
}
