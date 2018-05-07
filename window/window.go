package window

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/itchyny/bed/buffer"
	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/history"
	"github.com/itchyny/bed/mathutil"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/state"
)

type window struct {
	buffer      *buffer.Buffer
	changedTick uint64
	prevChanged bool
	history     *history.History
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
	visualStart int64
	focusText   bool
	redrawCh    chan<- struct{}
	eventCh     chan<- event.Event
	mu          *sync.Mutex
}

type position struct {
	cursor int64
	offset int64
}

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

func newWindow(r readAtSeeker, filename string, name string, eventCh chan<- event.Event, redrawCh chan<- struct{}) (*window, error) {
	buffer := buffer.NewBuffer(r)
	length, err := buffer.Len()
	if err != nil {
		return nil, err
	}
	history := history.NewHistory()
	history.Push(buffer, 0, 0)
	return &window{
		buffer:      buffer,
		history:     history,
		filename:    filename,
		name:        name,
		length:      length,
		visualStart: -1,
		redrawCh:    redrawCh,
		eventCh:     eventCh,
		mu:          new(sync.Mutex),
	}, nil
}

func (w *window) setSize(width, height int) {
	w.width, w.height = int64(width), int64(height)
	w.offset = w.offset / w.width * w.width
	if w.cursor < w.offset {
		w.offset = w.cursor / w.width * w.width
	} else if w.cursor >= w.offset+w.height*w.width {
		w.offset = (w.cursor - w.height*w.width + w.width) / w.width * w.width
	}
	w.offset = mathutil.MinInt64(
		w.offset,
		mathutil.MaxInt64(w.length-1-w.height*w.width+w.width, 0)/w.width*w.width,
	)
}

func (w *window) emit(e event.Event) {
	var newEvent event.Event
	w.mu.Lock()
	offset, cursor, changedTick := w.offset, w.cursor, w.changedTick
	switch e.Type {
	case event.CursorUp:
		w.cursorUp(e.Count)
	case event.CursorDown:
		w.cursorDown(e.Count)
	case event.CursorLeft:
		w.cursorLeft(e.Count)
	case event.CursorRight:
		w.cursorRight(e.Mode, e.Count)
	case event.CursorPrev:
		w.cursorPrev(e.Count)
	case event.CursorNext:
		w.cursorNext(e.Mode, e.Count)
	case event.CursorHead:
		w.cursorHead(e.Count)
	case event.CursorEnd:
		w.cursorEnd(e.Count)
	case event.CursorGoto:
		w.cursorGoto(e)
	case event.ScrollUp:
		w.scrollUp(e.Count)
	case event.ScrollDown:
		w.scrollDown(e.Count)
	case event.ScrollTop:
		w.scrollTop(e.Count)
	case event.ScrollTopHead:
		w.scrollTopHead(e.Count)
	case event.ScrollMiddle:
		w.scrollMiddle(e.Count)
	case event.ScrollMiddleHead:
		w.scrollMiddleHead(e.Count)
	case event.ScrollBottom:
		w.scrollBottom(e.Count)
	case event.ScrollBottomHead:
		w.scrollBottomHead(e.Count)
	case event.PageUp:
		w.pageUp()
	case event.PageDown:
		w.pageDown()
	case event.PageUpHalf:
		w.pageUpHalf()
	case event.PageDownHalf:
		w.pageDownHalf()
	case event.PageTop:
		w.pageTop()
	case event.PageEnd:
		w.pageEnd()
	case event.WindowTop:
		w.windowTop(e.Count)
	case event.WindowMiddle:
		w.windowMiddle()
	case event.WindowBottom:
		w.windowBottom(e.Count)
	case event.JumpTo:
		w.jumpTo()
	case event.JumpBack:
		w.jumpBack()

	case event.DeleteByte:
		newEvent = event.Event{Type: event.Copied, Buffer: w.deleteBytes(e.Count), Arg: "deleted"}
	case event.DeletePrevByte:
		newEvent = event.Event{Type: event.Copied, Buffer: w.deletePrevBytes(e.Count), Arg: "deleted"}
	case event.Increment:
		w.increment(e.Count)
	case event.Decrement:
		w.decrement(e.Count)

	case event.StartInsert:
		w.startInsert()
	case event.StartInsertHead:
		w.startInsertHead()
	case event.StartAppend:
		w.startAppend()
	case event.StartAppendEnd:
		w.startAppendEnd()
	case event.StartReplaceByte:
		w.startReplaceByte()
	case event.StartReplace:
		w.startReplace()
	case event.ExitInsert:
		w.exitInsert()
	case event.Rune:
		if w.insertRune(e.Mode, e.Rune) {
			newEvent = event.Event{Type: event.ExitInsert}
		}
	case event.Backspace:
		w.backspace(e.Mode)
	case event.Delete:
		w.deleteByte()
	case event.StartVisual:
		w.startVisual()
	case event.SwitchVisualEnd:
		w.switchVisualEnd()
	case event.ExitVisual:
		w.exitVisual()
	case event.SwitchFocus:
		w.focusText = !w.focusText
		if w.pending {
			w.pending = false
			w.pendingByte = '\x00'
		}
		w.changedTick++
	case event.Undo:
		if e.Mode != mode.Normal {
			panic("event.Undo should be emitted under normal mode")
		}
		w.undo(e.Count)
	case event.Redo:
		if e.Mode != mode.Normal {
			panic("event.Undo should be emitted under normal mode")
		}
		w.redo(e.Count)
	case event.Copy:
		newEvent = event.Event{Type: event.Copied, Buffer: w.copy(), Arg: "yanked"}
	case event.Cut:
		newEvent = event.Event{Type: event.Copied, Buffer: w.cut(), Arg: "deleted"}
	case event.Paste, event.PastePrev:
		newEvent = event.Event{Type: event.Pasted, Count: w.paste(e)}
	case event.ExecuteSearch:
		w.search(e.Arg, e.Rune == '/')
	case event.NextSearch:
		w.search(e.Arg, e.Rune == '/')
	case event.PreviousSearch:
		w.search(e.Arg, e.Rune != '/')
	default:
		w.mu.Unlock()
		return
	}
	changed := changedTick != w.changedTick
	if e.Type != event.Undo && e.Type != event.Redo {
		if (e.Mode == mode.Normal || e.Mode == mode.Visual) && changed || e.Type == event.ExitInsert && w.prevChanged {
			w.history.Push(w.buffer, w.offset, w.cursor)
		} else if e.Mode != mode.Normal && e.Mode != mode.Visual && w.prevChanged && !changed &&
			event.CursorUp <= e.Type && e.Type <= event.JumpBack {
			w.history.Push(w.buffer, offset, cursor)
		}
	}
	w.prevChanged = changed
	w.mu.Unlock()
	if newEvent.Type == event.Nop {
		w.redrawCh <- struct{}{}
	} else {
		w.eventCh <- newEvent
	}
}

