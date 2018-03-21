package window

import (
	"io"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/itchyny/bed/buffer"
	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/util"
)

type window struct {
	buffer      *buffer.Buffer
	filename    string
	name        string
	height      int64
	width       int64
	offset      int64
	cursor      int64
	length      int64
	stack       []position
	append      bool
	replaceByte bool
	extending   bool
	pending     bool
	pendingByte byte
	focusText   bool
	redrawCh    chan<- struct{}
	eventCh     chan Event
	mu          *sync.Mutex
}

type position struct {
	cursor int64
	offset int64
}

func newWindow(r io.ReadSeeker, filename string, name string, redrawCh chan<- struct{}) (*window, error) {
	buffer := buffer.NewBuffer(r)
	length, err := buffer.Len()
	if err != nil {
		return nil, err
	}
	return &window{
		buffer:   buffer,
		filename: filename,
		name:     name,
		length:   length,
		redrawCh: redrawCh,
		eventCh:  make(chan Event),
		mu:       new(sync.Mutex),
	}, nil
}

func (w *window) setSize(width, height int) {
	w.width, w.height = int64(width), int64(height)
	w.offset = w.offset / w.width * w.width
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
	w.offset = util.MinInt64(
		w.offset,
		util.MaxInt64(w.length-1-w.height*w.width+w.width, 0)/w.width*w.width,
	)
}

// Run the window.
func (w *window) Run() {
	for e := range w.eventCh {
		w.mu.Lock()
		switch e.Type {
		case EventCursorUp:
			w.cursorUp(e.Count)
		case EventCursorDown:
			w.cursorDown(e.Count)
		case EventCursorLeft:
			w.cursorLeft(e.Count)
		case EventCursorRight:
			w.cursorRight(e.Mode, e.Count)
		case EventCursorPrev:
			w.cursorPrev(e.Count)
		case EventCursorNext:
			w.cursorNext(e.Mode, e.Count)
		case EventCursorHead:
			w.cursorHead(e.Count)
		case EventCursorEnd:
			w.cursorEnd(e.Count)
		case EventCursorGotoAbs:
			w.cursorGotoAbs(e.Count)
		case EventCursorGotoRel:
			w.cursorGotoRel(e.Count)
		case EventScrollUp:
			w.scrollUp(e.Count)
		case EventScrollDown:
			w.scrollDown(e.Count)
		case EventPageUp:
			w.pageUp()
		case EventPageDown:
			w.pageDown()
		case EventPageUpHalf:
			w.pageUpHalf()
		case EventPageDownHalf:
			w.pageDownHalf()
		case EventPageTop:
			w.pageTop()
		case EventPageEnd:
			w.pageEnd()
		case EventJumpTo:
			w.jumpTo()
		case EventJumpBack:
			w.jumpBack()
		case EventDeleteByte:
			w.deleteByte(e.Count)
		case EventDeletePrevByte:
			w.deletePrevByte(e.Count)
		case EventIncrement:
			w.increment(e.Count)
		case EventDecrement:
			w.decrement(e.Count)

		case EventStartInsert:
			w.startInsert()
		case EventStartInsertHead:
			w.startInsertHead()
		case EventStartAppend:
			w.startAppend()
		case EventStartAppendEnd:
			w.startAppendEnd()
		case EventStartReplaceByte:
			w.startReplaceByte()
		case EventStartReplace:
			w.startReplace()
		case EventExitInsert:
			w.exitInsert()
		case EventRune:
			w.insertRune(e.Mode, e.Rune)
		case EventBackspace:
			w.backspace()
		case EventDelete:
			w.deleteByte(1)
		case EventSwitchFocus:
			w.focusText = !w.focusText
			if w.pending {
				w.pending = false
				w.pendingByte = '\x00'
			}
		default:
			w.mu.Unlock()
			continue
		}
		w.mu.Unlock()
		w.redrawCh <- struct{}{}
	}
}

