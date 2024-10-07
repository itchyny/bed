package buffer

import (
	"errors"
	"io"
	"math"
	"slices"
	"sync"
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
		m := int(min(int64(len(p)-i), rr.max-b.index))
		var k int
		if k, err = rr.r.ReadAt(p[i:i+m], b.index+rr.diff); err != nil && k == 0 {
			break
		}
		err = nil
		b.index += int64(m)
		i += k
	}
	if len(b.bytes) > 0 {
		j, k := max(b.offset-index, 0), max(index-b.offset, 0)
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
	return max(l-rr.diff, b.offset+int64(len(b.bytes))), nil
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
		case *bytesReader, constReader:
			// constReader can be adjacent to another bytesReader or constReader.
			if l := len(eis); l > 0 && eis[l-1] == rr.min {
				eis[l-1] = rr.max
				continue
			}
			eis = append(eis, rr.min, rr.max)
		}
	}
	if len(b.bytes) > 0 {
		eis = insertInterval(eis, b.offset, b.offset+int64(len(b.bytes)))
	}
	return eis
}

func insertInterval(xs []int64, start int64, end int64) []int64 {
	i, fi := slices.BinarySearch(xs, start)
	j, fj := slices.BinarySearch(xs, end)
	if i%2 == 0 {
		if i == j && !fi && !fj {
			return slices.Insert(xs, i, start, end)
		}
		xs[i] = start
		i++
	}
	if j%2 == 0 {
		if fj {
			j++
		} else {
			j--
			xs[j] = end
		}
	}
	return slices.Delete(xs, i, j)
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
	newBuf.bytes = slices.Clone(b.bytes)
	newBuf.offset = b.offset
	return newBuf
}

// Copy a part of the buffer.
func (b *Buffer) Copy(start, end int64) *Buffer {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
	newBuf := new(Buffer)
	rrs := make([]readerRange, 0, len(b.rrs)+1)
	index := start
	for _, rr := range b.rrs {
		if index < rr.min || index >= end {
			break
		}
		if index >= rr.max {
			continue
		}
		size := min(end-index, rr.max-index)
		rrs = append(rrs, readerRange{rr.r, index - start, index - start + size, rr.diff + start})
		index += size
	}
	newBuf.rrs = append(rrs, readerRange{newBytesReader(nil), index - start, math.MaxInt64, -index + start})
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
			if rr.max == math.MaxInt64 {
				max = math.MaxInt64
			} else {
				max = rr.max - rr.min + index
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
			if rr.max == math.MaxInt64 {
				max = math.MaxInt64
			} else {
				max = rr.max - end + index
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
			if rr.max == math.MaxInt64 {
				max = math.MaxInt64
			} else {
				max = rr.max - rr.min + index
			}
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
		if rr.max == math.MaxInt64 {
			max = math.MaxInt64
		} else {
			max = rr.max - offset + index
		}
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
		if offset > rr.max {
			continue
		}
		var r *bytesReader
		var ok bool
		if rr.max != math.MaxInt64 {
			if r, ok = rr.r.(*bytesReader); ok {
				r = r.clone()
				r.insert(offset+rr.diff, c)
				b.rrs[i], i = readerRange{r, rr.min, rr.max + 1, rr.diff}, i+1
			}
		}
		if !ok {
			b.rrs = append(b.rrs, readerRange{}, readerRange{})
			copy(b.rrs[i+2:], b.rrs[i:])
			b.rrs[i], i = readerRange{rr.r, rr.min, offset, rr.diff}, i+1
			b.rrs[i], i = readerRange{newBytesReader([]byte{c}), offset, offset + 1, -offset}, i+1
			b.rrs[i].min = offset
		}
		for ; i < len(b.rrs); i++ {
			b.rrs[i].min++
			if b.rrs[i].max != math.MaxInt64 {
				b.rrs[i].max++
			}
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

// ReplaceIn replaces bytes within a specific range.
func (b *Buffer) ReplaceIn(start, end int64, c byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	rrs := make([]readerRange, 0, len(b.rrs)+1)
	for _, rr := range b.rrs {
		if rr.max <= start || end <= rr.min {
			rrs = append(rrs, rr)
			continue
		}
		if start > rr.min {
			rrs = append(rrs, readerRange{rr.r, rr.min, start, rr.diff})
		}
		if start >= rr.min {
			rrs = append(rrs, readerRange{constReader(c), start, end, -start})
		}
		if end < rr.max {
			rrs = append(rrs, readerRange{rr.r, end, rr.max, rr.diff})
		}
	}
	b.rrs = rrs
	b.cleanup()
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
		if r, ok := rr.r.(*bytesReader); ok {
			r = r.clone()
			r.delete(offset + rr.diff)
			b.rrs[i] = readerRange{r, rr.min, rr.max - 1, rr.diff}
		} else {
			b.rrs = append(b.rrs, readerRange{})
			copy(b.rrs[i+1:], b.rrs[i:])
			b.rrs[i] = readerRange{rr.r, rr.min, offset, rr.diff}
			b.rrs[i+1] = readerRange{rr.r, offset + 1, rr.max, rr.diff}
		}
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
	panic("buffer.Buffer.Delete: unreachable")
}

func (b *Buffer) cleanup() {
	for i := 0; i < len(b.rrs); i++ {
		if rr := b.rrs[i]; rr.min == rr.max {
			b.rrs = slices.Delete(b.rrs, i, i+1)
		}
	}
	for i := len(b.rrs) - 1; i > 0; i-- {
		rr1, rr2 := b.rrs[i-1], b.rrs[i]
		switch r1 := rr1.r.(type) {
		case constReader:
			if r1 == rr2.r {
				b.rrs[i-1].max = rr2.max
				b.rrs = slices.Delete(b.rrs, i, i+1)
			}
		case *bytesReader:
			if r2, ok := rr2.r.(*bytesReader); ok {
				bs := make([]byte, int(rr1.max-rr1.min)+len(r2.bs)-int(rr2.min+rr2.diff))
				copy(bs, r1.bs[rr1.min+rr1.diff:rr1.max+rr1.diff])
				copy(bs[rr1.max-rr1.min:], r2.bs[rr2.min+rr2.diff:])
				b.rrs[i-1] = readerRange{newBytesReader(bs), rr1.min, rr2.max, -rr1.min}
				b.rrs = slices.Delete(b.rrs, i, i+1)
			}
		default:
			if r1 == rr2.r && rr1.diff == rr2.diff && rr1.max == rr2.min {
				b.rrs[i-1].max = rr2.max
				b.rrs = slices.Delete(b.rrs, i, i+1)
			}
		}
	}
}