func (w *window) readBytes(offset int64, len int) (int, []byte, error) {
	bytes := make([]byte, len)
	n, err := w.buffer.ReadAt(bytes, offset)
	if err != nil && err != io.EOF {
		return 0, bytes, err
	}
	return n, bytes, nil
}

func (w *window) writeTo(r *event.Range, dst io.Writer) (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if r == nil {
		if _, err := w.buffer.Seek(0, io.SeekStart); err != nil {
			return 0, err
		}
		return io.Copy(dst, w.buffer)
	}
	var from, to int64
	var err error
	if from, err = w.positionToOffset(r.From); err != nil {
		return 0, err
	}
	if to, err = w.positionToOffset(r.To); err != nil {
		return 0, err
	}
	if from > to {
		from, to = to, from
	}
	if _, err := w.buffer.Seek(from, io.SeekStart); err != nil {
		return 0, err
	}
	return io.Copy(dst, io.LimitReader(w.buffer, to-from+1))
}

func (w *window) positionToOffset(pos event.Position) (int64, error) {
	switch pos := pos.(type) {
	case event.Absolute:
		return mathutil.MaxInt64(
			mathutil.MinInt64(pos.Offset, mathutil.MaxInt64(w.length, 1)-1),
			0,
		), nil
	case event.Relative:
		return w.cursor + mathutil.MaxInt64(
			mathutil.MinInt64(pos.Offset, mathutil.MaxInt64(w.length, 1)-1-w.cursor),
			-w.cursor,
		), nil
	case event.End:
		return mathutil.MaxInt64(w.length, 1) - 1 + mathutil.MaxInt64(
			mathutil.MinInt64(pos.Offset, 0),
			-(mathutil.MaxInt64(w.length, 1)-1),
		), nil
	case event.VisualStart:
		if w.visualStart < 0 {
			return 0, errors.New("no visual selection found")
		}
		// TODO: save visualStart after exitting visual mode
		return w.visualStart + mathutil.MaxInt64(
			mathutil.MinInt64(pos.Offset, mathutil.MaxInt64(w.length, 1)-1-w.visualStart),
			-w.visualStart,
		), nil
	case event.VisualEnd:
		if w.visualStart < 0 {
			return 0, errors.New("no visual selection found")
		}
		return w.cursor + mathutil.MaxInt64(
			mathutil.MinInt64(pos.Offset, mathutil.MaxInt64(w.length, 1)-1-w.cursor),
			-w.cursor,
		), nil
	default:
		return 0, errors.New("invalid range")
	}
}