func (w *window) readBytes(pos int64, len int) (int, []byte, error) {
	bytes := make([]byte, len)
	_, err := w.buffer.Seek(pos, io.SeekStart)
	if err != nil {
		return 0, bytes, err
	}
	n, err := w.buffer.Read(bytes)
	if err != nil && err != io.EOF {
		return 0, bytes, err
	}
	return n, bytes, nil
}

// State returns the current state of the buffer.
func (w *window) State() (*WindowState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, bytes, err := w.readBytes(w.offset, int(w.height*w.width))
	if err != nil {
		return nil, err
	}
	return &WindowState{
		Name:          w.name,
		Width:         int(w.width),
		Offset:        w.offset,
		Cursor:        w.cursor,
		Bytes:         bytes,
		Size:          n,
		Length:        w.length,
		Pending:       w.pending,
		PendingByte:   w.pendingByte,
		EditedIndices: w.buffer.EditedIndices(),
		FocusText:     w.focusText,
	}, nil
}

func (w *window) insert(offset int64, c byte) {
	w.buffer.Insert(offset, c)
}

func (w *window) replace(offset int64, c byte) {
	w.buffer.Replace(offset, c)
}

func (w *window) delete(offset int64) {
	w.buffer.Delete(offset)
}

func (w *window) cursorUp(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor/w.width) * w.width
	if w.cursor < w.offset {
		w.offset = w.cursor / w.width * w.width
	}
}

func (w *window) cursorDown(count int64) {
	w.cursor += util.MinInt64(
		util.MinInt64(
			util.MaxInt64(count, 1),
			(util.MaxInt64(w.length, 1)-1)/w.width-w.cursor/w.width,
		)*w.width,
		util.MaxInt64(w.length, 1)-1-w.cursor)
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *window) cursorLeft(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor%w.width)
	if w.append && w.extending && w.cursor < w.length-1 {
		w.append = false
		w.extending = false
		if w.length > 0 {
			w.length--
		}
	}
}

func (w *window) cursorRight(mode Mode, count int64) {
	if mode == ModeNormal {
		w.cursor += util.MinInt64(
			util.MinInt64(util.MaxInt64(count, 1), w.width-1-w.cursor%w.width),
			util.MaxInt64(w.length, 1)-1-w.cursor,
		)
	} else if !w.extending {
		w.cursor += util.MinInt64(
			util.MinInt64(util.MaxInt64(count, 1), w.width-1-w.cursor%w.width),
			w.length-w.cursor,
		)
		if w.cursor == w.length {
			w.append = true
			w.extending = true
			w.length++
		}
	}
}

func (w *window) cursorPrev(count int64) {
	w.cursor -= util.MinInt64(util.MaxInt64(count, 1), w.cursor)
	if w.cursor < w.offset {
		w.offset = w.cursor / w.width * w.width
	}
	if w.append && w.extending && w.cursor != w.length {
		w.append = false
		w.extending = false
		if w.length > 0 {
			w.length--
		}
	}
}

func (w *window) cursorNext(mode Mode, count int64) {
	if mode == ModeNormal {
		w.cursor += util.MinInt64(util.MaxInt64(count, 1), util.MaxInt64(w.length, 1)-1-w.cursor)
	} else if !w.extending {
		w.cursor += util.MinInt64(util.MaxInt64(count, 1), w.length-w.cursor)
		if w.cursor == w.length {
			w.append = true
			w.extending = true
			w.length++
		}
	}
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *window) cursorHead(_ int64) {
	w.cursor -= w.cursor % w.width
}

func (w *window) cursorEnd(count int64) {
	w.cursor = util.MinInt64(
		(w.cursor/w.width+util.MaxInt64(count, 1))*w.width-1,
		util.MaxInt64(w.length, 1)-1,
	)
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *window) cursorGotoAbs(count int64) {
	w.cursor = util.MinInt64(count, util.MaxInt64(w.length, 1)-1)
	if w.cursor < w.offset {
		w.offset = (util.MaxInt64(w.cursor/w.width, w.height/2) - w.height/2) * w.width
	} else if w.cursor >= w.offset+w.height*w.width {
		h := (util.MaxInt64(w.length, 1)+w.width-1)/w.width - w.height
		w.offset = util.MinInt64((w.cursor-w.height*w.width+w.width)/w.width+w.height/2, h) * w.width
	}
}

