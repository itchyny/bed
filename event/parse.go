package event

import (
	"strings"
	"unicode"
)

// ParseRange parses a Range.
func ParseRange(src string) (*Range, string) {
	var from, to Position
	from, src = parsePosition(src)
	if from == nil {
		return nil, src
	}
	var ok bool
	if src, ok = strings.CutPrefix(src, ","); !ok {
		return &Range{From: from}, src
	}
	to, src = parsePosition(src)
	return &Range{From: from, To: to}, src
}

func parsePosition(src string) (Position, string) {
	var pos Position
	var offset int64
	src = strings.TrimLeftFunc(src, unicode.IsSpace)
	if src == "" {
		return nil, src
	}
	switch src[0] {
	case '.':
		src = src[1:]
		fallthrough
	case '-', '+':
		pos = Relative{}
	case '$':
		pos = End{}
		src = src[1:]
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		offset, src = parseNum(src)
		pos = Absolute{offset}
	case '\'':
		if len(src) == 1 {
			return nil, src
		}
		switch src[1] {
		case '<':
			pos = VisualStart{}
		case '>':
			pos = VisualEnd{}
		default:
			return nil, src
		}
		src = src[2:]
	default:
		return nil, src
	}
	for src != "" {
		src = strings.TrimLeftFunc(src, unicode.IsSpace)
		if src == "" {
			break
		}
		sign := int64(1)
		switch src[0] {
		case '-':
			sign = -1
			fallthrough
		case '+':
			offset, src = parseNum(src[1:])
			pos = pos.add(sign * offset)
		default:
			return pos, src
		}
	}
	return pos, src
}

func parseNum(src string) (int64, string) {
	offset, radix, ishex := int64(0), int64(10), false
	if src, ishex = strings.CutPrefix(src, "0x"); ishex {
		radix = 16
	}
	for src != "" {
		c := src[0]
		switch {
		case '0' <= c && c <= '9':
			offset = offset*radix + int64(c-'0')
		case ('A' <= c && c <= 'F' || 'a' <= c && c <= 'f') && ishex:
			offset = offset*radix + int64(c|('a'-'A')-'a'+10)
		default:
			return offset, src
		}
		src = src[1:]
	}
	return offset, src
}