func (w *window) state(width, height int) (*state.WindowState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.setSize(width, height)
	n, bytes, err := w.readBytes(w.offset, int(w.height*w.width))
	if err != nil {
		return nil, err
	}
	return &state.WindowState{
		Name:          w.name,
		Width:         int(w.width),
		Offset:        w.offset,
		Cursor:        w.cursor,
		Bytes:         bytes,
		Size:          n,
		Length:        w.length,
		Pending:       w.pending,
		PendingByte:   w.pendingByte,
		VisualStart:   w.visualStart,
		EditedIndices: w.buffer.EditedIndices(),
		FocusText:     w.focusText,
	}, nil
}

func (w *window) insert(offset int64, c byte) {
	w.buffer.Insert(offset, c)
	w.changedTick++
}

func (w *window) replace(offset int64, c byte) {
	w.buffer.Replace(offset, c)
	w.changedTick++
}

func (w *window) undoReplace(offset int64) {
	w.buffer.UndoReplace(offset)
	w.changedTick++
}

func (w *window) delete(offset int64) {
	w.buffer.Delete(offset)
	w.changedTick++
}

func (w *window) undo(count int64) {
	for i := int64(0); i < mathutil.MaxInt64(count, 1); i++ {
		buffer, _, offset, cursor := w.history.Undo()
		if buffer == nil {
			return
		}
		w.buffer, w.offset, w.cursor = buffer, offset, cursor
		w.length, _ = w.buffer.Len()
	}
}

func (w *window) redo(count int64) {
	for i := int64(0); i < mathutil.MaxInt64(count, 1); i++ {
		buffer, offset, cursor := w.history.Redo()
		if buffer == nil {
			return
		}
		w.buffer, w.offset, w.cursor = buffer, offset, cursor
		w.length, _ = w.buffer.Len()
	}
}

func (w *window) cursorUp(count int64) {
	w.cursor -= mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.cursor/w.width) * w.width
	if w.append && w.extending && w.cursor < w.length-1 {
		w.append = false
		w.extending = false
		if w.length > 0 {
			w.length--
		}
	}
}

func (w *window) cursorDown(count int64) {
	w.cursor += mathutil.MinInt64(
		mathutil.MinInt64(
			mathutil.MaxInt64(count, 1),
			(mathutil.MaxInt64(w.length, 1)-1)/w.width-w.cursor/w.width,
		)*w.width,
		mathutil.MaxInt64(w.length, 1)-1-w.cursor)
}

func (w *window) cursorLeft(count int64) {
	w.cursor -= mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.cursor%w.width)
	if w.append && w.extending && w.cursor < w.length-1 {
		w.append = false
		w.extending = false
		if w.length > 0 {
			w.length--
		}
	}
}

func (w *window) cursorRight(m mode.Mode, count int64) {
	if m != mode.Insert {
		w.cursor += mathutil.MinInt64(
			mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.width-1-w.cursor%w.width),
			mathutil.MaxInt64(w.length, 1)-1-w.cursor,
		)
	} else if !w.extending {
		w.cursor += mathutil.MinInt64(
			mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.width-1-w.cursor%w.width),
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
	w.cursor -= mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.cursor)
	if w.append && w.extending && w.cursor != w.length {
		w.append = false
		w.extending = false
		if w.length > 0 {
			w.length--
		}
	}
}

func (w *window) cursorNext(m mode.Mode, count int64) {
	if m != mode.Insert {
		w.cursor += mathutil.MinInt64(mathutil.MaxInt64(count, 1), mathutil.MaxInt64(w.length, 1)-1-w.cursor)
	} else if !w.extending {
		w.cursor += mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.length-w.cursor)
		if w.cursor == w.length {
			w.append = true
			w.extending = true
			w.length++
		}
	}
}

func (w *window) cursorHead(_ int64) {
	w.cursor -= w.cursor % w.width
}

