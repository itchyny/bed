package buffer

import (
	"errors"
	"io"
	"math"
	"sync"

	"github.com/itchyny/bed/mathutil"
)

// Buffer represents a buffer.
type Buffer struct {
	rrs    []readerRange
	index  int64
	mu     *sync.Mutex
	bytes  []byte
	offset int64
}

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

type readerRange struct {
	r    readAtSeeker
	min  int64
	max  int64
	diff int64
}

// NewBuffer creates a new buffer.
func NewBuffer(r readAtSeeker) *Buffer {
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
	index := b.index
	for _, rr := range b.rrs {
		if b.index < rr.min {
			break
		}
		if b.index >= rr.max {
			continue
		}
		m := int(mathutil.MinInt64(int64(len(p)-i), rr.max-b.index))
		var k int
		if k, err = rr.r.ReadAt(p[i:i+m], b.index+rr.diff); err != nil && k == 0 {
			break
		}
		err = nil
		b.index += int64(m)
		i += k
	}
	if len(b.bytes) > 0 {
		j := mathutil.MaxInt64(b.offset-index, 0)
		k := mathutil.MaxInt64(index-b.offset, 0)
		if j < int64(len(p)) && k < int64(len(b.bytes)) {
			if cnt := copy(p[j:], b.bytes[k:]); i < int(j)+cnt {
				i = int(j) + cnt
			}
		}
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
	return mathutil.MaxInt64(l-rr.diff, b.offset+int64(len(b.bytes))), nil
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
	if len(b.bytes) > 0 {
		eis = insertInterval(eis, b.offset, b.offset+int64(len(b.bytes)))
	}
	return eis
}

func insertInterval(xs []int64, start int64, end int64) []int64 {
	ys := make([]int64, 0, len(xs)+2)
	var inserted bool
	for i := 0; i < len(xs); i += 2 {
		if !inserted && start <= xs[i] {
			ys = append(ys, start)
			ys = append(ys, end)
			inserted = true
		}
		if len(ys) > 0 && xs[i] <= ys[len(ys)-1] {
			if ys[len(ys)-1] < xs[i+1] {
				ys[len(ys)-1] = xs[i+1]
			}
			continue
		}
		ys = append(ys, xs[i])
		ys = append(ys, xs[i+1])
		if !inserted && start <= ys[len(ys)-1] {
			if ys[len(ys)-1] < end {
				ys[len(ys)-1] = end
			}
			inserted = true
		}
	}
	if !inserted {
		ys = append(ys, start)
		ys = append(ys, end)
	}
	return ys
}

// Clone the buffer.
func (b *Buffer) Clone() *Buffer {
	b.mu.Lock()
	defer b.mu.Unlock()
	newBuf := new(Buffer)
	newBuf.rrs = make([]readerRange, len(b.rrs))
	for i, rr := range b.rrs {
		newBuf.rrs[i] = readerRange{rr.r, rr.min, rr.max, rr.diff}
	}
	newBuf.index = b.index
	newBuf.mu = new(sync.Mutex)
	if len(b.bytes) > 0 {
		newBuf.bytes = make([]byte, len(b.bytes))
		copy(newBuf.bytes, b.bytes)
	}
	newBuf.offset = b.offset
	return newBuf
}

// Copy a part of the buffer.
func (b *Buffer) Copy(start, end int64) *Buffer {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
	newBuf := new(Buffer)
	newBuf.rrs = make([]readerRange, 0, len(b.rrs)+1)
	index := start
	for _, rr := range b.rrs {
		if index < rr.min || index >= end {
			break
		}
		if index >= rr.max {
			continue
		}
		size := mathutil.MinInt64(end-index, rr.max-index)
		newBuf.rrs = append(newBuf.rrs, readerRange{rr.r, index - start, index - start + size, rr.diff + start})
		index += size
	}
	newBuf.rrs = append(newBuf.rrs, readerRange{newBytesReader(nil), index - start, math.MaxInt64, -index + start})
	newBuf.cleanup()
	newBuf.index = 0
	newBuf.mu = new(sync.Mutex)
	return newBuf
}

// Cut a part of the buffer.
func (b *Buffer) Cut(start, end int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
	rrs := make([]readerRange, 0, len(b.rrs)+1)
	var index, max int64
	for _, rr := range b.rrs {
		if start >= rr.max {
			rrs = append(rrs, rr)
			index = rr.max
			continue
		}
		if end <= rr.min {
			max = rr.max - rr.min + index
			if rr.max == math.MaxInt64 {
				max = math.MaxInt64
			}
			rrs = append(rrs, readerRange{rr.r, index, max, rr.diff - index + rr.min})
			index = max
			continue
		}
		if start >= rr.min {
			max = start
			rrs = append(rrs, readerRange{rr.r, index, max, rr.diff})
			index = max
		}
		if end < rr.max {
			max = rr.max - end + index
			if rr.max == math.MaxInt64 {
				max = math.MaxInt64
			}
			rrs = append(rrs, readerRange{rr.r, index, max, rr.diff + end - index})
			index = max
		}
	}
	if index != math.MaxInt64 {
		rrs = append(rrs, readerRange{newBytesReader(nil), index, math.MaxInt64, -index})
	}
	b.rrs = rrs
	b.index = 0
	b.cleanup()
}

// Paste a buffer into a buffer.
func (b *Buffer) Paste(offset int64, c *Buffer) {
	b.mu.Lock()
	c.mu.Lock()
	defer b.mu.Unlock()
	defer c.mu.Unlock()
	b.flush()
	rrs := make([]readerRange, 0, len(b.rrs)+len(c.rrs)+1)
	var index, max int64
	for _, rr := range b.rrs {
		if offset >= rr.max {
			rrs = append(rrs, rr)
			continue
		}
		if offset < rr.min {
			max = mathutil.MinInt64(rr.max, math.MaxInt64-index+rr.min) + index - rr.min
			rrs = append(rrs, readerRange{rr.r, index, max, rr.diff - index + rr.min})
			index = max
			continue
		}
		rrs = append(rrs, readerRange{rr.r, rr.min, offset, rr.diff})
		index = offset
		for _, rr := range c.rrs {
			if rr.max == math.MaxInt64 {
				l, _ := rr.r.Seek(0, io.SeekEnd)
				max = l + index
			} else {
				max = rr.max - rr.min + index
			}
			rrs = append(rrs, readerRange{rr.r, index, max, rr.diff - index + rr.min})
			index = max
		}
		max = mathutil.MinInt64(rr.max, math.MaxInt64-index+offset) + index - offset
		rrs = append(rrs, readerRange{rr.r, index, max, rr.diff - index + offset})
		index = max
	}
	b.rrs = rrs
	b.cleanup()
}

// Insert inserts a byte at the specific position.
func (b *Buffer) Insert(offset int64, c byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
	for i, rr := range b.rrs {
		if offset >= rr.max {
			continue
		}
		if offset == rr.min && i > 0 {
			switch r := b.rrs[i-1].r.(type) {
			case *bytesReader:
				r = r.clone()
				r.replaceByte(offset+b.rrs[i-1].diff, c)
				b.rrs[i-1].r = r
				b.rrs[i-1].max++
				for ; i < len(b.rrs); i++ {
					b.rrs[i].min++
					b.rrs[i].max = mathutil.MinInt64(b.rrs[i].max, math.MaxInt64-1) + 1
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
		b.rrs[i+2] = readerRange{rr.r, offset + 1, mathutil.MinInt64(rr.max, math.MaxInt64-1) + 1, rr.diff - 1}
		for i = i + 3; i < len(b.rrs); i++ {
			b.rrs[i].min++
			b.rrs[i].max = mathutil.MinInt64(b.rrs[i].max, math.MaxInt64-1) + 1
			b.rrs[i].diff--
		}
		b.cleanup()
		return
	}
	panic("buffer.Buffer.Insert: unreachable")
}

// Replace replaces a byte at the specific position.
// This method does not overwrite the reader ranges,
// but just append the byte to the temporary byte slice
// in order to cancel the replacement with backspace key.
func (b *Buffer) Replace(offset int64, c byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.offset+int64(len(b.bytes)) != offset {
		b.flush()
	}
	if len(b.bytes) == 0 {
		b.offset = offset
	}
	b.bytes = append(b.bytes, c)
}

// UndoReplace removes the last byte of the replacing byte slice.
func (b *Buffer) UndoReplace(offset int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.bytes) > 0 && b.offset+int64(len(b.bytes))-1 == offset {
		b.bytes = b.bytes[:len(b.bytes)-1]
	}
}

// Flush temporary bytes.
func (b *Buffer) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
}

func (b *Buffer) flush() {
	if len(b.bytes) == 0 {
		return
	}
	rrs := make([]readerRange, 0, len(b.rrs)+1)
	end := b.offset + int64(len(b.bytes))
	for _, rr := range b.rrs {
		if b.offset >= rr.max || end <= rr.min {
			rrs = append(rrs, rr)
			continue
		}
		if b.offset >= rr.min {
			if rr.min < b.offset {
				rrs = append(rrs, readerRange{rr.r, rr.min, b.offset, rr.diff})
			}
			rrs = append(rrs, readerRange{newBytesReader(b.bytes), b.offset, end, -b.offset})
		}
		if rr.max == math.MaxInt64 {
			l, _ := rr.r.Seek(0, io.SeekEnd)
			if l-rr.diff <= end {
				rrs = append(rrs, readerRange{newBytesReader(nil), end, math.MaxInt64, -end})
				continue
			}
		}
		if end < rr.max {
			rrs = append(rrs, readerRange{rr.r, end, rr.max, rr.diff})
		}
	}
	b.rrs = rrs
	b.offset = 0
	b.bytes = nil
	b.cleanup()
}

// Delete deletes a byte at the specific position.
func (b *Buffer) Delete(offset int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
	for i, rr := range b.rrs {
		if offset >= rr.max {
			continue
		}
		switch r := rr.r.(type) {
		case *bytesReader:
			r = r.clone()
			r.deleteByte(offset + rr.diff)
			b.rrs[i].r = r
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
				bs := make([]byte, int(rr1.max-rr1.min)+len(r2.bs)-int(rr2.min+rr2.diff))
				copy(bs, r1.bs[rr1.min+rr1.diff:rr1.max+rr1.diff])
				copy(bs[rr1.max-rr1.min:], r2.bs[rr2.min+rr2.diff:])
				b.rrs[i-1].r = newBytesReader(bs)
				b.rrs[i-1].max = b.rrs[i].max
				b.rrs[i-1].diff = -b.rrs[i-1].min
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
