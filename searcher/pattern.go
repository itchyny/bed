package searcher

import (
	"errors"
	"unicode/utf8"
)

func patternToTarget(pattern []byte) ([]byte, error) {
	if len(pattern) > 3 && pattern[0] == '0' {
		switch pattern[1] {
		case 'x', 'X':
			return decodeHexLiteral(pattern)
		case 'b', 'B':
			return decodeBinLiteral(pattern)
		}
	}
	return unescapePattern(pattern), nil
}

func decodeHexLiteral(pattern []byte) ([]byte, error) {
	bs := make([]byte, 0, len(pattern)/2+1)
	var c byte
	var lower bool
	for i := 2; i < len(pattern); i++ {
		if !isHex(pattern[i]) {
			return nil, errors.New("invalid hex pattern: " + string(pattern))
		}
		c = c<<4 | hexToDigit(pattern[i])
		if lower {
			bs = append(bs, c)
			c = 0
		}
		lower = !lower
	}
	if lower {
		bs = append(bs, c<<4)
	}
	return bs, nil
}

func decodeBinLiteral(pattern []byte) ([]byte, error) {
	bs := make([]byte, 0, len(pattern)/16+1)
	var c byte
	var bits int
	for i := 2; i < len(pattern); i++ {
		if !isBin(pattern[i]) {
			return nil, errors.New("invalid bin pattern: " + string(pattern))
		}
		c = c<<1 | hexToDigit(pattern[i])
		bits++
		if bits == 8 {
			bits = 0
			bs = append(bs, c)
			c = 0
		}
	}
	if bits > 0 {
		bs = append(bs, c<<uint(8-bits))
	}
	return bs, nil
}

func unescapePattern(pattern []byte) []byte {
	var escape bool
	var buf [4]byte
	bs := make([]byte, 0, len(pattern))
	for i := 0; i < len(pattern); i++ {
		b := pattern[i]
		if escape {
			switch b {
			case '0':
				bs = append(bs, 0)
			case 'a':
				bs = append(bs, '\a')
			case 'b':
				bs = append(bs, '\b')
			case 'f':
				bs = append(bs, '\f')
			case 'n':
				bs = append(bs, '\n')
			case 'r':
				bs = append(bs, '\r')
			case 't':
				bs = append(bs, '\t')
			case 'v':
				bs = append(bs, '\v')
			case 'x', 'u', 'U':
				var n int
				switch b {
				case 'x':
					n = 2
				case 'u':
					n = 4
				case 'U':
					n = 8
				}
				appended := true
				var c rune
				if i+n < len(pattern) {
					for k := 1; k <= n; k++ {
						if !isHex(pattern[i+k]) {
							appended = false
							break
						}
						c = c<<4 | rune(hexToDigit(pattern[i+k]))
					}
					if appended {
						if b == 'x' {
							bs = append(bs, byte(c))
						} else {
							n := utf8.EncodeRune(buf[:], c)
							bs = append(bs, buf[:n]...)
						}
						i += n
					}
				}
				if !appended {
					bs = append(bs, b)
				}
			default:
				bs = append(bs, b)
			}
			escape = false
		} else if b == '\\' {
			escape = true
		} else {
			bs = append(bs, b)
		}
	}
	if escape {
		bs = append(bs, '\\')
	}
	return bs
}

func isHex(b byte) bool {
	return '0' <= b && b <= '9' || 'A' <= b && b <= 'F' || 'a' <= b && b <= 'f'
}

func hexToDigit(b byte) byte {
	switch {
	case '0' <= b && b <= '9':
		return b - '0'
	case 'A' <= b && b <= 'F':
		return b - 'A' + 10
	case 'a' <= b && b <= 'f':
		return b - 'a' + 10
	default:
		return 0
	}
}

func isBin(b byte) bool {
	return b == '0' || b == '1'
}