func (w *window) cursorEnd(count int64) {
	w.cursor = mathutil.MinInt64(
		(w.cursor/w.width+mathutil.MaxInt64(count, 1))*w.width-1,
		mathutil.MaxInt64(w.length, 1)-1,
	)
}

func (w *window) cursorGoto(e event.Event) {
	if e.Range != nil {
		if e.Range.To != nil {
			w.cursorGotoPos(e.Range.To)
		} else if e.Range.From != nil {
			w.cursorGotoPos(e.Range.From)
		}
	}
}

func (w *window) cursorGotoPos(pos event.Position) {
	if offset, err := w.positionToOffset(pos); err == nil {
		w.cursor = mathutil.MaxInt64(mathutil.MinInt64(offset, mathutil.MaxInt64(w.length, 1)-1), 0)
		if w.cursor < w.offset {
			w.offset = (mathutil.MaxInt64(w.cursor/w.width, w.height/2) - w.height/2) * w.width
		} else if w.cursor >= w.offset+w.height*w.width {
			h := (mathutil.MaxInt64(w.length, 1)+w.width-1)/w.width - w.height
			w.offset = mathutil.MinInt64((w.cursor-w.height*w.width+w.width)/w.width+w.height/2, h) * w.width
		}
	}
}

func (w *window) scrollUp(count int64) {
	w.offset -= mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.offset/w.width) * w.width
	if w.cursor >= w.offset+w.height*w.width {
		w.cursor -= ((w.cursor-w.offset-w.height*w.width)/w.width + 1) * w.width
	}
}

func (w *window) scrollDown(count int64) {
	h := mathutil.MaxInt64((mathutil.MaxInt64(w.length, 1)+w.width-1)/w.width-w.height, 0)
	w.offset += mathutil.MinInt64(mathutil.MaxInt64(count, 1), h-w.offset/w.width) * w.width
	if w.cursor < w.offset {
		w.cursor += mathutil.MinInt64(
			(w.offset-w.cursor+w.width-1)/w.width*w.width,
			mathutil.MaxInt64(w.length, 1)-1-w.cursor,
		)
	}
}

func (w *window) scrollTop(count int64) {
	if count > 0 {
		w.cursor = mathutil.MinInt64(
			mathutil.MinInt64(
				count*w.width+w.cursor%w.width,
				(mathutil.MaxInt64(w.length, 1)-1)/w.width*w.width+w.cursor%w.width,
			),
			mathutil.MaxInt64(w.length, 1)-1,
		)
	}
	w.offset = w.cursor / w.width * w.width
}

func (w *window) scrollTopHead(count int64) {
	w.cursorHead(0)
	w.scrollTop(count)
}

func (w *window) scrollMiddle(count int64) {
	if count > 0 {
		w.cursor = mathutil.MinInt64(
			mathutil.MinInt64(
				count*w.width+w.cursor%w.width,
				(mathutil.MaxInt64(w.length, 1)-1)/w.width*w.width+w.cursor%w.width,
			),
			mathutil.MaxInt64(w.length, 1)-1,
		)
	}
	w.offset = mathutil.MaxInt64(w.cursor/w.width-w.height/2, 0) * w.width
}

func (w *window) scrollMiddleHead(count int64) {
	w.cursorHead(0)
	w.scrollMiddle(count)
}

func (w *window) scrollBottom(count int64) {
	if count > 0 {
		w.cursor = mathutil.MinInt64(
			mathutil.MinInt64(
				count*w.width+w.cursor%w.width,
				(mathutil.MaxInt64(w.length, 1)-1)/w.width*w.width+w.cursor%w.width,
			),
			mathutil.MaxInt64(w.length, 1)-1,
		)
	}
	w.offset = mathutil.MaxInt64(w.cursor/w.width-w.height, 0) * w.width
}

func (w *window) scrollBottomHead(count int64) {
	w.cursorHead(0)
	w.scrollBottom(count)
}

