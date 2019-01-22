package searcher

import "unicode/utf8"

func patternToTarget(pattern []byte) []byte {
	var escape bool
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
							buf := make([]byte, 4)
							n := utf8.EncodeRune(buf, c)
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
