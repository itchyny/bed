package searcher

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"
)

const loadSize = 1024 * 1024

// Searcher represents a searcher.
type Searcher struct {
	r       io.ReaderAt
	bytes   []byte
	loopCh  chan struct{}
	cursor  int64
	pattern string
	mu      *sync.Mutex
}

// NewSearcher creates a new searcher.
func NewSearcher(r io.ReaderAt) *Searcher {
	return &Searcher{r: r, mu: new(sync.Mutex)}
}

type errNotFound string

func (err errNotFound) Error() string {
	return "pattern not found: " + string(err)
}

// Search the pattern.
func (s *Searcher) Search(cursor int64, pattern string, forward bool) <-chan any {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bytes == nil {
		s.bytes = make([]byte, loadSize)
	}
	s.cursor, s.pattern = cursor, pattern
	ch := make(chan any)
	if forward {
		s.loop(s.forward, ch)
	} else {
		s.loop(s.backward, ch)
	}
	return ch
}

func (s *Searcher) forward() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	target, err := patternToTarget(s.pattern)
	if err != nil {
		return -1, err
	}
	base := s.cursor + 1
	n, err := s.r.ReadAt(s.bytes, base)
	if err != nil && err != io.EOF {
		return -1, err
	}
	if n == 0 {
		return -1, errNotFound(s.pattern)
	}
	if err == io.EOF {
		s.cursor += int64(n)
	} else {
		s.cursor += int64(n - len(target) + 1)
	}
	i := bytes.Index(s.bytes[:n], target)
	if i >= 0 {
		return base + int64(i), nil
	}
	return -1, nil
}

func (s *Searcher) backward() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	target, err := patternToTarget(s.pattern)
	if err != nil {
		return -1, err
	}
	base := max(0, s.cursor-int64(loadSize))
	size := int(min(int64(loadSize), s.cursor))
	n, err := s.r.ReadAt(s.bytes[:size], base)
	if err != nil && err != io.EOF {
		return -1, err
	}
	if n == 0 {
		return -1, errNotFound(s.pattern)
	}
	if s.cursor == int64(n) {
		s.cursor = 0
	} else {
		s.cursor = base + int64(len(target)-1)
	}
	i := bytes.LastIndex(s.bytes[:n], target)
	if i >= 0 {
		return base + int64(i), nil
	}
	return -1, nil
}

func (s *Searcher) loop(f func() (int64, error), ch chan<- any) {
	if s.loopCh != nil {
		close(s.loopCh)
	}
	loopCh := make(chan struct{})
	s.loopCh = loopCh
	go func() {
		defer close(ch)
		for {
			select {
			case <-loopCh:
				return
			case <-time.After(10 * time.Millisecond):
				idx, err := f()
				if err != nil {
					ch <- err
					return
				}
				if idx >= 0 {
					ch <- idx
					return
				}
			}
		}
	}()
}

// Abort the searching.
func (s *Searcher) Abort() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loopCh != nil {
		close(s.loopCh)
		s.loopCh = nil
		return errors.New("search is aborted")
	}
	return nil
}