func (w *window) pageUp() {
	w.offset = mathutil.MaxInt64(w.offset-(w.height-2)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *window) pageDown() {
	offset := mathutil.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = mathutil.MinInt64(w.offset+(w.height-2)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((mathutil.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *window) pageUpHalf() {
	w.offset = mathutil.MaxInt64(w.offset-mathutil.MaxInt64(w.height/2, 1)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *window) pageDownHalf() {
	offset := mathutil.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = mathutil.MinInt64(w.offset+mathutil.MaxInt64(w.height/2, 1)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((mathutil.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *window) pageTop() {
	w.offset = 0
	w.cursor = 0
}

func (w *window) pageEnd() {
	w.offset = mathutil.MaxInt64(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.cursor = ((mathutil.MaxInt64(w.length, 1)+w.width-1)/w.width - 1) * w.width
}

func (w *window) windowTop(count int64) {
	w.cursor = (w.offset/w.width + mathutil.MinInt64(
		mathutil.MinInt64(mathutil.MaxInt64(count, 1)-1, (w.length-w.offset)/w.width),
		mathutil.MaxInt64(w.height, 1)-1,
	)) * w.width
}

func (w *window) windowMiddle() {
	h := mathutil.MinInt64((w.length-w.offset)/w.width, mathutil.MaxInt64(w.height, 1)-1)
	w.cursor = (w.offset/w.width + h/2) * w.width
}

func (w *window) windowBottom(count int64) {
	h := mathutil.MinInt64((w.length-w.offset)/w.width, mathutil.MaxInt64(w.height, 1)-1)
	w.cursor = (w.offset/w.width + h - mathutil.MinInt64(h, mathutil.MaxInt64(count, 1)-1)) * w.width
}

func isDigit(b byte) bool {
	return '\x30' <= b && b <= '\x39'
}

func isWhite(b byte) bool {
	return b == '\x00' || b == '\x09' || b == '\x0a' || b == '\x0d' || b == '\x20'
}

func (w *window) jumpTo() {
	s := 50
	_, bytes, err := w.readBytes(mathutil.MaxInt64(w.cursor-int64(s), 0), 2*s)
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
	w.offset = mathutil.MaxInt64(offset-offset%w.width-mathutil.MaxInt64(w.height/3, 0)*w.width, 0)
}

func (w *window) jumpBack() {
	if len(w.stack) == 0 {
		return
	}
	w.cursor = w.stack[len(w.stack)-1].cursor
	w.offset = w.stack[len(w.stack)-1].offset
	w.stack = w.stack[:len(w.stack)-1]
}

func (w *window) deleteBytes(count int64) *buffer.Buffer {
	if w.length == 0 {
		return nil
	}
	count = mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.length-w.cursor)
	b := w.buffer.Copy(w.cursor, w.cursor+count)
	w.buffer.Cut(w.cursor, w.cursor+count)
	w.length, _ = w.buffer.Len()
	w.cursor = mathutil.MinInt64(w.cursor, mathutil.MaxInt64(w.length, 1)-1)
	w.changedTick++
	return b
}

func (w *window) deletePrevBytes(count int64) *buffer.Buffer {
	if w.cursor == 0 {
		return nil
	}
	count = mathutil.MinInt64(mathutil.MaxInt64(count, 1), w.cursor)
	b := w.buffer.Copy(w.cursor-count, w.cursor)
	w.buffer.Cut(w.cursor-count, w.cursor)
	w.length, _ = w.buffer.Len()
	w.cursor -= count
	w.changedTick++
	return b
}

func (w *window) increment(count int64) {
	_, bytes, err := w.readBytes(w.cursor, 1)
	if err != nil {
		return
	}
	w.replace(w.cursor, bytes[0]+byte(mathutil.MaxInt64(count, 1)%256))
	if w.length == 0 {
		w.length++
	}
}

func (w *window) decrement(count int64) {
	_, bytes, err := w.readBytes(w.cursor, 1)
	if err != nil {
		return
	}
	w.replace(w.cursor, bytes[0]-byte(mathutil.MaxInt64(count, 1)%256))
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
	w.append = true
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
	w.buffer.Flush()
}

func (w *window) insertRune(m mode.Mode, ch rune) (exitInsert bool) {
	if m == mode.Insert || m == mode.Replace {
		if w.focusText {
			buf := make([]byte, 4)
			n := utf8.EncodeRune(buf, ch)
			for i := 0; i < n; i++ {
				exitInsert = exitInsert || w.insertByte(m, byte(buf[i]>>4))
				exitInsert = exitInsert || w.insertByte(m, byte(buf[i]&0x0f))
			}
		} else if '0' <= ch && ch <= '9' {
			exitInsert = w.insertByte(m, byte(ch-'0'))
		} else if 'a' <= ch && ch <= 'f' {
			exitInsert = w.insertByte(m, byte(ch-'a'+0x0a))
		}
	}
	return
}

func (w *window) insertByte(m mode.Mode, b byte) bool {
	if w.pending {
		switch m {
		case mode.Insert:
			w.insert(w.cursor, w.pendingByte|b)
			w.cursor++
			w.length++
		case mode.Replace:
			w.replace(w.cursor, w.pendingByte|b)
			if w.length == 0 {
				w.length++
			}
			if w.replaceByte {
				w.exitInsert()
				return true
			}
			w.cursor++
			if w.cursor == w.length {
				w.append = true
				w.extending = true
				w.length++
			}
		}
		w.pending = false
		w.pendingByte = '\x00'
	} else {
		w.pending = true
		w.pendingByte = b << 4
	}
	return false
}

func (w *window) backspace(m mode.Mode) {
	if w.pending {
		w.pending = false
		w.pendingByte = '\x00'
	} else if m == mode.Replace {
		if w.cursor > 0 {
			w.cursor--
			w.undoReplace(w.cursor)
		}
	} else if w.cursor > 0 {
		w.delete(w.cursor - 1)
		w.cursor--
		w.length--
	}
}

func (w *window) deleteByte() {
	if w.length == 0 {
		return
	}
	w.delete(w.cursor)
	w.length--
	if w.cursor == w.length && w.cursor > 0 {
		w.cursor--
	}
}

func (w *window) startVisual() {
	w.visualStart = w.cursor
}

func (w *window) switchVisualEnd() {
	if w.visualStart < 0 {
		panic("window#switchVisualEnd should be called in visual mode")
	}
	w.cursor, w.visualStart = w.visualStart, w.cursor
}

func (w *window) exitVisual() {
	w.visualStart = -1
}

func (w *window) copy() *buffer.Buffer {
	if w.visualStart < 0 {
		panic("window#copy should be called in visual mode")
	}
	start, end := w.visualStart, w.cursor
	if start > end {
		start, end = end, start
	}
	if end == w.length {
		return nil
	}
	w.visualStart = -1
	w.cursor = start
	return w.buffer.Copy(start, end+1)
}

func (w *window) cut() *buffer.Buffer {
	if w.visualStart < 0 {
		panic("window#cut should be called in visual mode")
	}
	start, end := w.visualStart, w.cursor
	if start > end {
		start, end = end, start
	}
	if end == w.length {
		return nil
	}
	w.visualStart = -1
	b := w.buffer.Copy(start, end+1)
	w.buffer.Cut(start, end+1)
	w.length, _ = w.buffer.Len()
	w.cursor = mathutil.MinInt64(start, mathutil.MaxInt64(w.length, 1)-1)
	w.changedTick++
	return b
}

func (w *window) paste(e event.Event) int64 {
	count := mathutil.MaxInt64(e.Count, 1)
	pos := w.cursor
	if e.Type != event.PastePrev {
		pos = mathutil.MinInt64(w.cursor+1, w.length)
	}
	for i := int64(0); i < count; i++ {
		w.buffer.Paste(pos, e.Buffer)
	}
	l, _ := e.Buffer.Len()
	w.length, _ = w.buffer.Len()
	w.cursor = mathutil.MinInt64(mathutil.MaxInt64(pos+l*count-1, 0), mathutil.MaxInt64(w.length, 1)-1)
	w.changedTick++
	return l * count
}

func (w *window) search(str string, forward bool) {
	if forward {
		w.searchForward(str)
	} else {
		w.searchBackward(str)
	}
}

func (w *window) searchForward(str string) {
	target := []byte(str)
	base, size := w.cursor+1, mathutil.MaxInt(int(w.height*w.width)*50, len(target)*500)
	_, bs, err := w.readBytes(base, size)
	if err != nil {
		return
	}
	i := bytes.Index(bs, target)
	if i >= 0 {
		w.cursor = base + int64(i)
		if w.cursor >= w.offset+w.height*w.width {
			w.offset = (w.cursor - w.height*w.width + w.width + 1) / w.width * w.width
		}
	}
}

func (w *window) searchBackward(str string) {
	target := []byte(str)
	size := mathutil.MaxInt(int(w.height*w.width)*50, len(target)*500)
	base := mathutil.MaxInt64(0, w.cursor-int64(size))
	_, bs, err := w.readBytes(base, int(mathutil.MinInt64(int64(size), w.cursor)))
	if err != nil {
		return
	}
	i := bytes.LastIndex(bs, target)
	if i >= 0 {
		w.cursor = base + int64(i)
	}
}