func (w *window) cursorGotoRel(count int64) {
	w.cursor += util.MaxInt64(util.MinInt64(count, util.MaxInt64(w.length, 1)-1-w.cursor), -w.cursor)
	if w.cursor < w.offset {
		w.offset = (util.MaxInt64(w.cursor/w.width, w.height/2) - w.height/2) * w.width
	} else if w.cursor >= w.offset+w.height*w.width {
		h := (util.MaxInt64(w.length, 1)+w.width-1)/w.width - w.height
		w.offset = util.MinInt64((w.cursor-w.height*w.width+w.width)/w.width+w.height/2, h) * w.width
	}
}

func (w *window) scrollUp(count int64) {
	w.offset -= util.MinInt64(util.MaxInt64(count, 1), w.offset/w.width) * w.width
	if w.cursor >= w.offset+w.height*w.width {
		w.cursor -= ((w.cursor-w.offset-w.height*w.width)/w.width + 1) * w.width
	}
}

func (w *window) scrollDown(count int64) {
	h := util.MaxInt64((util.MaxInt64(w.length, 1)+w.width-1)/w.width-w.height, 0)
	w.offset += util.MinInt64(util.MaxInt64(count, 1), h-w.offset/w.width) * w.width
	if w.cursor < w.offset {
		w.cursor += util.MinInt64(
			(w.offset-w.cursor+w.width-1)/w.width*w.width,
			util.MaxInt64(w.length, 1)-1-w.cursor,
		)
	}
}

