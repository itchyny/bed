package buffer

import (
	"errors"
	"io"
	"math"
	"sync"

	"github.com/itchyny/bed/util"
)

// Buffer represents a buffer.
type Buffer struct {
	rrs   []readerRange
	index int64
	mu    *sync.Mutex
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
		mu:    new(sync.Mutex),
	}
}

// Read reads bytes.
func (b *Buffer) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.read(p)
}

func (b *Buffer) read(p []byte) (i int, err error) {
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
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.seek(offset, whence)
}

func (b *Buffer) seek(offset int64, whence int) (int64, error) {
	var index int64
	switch whence {
	case io.SeekStart:
		index = offset
	case io.SeekCurrent:
		index = b.index + offset
	case io.SeekEnd:
		var l int64
		var err error
		if l, err = b.len(); err != nil {
			return 0, err
		}
		index = l + offset
	default:
		return 0, errors.New("buffer.Buffer.Seek: invalid whence")
	}
	if index < 0 {
		return 0, errors.New("buffer.Buffer.Seek: negative position")
	}
	b.index = index
	return index, nil
}

// Len returns the total size of the buffer.
func (b *Buffer) Len() (int64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.len()
}

func (b *Buffer) len() (int64, error) {
	rr := b.rrs[len(b.rrs)-1]
	l, err := rr.r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	return l - rr.diff, nil
}

// ReadAt reads bytes at the specific offset.
func (b *Buffer) ReadAt(p []byte, offset int64) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, err := b.seek(offset, io.SeekStart); err != nil {
		return 0, err
	}
	return b.read(p)
}

// EditedIndices returns the indices of edited regions.
func (b *Buffer) EditedIndices() []int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
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

// Clone the buffer.
func (b *Buffer) Clone() *Buffer {
	b.mu.Lock()
	defer b.mu.Unlock()
	newBuf := new(Buffer)
	newBuf.rrs = make([]readerRange, len(b.rrs))
	for i, rr := range b.rrs {
		newBuf.rrs[i] = readerRange{b.clone(rr.r), rr.min, rr.max, rr.diff}
	}
	newBuf.index = b.index
	newBuf.mu = new(sync.Mutex)
	return newBuf
}

// Insert inserts a byte at the specific position.
func (b *Buffer) Insert(offset int64, c byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
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
	panic("buffer.Buffer.Insert: unreachable")
}

// Replace replaces a byte at the specific position.
func (b *Buffer) Replace(offset int64, c byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
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
	panic("buffer.Buffer.Replace: unreachable")
}

// Delete deletes a byte at the specific position.
func (b *Buffer) Delete(offset int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
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
	panic("buffer.Buffer.Delete: unreachable")
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
