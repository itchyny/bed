package buffer

import (
	"io"
	"math"

	"github.com/itchyny/bed/util"
)

// Buffer represents a buffer.
type Buffer struct {
	rrs   []readerRange
	index int64
}

type readerRange struct {
	r    io.ReadSeeker
	min  int64
	max  int64
	diff int64
}

// NewBuffer creates a new buffer.
func NewBuffer(r io.ReadSeeker) *Buffer {
	return &Buffer{
		rrs:   []readerRange{{r: r, min: 0, max: math.MaxInt64, diff: 0}},
		index: 0,
	}
}

// Read reads bytes.
func (b *Buffer) Read(p []byte) (i int, err error) {
	for _, rr := range b.rrs {
		if b.index < rr.min {
			break
		}
		if b.index >= rr.max {
			continue
		}
		if _, err = rr.r.Seek(b.index+rr.diff, io.SeekStart); err != nil {
			return
		}
		m := int(util.MinInt64(int64(len(p)-i), rr.max-b.index))
		var k int
		if k, err = rr.r.Read(p[i : i+m]); err != nil {
			return
		}
		b.index += int64(m)
		i += k
	}
	return
}

// Seek sets the offset.
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		b.index = offset
	case io.SeekCurrent:
		b.index += offset
	case io.SeekEnd:
		var l int64
		var err error
		if l, err = b.Len(); err != nil {
			return 0, err
		}
		b.index = l + offset
	}
	return b.index, nil
}

// Len returns the total size of the buffer.
func (b *Buffer) Len() (int64, error) {
	rr := b.rrs[len(b.rrs)-1]
	l, err := rr.r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	return l - rr.diff, nil
}

// EditedIndices returns the indices of edited regions.
func (b *Buffer) EditedIndices() []int64 {
	eis := make([]int64, 0, len(b.rrs))
	for _, rr := range b.rrs {
		switch rr.r.(type) {
		case *bytesReader:
			eis = append(eis, rr.min)
			eis = append(eis, rr.max)
		}
	}
	return eis
}

// Insert inserts a byte at the specific position.
func (b *Buffer) Insert(offset int64, c byte) {
	for i, rr := range b.rrs {
		if offset >= rr.max {
			continue
		}
		if offset == rr.min && i > 0 {
			switch r := b.rrs[i-1].r.(type) {
			case *bytesReader:
				r.appendByte(c)
				b.rrs[i-1].max++
				for ; i < len(b.rrs); i++ {
					b.rrs[i].min++
					b.rrs[i].max = util.MinInt64(b.rrs[i].max, math.MaxInt64-1) + 1
					b.rrs[i].diff--
				}
				return
			}
		}
		b.rrs = append(b.rrs, readerRange{})
		b.rrs = append(b.rrs, readerRange{})
		copy(b.rrs[i+2:], b.rrs[i:])
		b.rrs[i] = readerRange{rr.r, rr.min, offset, rr.diff}
		b.rrs[i+1] = readerRange{newBytesReader([]byte{c}), offset, offset + 1, -offset}
		b.rrs[i+2] = readerRange{b.clone(rr.r), offset + 1, util.MinInt64(rr.max, math.MaxInt64-1) + 1, rr.diff - 1}
		for i = i + 3; i < len(b.rrs); i++ {
			b.rrs[i].min++
			b.rrs[i].max = util.MinInt64(b.rrs[i].max, math.MaxInt64-1) + 1
			b.rrs[i].diff--
		}
		b.cleanup()
		return
	}
	panic("Buffer#Insert: unreachable")
}

// Replace replaces a byte at the specific position.
func (b *Buffer) Replace(offset int64, c byte) {
	for i, rr := range b.rrs {
		if offset >= rr.max {
			continue
		}
		switch r := rr.r.(type) {
		case *bytesReader:
			r.replaceByte(offset+rr.diff, c)
			return
		}
		if offset == rr.min && i > 0 {
			switch r := b.rrs[i-1].r.(type) {
			case *bytesReader:
				r.appendByte(c)
				b.rrs[i-1].max++
				b.rrs[i].min++
				b.cleanup()
				return
			}
		}
		b.rrs = append(b.rrs, readerRange{})
		b.rrs = append(b.rrs, readerRange{})
		copy(b.rrs[i+2:], b.rrs[i:])
		b.rrs[i] = readerRange{rr.r, rr.min, offset, rr.diff}
		b.rrs[i+1] = readerRange{newBytesReader([]byte{c}), offset, offset + 1, -offset}
		b.rrs[i+2] = readerRange{b.clone(rr.r), offset + 1, rr.max, rr.diff}
		b.cleanup()
		return
	}
	panic("Buffer#Replace: unreachable")
}

// Delete deletes a byte at the specific position.
func (b *Buffer) Delete(offset int64) {
	for i, rr := range b.rrs {
		if offset >= rr.max {
			continue
		}
		switch r := rr.r.(type) {
		case *bytesReader:
			r.deleteByte(offset + rr.diff)
			b.rrs[i].max--
			for i++; i < len(b.rrs); i++ {
				b.rrs[i].min--
				if b.rrs[i].max != math.MaxInt64 {
					b.rrs[i].max--
				}
				b.rrs[i].diff++
			}
			b.cleanup()
			return
		}
		b.rrs = append(b.rrs, readerRange{})
		copy(b.rrs[i+1:], b.rrs[i:])
		b.rrs[i] = readerRange{rr.r, rr.min, offset, rr.diff}
		b.rrs[i+1].min = offset
		if b.rrs[i+1].max != math.MaxInt64 {
			b.rrs[i+1].max--
		}
		b.rrs[i+1].diff++
		for i += 2; i < len(b.rrs); i++ {
			b.rrs[i].min--
			if b.rrs[i].max != math.MaxInt64 {
				b.rrs[i].max--
			}
			b.rrs[i].diff++
		}
		b.cleanup()
		return
	}
	panic("Buffer#Delete: unreachable")
}

func (b *Buffer) clone(r io.ReadSeeker) io.ReadSeeker {
	switch br := r.(type) {
	case *bytesReader:
		bs := make([]byte, len(br.bs))
		copy(bs, br.bs)
		return newBytesReader(bs)
	default:
		return r
	}
}

func (b *Buffer) cleanup() {
	for i := 0; i < len(b.rrs); i++ {
		if b.rrs[i].min == b.rrs[i].max {
			copy(b.rrs[i:], b.rrs[i+1:])
			b.rrs = b.rrs[:len(b.rrs)-1]
		}
	}
	for i := 1; i < len(b.rrs); i++ {
		rr1, rr2 := b.rrs[i-1], b.rrs[i]
		switch r1 := rr1.r.(type) {
		case *bytesReader:
			switch r2 := rr2.r.(type) {
			case *bytesReader:
				r1.bs = append(r1.bs[:rr1.max+rr1.diff], r2.bs[rr2.min+rr2.diff:]...)
				b.rrs[i-1].max = b.rrs[i].max
				copy(b.rrs[i:], b.rrs[i+1:])
				b.rrs = b.rrs[:len(b.rrs)-1]
				i--
			}
		}
	}
	for i := 1; i < len(b.rrs); i++ {
		rr1, rr2 := b.rrs[i-1], b.rrs[i]
		if rr1.diff == rr2.diff && rr1.r == rr2.r && rr1.max == rr2.min {
			b.rrs[i-1].max = b.rrs[i].max
			copy(b.rrs[i:], b.rrs[i+1:])
			b.rrs = b.rrs[:len(b.rrs)-1]
		}
	}
}