func (w *window) pageUp() {
	w.offset = util.MaxInt64(w.offset-(w.height-2)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *window) pageDown() {
	offset := util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = util.MinInt64(w.offset+(w.height-2)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *window) pageUpHalf() {
	w.offset = util.MaxInt64(w.offset-util.MaxInt64(w.height/2, 1)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *window) pageDownHalf() {
	offset := util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = util.MinInt64(w.offset+util.MaxInt64(w.height/2, 1)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *window) pageTop() {
	w.offset = 0
	w.cursor = 0
}

func (w *window) pageEnd() {
	w.offset = util.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.cursor = ((util.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
}

func isDigit(b byte) bool {
	return '\x30' <= b && b <= '\x39'
}

func isWhite(b byte) bool {
	return b == '\x00' || b == '\x09' || b == '\x0a' || b == '\x0d' || b == '\x20'
}

func (w *window) jumpTo() {
	s := 50
	_, bytes, err := w.readBytes(util.MaxInt64(w.cursor-int64(s), 0), 2*s)
	if err != nil {
		return
	}
	var i, j int
	for i = s; i < 2*s && isWhite(bytes[i]); i++ {
	}
	if i == 2*s || !isDigit(bytes[i]) {
		return
	}
	for ; 0 < i && isDigit(bytes[i-1]); i-- {
	}
	for j = i; j < 2*s && isDigit(bytes[j]); j++ {
	}
	if j == 2*s {
		return
	}
	offset, _ := strconv.ParseInt(string(bytes[i:j]), 10, 64)
	if offset <= 0 || w.length <= offset {
		return
	}
	w.stack = append(w.stack, position{w.cursor, w.offset})
	w.cursor = offset
	w.offset = util.MaxInt64(offset-offset%w.width-util.MaxInt64(w.height/3, 0)*w.width, 0)
}

func (w *window) jumpBack() {
	if len(w.stack) == 0 {
		return
	}
	w.cursor = w.stack[len(w.stack)-1].cursor
	w.offset = w.stack[len(w.stack)-1].offset
	w.stack = w.stack[:len(w.stack)-1]
}

func (w *window) deleteByte(count int64) {
	if w.length == 0 {
		return
	}
	cnt := int(util.MinInt64(
		util.MinInt64(util.MaxInt64(count, 1), w.width-w.cursor%w.width),
		w.length-w.cursor,
	))
	for i := 0; i < cnt; i++ {
		w.delete(w.cursor)
		w.length--
		if w.cursor == w.length && w.cursor > 0 {
			w.cursor--
		}
	}
}

func (w *window) deletePrevByte(count int64) {
	cnt := int(util.MinInt64(util.MaxInt64(count, 1), w.cursor%w.width))
	for i := 0; i < cnt; i++ {
		w.delete(w.cursor - 1)
		w.cursor--
		w.length--
	}
}

func (w *window) increment(count int64) {
	_, bytes, err := w.readBytes(w.cursor, 1)
	if err != nil {
		return
	}
	w.replace(w.cursor, bytes[0]+byte(util.MaxInt64(count, 1)%256))
	if w.length == 0 {
		w.length++
	}
}

func (w *window) decrement(count int64) {
	_, bytes, err := w.readBytes(w.cursor, 1)
	if err != nil {
		return
	}
	w.replace(w.cursor, bytes[0]-byte(util.MaxInt64(count, 1)%256))
	if w.length == 0 {
		w.length++
	}
}

func (w *window) startInsert() {
	w.append = false
	w.extending = false
	w.pending = false
	if w.cursor == w.length {
		w.append = true
		w.extending = true
		w.length++
	}
}

func (w *window) startInsertHead() {
	w.cursorHead(0)
	w.append = false
	w.extending = false
	w.pending = false
	if w.cursor == w.length {
		w.append = true
		w.extending = true
		w.length++
	}
}

func (w *window) startAppend() {
	w.append = true
	w.extending = false
	w.pending = false
	if w.length > 0 {
		w.cursor++
	}
	if w.cursor == w.length {
		w.extending = true
		w.length++
	}
	if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
}

func (w *window) startAppendEnd() {
	w.cursorEnd(0)
	w.startAppend()
}

func (w *window) startReplaceByte() {
	w.replaceByte = true
	w.append = false
	w.extending = false
	w.pending = false
}

func (w *window) startReplace() {
	w.replaceByte = false
	w.append = false
	w.extending = false
	w.pending = false
}

func (w *window) exitInsert() {
	w.pending = false
	if w.append {
		if w.extending && w.length > 0 {
			w.length--
		}
		if w.cursor > 0 {
			w.cursor--
		}
		w.replaceByte = false
		w.append = false
		w.extending = false
		w.pending = false
	}
}

func (w *window) insertRune(mode Mode, ch rune) {
	if mode == ModeInsert || mode == ModeReplace {
		if w.focusText {
			buf := make([]byte, 4)
			n := utf8.EncodeRune(buf, ch)
			for i := 0; i < n; i++ {
				w.insertByte(mode, byte(buf[i]>>4))
				w.insertByte(mode, byte(buf[i]&0x0f))
			}
		} else if '0' <= ch && ch <= '9' {
			w.insertByte(mode, byte(ch-'0'))
		} else if 'a' <= ch && ch <= 'f' {
			w.insertByte(mode, byte(ch-'a'+0x0a))
		}
	}
}

func (w *window) insertByte(mode Mode, b byte) {
	if w.pending {
		switch mode {
		case ModeInsert:
			w.insert(w.cursor, w.pendingByte|b)
			w.cursor++
			w.length++
		case ModeReplace:
			w.replace(w.cursor, w.pendingByte|b)
			if w.length == 0 {
				w.length++
			}
			if w.replaceByte {
				w.exitInsert()
			} else {
				w.cursor++
				if w.cursor == w.length {
					w.append = true
					w.extending = true
					w.length++
				}
			}
		}
		if w.cursor >= w.offset+w.height*w.width {
			w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
		}
		w.pending = false
		w.pendingByte = '\x00'
	} else {
		w.pending = true
		w.pendingByte = b << 4
	}
}

func (w *window) backspace() {
	if w.pending {
		w.pending = false
		w.pendingByte = '\x00'
	} else if w.cursor > 0 {
		w.delete(w.cursor - 1)
		w.cursor--
		w.length--
	}
}

// Close the Window.
func (w *window) Close() {
	close(w.eventCh)
}
