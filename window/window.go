package window

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/itchyny/bed/buffer"
	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/history"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/searcher"
	"github.com/itchyny/bed/state"
)

type window struct {
	buffer           *buffer.Buffer
	changedTick      uint64
	prevChanged      bool
	maxChangedTick   uint64
	savedChangedTick uint64
	history          *history.History
	searcher         *searcher.Searcher
	searchTick       uint64
	path             string
	name             string
	height           int64
	width            int64
	offset           int64
	cursor           int64
	length           int64
	stack            []position
	append           bool
	replaceByte      bool
	extending        bool
	pending          bool
	pendingByte      byte
	visualStart      int64
	focusText        bool
	buf              []byte
	buf1             [1]byte
	redrawCh         chan<- struct{}
	eventCh          chan<- event.Event
	mu               *sync.Mutex
}

type position struct {
	cursor int64
	offset int64
}

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

func newWindow(
	r readAtSeeker, path string, name string,
	eventCh chan<- event.Event, redrawCh chan<- struct{},
) (*window, error) {
	buffer := buffer.NewBuffer(r)
	length, err := buffer.Len()
	if err != nil {
		return nil, err
	}
	history := history.NewHistory()
	history.Push(buffer, 0, 0, 0)
	return &window{
		buffer:      buffer,
		history:     history,
		searcher:    searcher.NewSearcher(r),
		path:        path,
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
	w.offset = min(
		w.offset,
		max(w.length-1-w.height*w.width+w.width, 0)/w.width*w.width,
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
	case event.ShiftLeft:
		w.shiftLeft(e.Count)
	case event.ShiftRight:
		w.shiftRight(e.Count)
	case event.ShowBinary:
		if str := w.showBinary(); str != "" {
			newEvent = event.Event{Type: event.Info, Error: errors.New(str)}
		}
	case event.ShowDecimal:
		if str := w.showDecimal(); str != "" {
			newEvent = event.Event{Type: event.Info, Error: errors.New(str)}
		}

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
	case event.AbortSearch:
		w.abortSearch()
	default:
		w.mu.Unlock()
		return
	}
	changed := changedTick != w.changedTick
	if e.Type != event.Undo && e.Type != event.Redo {
		if (e.Mode == mode.Normal || e.Mode == mode.Visual) && changed || e.Type == event.ExitInsert && w.prevChanged {
			w.history.Push(w.buffer, w.offset, w.cursor, w.changedTick)
		} else if e.Mode != mode.Normal && e.Mode != mode.Visual && w.prevChanged && !changed &&
			event.CursorUp <= e.Type && e.Type <= event.JumpBack {
			w.history.Push(w.buffer, offset, cursor, w.changedTick)
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

func (w *window) readByte(offset int64) (byte, error) {
	n, err := w.buffer.ReadAt(w.buf1[:], offset)
	if err != nil && err != io.EOF {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}
	return w.buf1[0], nil
}

func (w *window) readBytes(offset int64, l int) (int, []byte, error) {
	var reused bool
	if l <= cap(w.buf) {
		w.buf, reused = w.buf[:l], true
	} else {
		w.buf = make([]byte, l)
	}
	n, err := w.buffer.ReadAt(w.buf, offset)
	if err != nil && err != io.EOF {
		return 0, w.buf, err
	}
	if reused {
		for i := n; i < len(w.buf); i++ {
			w.buf[i] = 0
		}
	}
	return n, w.buf, nil
}

func (w *window) writeTo(r *event.Range, dst io.Writer) (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	var from, to int64
	if r == nil {
		from, to = 0, w.length-1
	} else {
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
	}
	return io.Copy(dst, io.NewSectionReader(w.buffer, from, to-from+1))
}

func (w *window) positionToOffset(pos event.Position) (int64, error) {
	var offset int64
	switch pos := pos.(type) {
	case event.Absolute:
		offset = pos.Offset
	case event.Relative:
		offset = w.cursor + pos.Offset
	case event.End:
		offset = max(w.length, 1) - 1 + pos.Offset
	case event.VisualStart:
		if w.visualStart < 0 {
			return 0, errors.New("no visual selection found")
		}
		// TODO: save visualStart after exiting visual mode
		offset = w.visualStart + pos.Offset
	case event.VisualEnd:
		if w.visualStart < 0 {
			return 0, errors.New("no visual selection found")
		}
		offset = w.cursor + pos.Offset
	default:
		return 0, errors.New("invalid range")
	}
	return max(min(offset, max(w.length, 1)-1), 0), nil
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
		Modified:      w.changedTick != w.savedChangedTick,
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

func (w *window) updateTick() {
	w.maxChangedTick++
	w.changedTick = w.maxChangedTick
}

func (w *window) insert(offset int64, c byte) {
	w.buffer.Insert(offset, c)
	w.updateTick()
}

func (w *window) replace(offset int64, c byte) {
	w.buffer.Replace(offset, c)
	w.updateTick()
}

func (w *window) undoReplace(offset int64) {
	w.buffer.UndoReplace(offset)
	w.updateTick()
}

func (w *window) replaceIn(start, end int64, c byte) {
	w.buffer.ReplaceIn(start, end, c)
	w.updateTick()
}

func (w *window) delete(offset int64) {
	w.buffer.Delete(offset)
	w.updateTick()
}

func (w *window) undo(count int64) {
	for range max(count, 1) {
		buffer, _, offset, cursor, tick := w.history.Undo()
		if buffer == nil {
			return
		}
		w.buffer, w.offset, w.cursor, w.changedTick = buffer, offset, cursor, tick
		w.length, _ = w.buffer.Len()
	}
}

func (w *window) redo(count int64) {
	for range max(count, 1) {
		buffer, offset, cursor, tick := w.history.Redo()
		if buffer == nil {
			return
		}
		w.buffer, w.offset, w.cursor, w.changedTick = buffer, offset, cursor, tick
		w.length, _ = w.buffer.Len()
	}
}

func (w *window) cursorUp(count int64) {
	w.cursor -= min(max(count, 1), w.cursor/w.width) * w.width
	if w.append && w.extending && w.cursor < w.length-1 {
		w.append = false
		w.extending = false
		if w.length > 0 {
			w.length--
		}
	}
}

func (w *window) cursorDown(count int64) {
	w.cursor += min(
		min(
			max(count, 1),
			(max(w.length, 1)-1)/w.width-w.cursor/w.width,
		)*w.width,
		max(w.length, 1)-1-w.cursor)
}

func (w *window) cursorLeft(count int64) {
	w.cursor -= min(max(count, 1), w.cursor%w.width)
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
		w.cursor += min(
			min(max(count, 1), w.width-1-w.cursor%w.width),
			max(w.length, 1)-1-w.cursor,
		)
	} else if !w.extending {
		w.cursor += min(
			min(max(count, 1), w.width-1-w.cursor%w.width),
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
	w.cursor -= min(max(count, 1), w.cursor)
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
		w.cursor += min(max(count, 1), max(w.length, 1)-1-w.cursor)
	} else if !w.extending {
		w.cursor += min(max(count, 1), w.length-w.cursor)
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
	w.cursor = min(
		(w.cursor/w.width+max(count, 1))*w.width-1,
		max(w.length, 1)-1,
	)
}

func (w *window) cursorGoto(e event.Event) {
	if e.Range != nil {
		if e.Range.To != nil {
			w.cursorGotoPos(e.Range.To, e.CmdName)
		} else if e.Range.From != nil {
			w.cursorGotoPos(e.Range.From, e.CmdName)
		}
	}
}

func (w *window) cursorGotoPos(pos event.Position, cmdName string) {
	switch cmdName {
	case "go[to]":
		switch p := pos.(type) {
		case event.Absolute:
			pos = event.Absolute{Offset: p.Offset * w.width}
		case event.Relative:
			pos = event.Relative{Offset: p.Offset * w.width}
		case event.End:
			pos = event.End{Offset: p.Offset * w.width}
		case event.VisualStart:
			pos = event.VisualStart{Offset: p.Offset * w.width}
		case event.VisualEnd:
			pos = event.VisualEnd{Offset: p.Offset * w.width}
		}
	case "%":
		switch p := pos.(type) {
		case event.Absolute:
			pos = event.Absolute{Offset: p.Offset * w.length / 100}
		case event.Relative:
			pos = event.Relative{Offset: p.Offset * w.length / 100}
		case event.End:
			pos = event.End{Offset: p.Offset * w.length / 100}
		case event.VisualStart:
			pos = event.VisualStart{Offset: p.Offset * w.length / 100}
		case event.VisualEnd:
			pos = event.VisualEnd{Offset: p.Offset * w.length / 100}
		}
	}
	if offset, err := w.positionToOffset(pos); err == nil {
		w.cursor = offset
		if w.cursor < w.offset {
			w.offset = (max(w.cursor/w.width, w.height/2) - w.height/2) * w.width
		} else if w.cursor >= w.offset+w.height*w.width {
			h := (max(w.length, 1)+w.width-1)/w.width - w.height
			w.offset = min((w.cursor-w.height*w.width+w.width)/w.width+w.height/2, h) * w.width
		}
	}
}

func (w *window) scrollUp(count int64) {
	w.offset -= min(max(count, 1), w.offset/w.width) * w.width
	if w.cursor >= w.offset+w.height*w.width {
		w.cursor -= ((w.cursor-w.offset-w.height*w.width)/w.width + 1) * w.width
	}
}

func (w *window) scrollDown(count int64) {
	h := max((max(w.length, 1)+w.width-1)/w.width-w.height, 0)
	w.offset += min(max(count, 1), h-w.offset/w.width) * w.width
	if w.cursor < w.offset {
		w.cursor += min(
			(w.offset-w.cursor+w.width-1)/w.width*w.width,
			max(w.length, 1)-1-w.cursor,
		)
	}
}

func (w *window) scrollTop(count int64) {
	if count > 0 {
		w.cursor = min(
			min(
				count*w.width+w.cursor%w.width,
				(max(w.length, 1)-1)/w.width*w.width+w.cursor%w.width,
			),
			max(w.length, 1)-1,
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
		w.cursor = min(
			min(
				count*w.width+w.cursor%w.width,
				(max(w.length, 1)-1)/w.width*w.width+w.cursor%w.width,
			),
			max(w.length, 1)-1,
		)
	}
	w.offset = max(w.cursor/w.width-w.height/2, 0) * w.width
}

func (w *window) scrollMiddleHead(count int64) {
	w.cursorHead(0)
	w.scrollMiddle(count)
}

func (w *window) scrollBottom(count int64) {
	if count > 0 {
		w.cursor = min(
			min(
				count*w.width+w.cursor%w.width,
				(max(w.length, 1)-1)/w.width*w.width+w.cursor%w.width,
			),
			max(w.length, 1)-1,
		)
	}
	w.offset = max(w.cursor/w.width-w.height, 0) * w.width
}

func (w *window) scrollBottomHead(count int64) {
	w.cursorHead(0)
	w.scrollBottom(count)
}

func (w *window) pageUp() {
	w.offset = max(w.offset-(w.height-2)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *window) pageDown() {
	offset := max(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = min(w.offset+(w.height-2)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((max(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *window) pageUpHalf() {
	w.offset = max(w.offset-max(w.height/2, 1)*w.width, 0)
	if w.offset == 0 {
		w.cursor = 0
	} else if w.cursor >= w.offset+w.height*w.width {
		w.cursor = w.offset + (w.height-1)*w.width
	}
}

func (w *window) pageDownHalf() {
	offset := max(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.offset = min(w.offset+max(w.height/2, 1)*w.width, offset)
	if w.cursor < w.offset {
		w.cursor = w.offset
	} else if w.offset == offset {
		w.cursor = ((max(w.length, 1)+w.width-1)/w.width - 1) * w.width
	}
}

func (w *window) pageTop() {
	w.offset = 0
	w.cursor = 0
}

func (w *window) pageEnd() {
	w.offset = max(((w.length+w.width-1)/w.width-w.height)*w.width, 0)
	w.cursor = ((max(w.length, 1)+w.width-1)/w.width - 1) * w.width
}

func (w *window) windowTop(count int64) {
	w.cursor = (w.offset/w.width + min(
		min(max(count, 1)-1, (w.length-w.offset)/w.width),
		max(w.height, 1)-1,
	)) * w.width
}

func (w *window) windowMiddle() {
	h := min((w.length-w.offset)/w.width, max(w.height, 1)-1)
	w.cursor = (w.offset/w.width + h/2) * w.width
}

func (w *window) windowBottom(count int64) {
	h := min((w.length-w.offset)/w.width, max(w.height, 1)-1)
	w.cursor = (w.offset/w.width + h - min(h, max(count, 1)-1)) * w.width
}

func (w *window) jumpTo() {
	i := min(w.cursor, 16)
	_, bytes, err := w.readBytes(w.cursor-i, 32)
	if err != nil {
		return
	}
	for ; i >= 0; i-- {
		if !unicode.IsDigit(rune(bytes[i])) {
			bytes = bytes[i+1:]
			break
		}
	}
	for i := 0; i < len(bytes); i++ {
		if !unicode.IsDigit(rune(bytes[i])) {
			bytes = bytes[:i]
			break
		}
	}
	offset, _ := strconv.ParseInt(string(bytes), 10, 64)
	if offset <= 0 || w.length <= offset {
		return
	}
	w.stack = append(w.stack, position{w.cursor, w.offset})
	w.cursor = offset
	w.offset = max(offset-offset%w.width-max(w.height/3, 0)*w.width, 0)
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
	count = min(max(count, 1), w.length-w.cursor)
	b := w.buffer.Copy(w.cursor, w.cursor+count)
	w.buffer.Cut(w.cursor, w.cursor+count)
	w.length, _ = w.buffer.Len()
	w.cursor = min(w.cursor, max(w.length, 1)-1)
	w.updateTick()
	return b
}

func (w *window) deletePrevBytes(count int64) *buffer.Buffer {
	if w.cursor == 0 {
		return nil
	}
	count = min(max(count, 1), w.cursor)
	b := w.buffer.Copy(w.cursor-count, w.cursor)
	w.buffer.Cut(w.cursor-count, w.cursor)
	w.length, _ = w.buffer.Len()
	w.cursor -= count
	w.updateTick()
	return b
}

func (w *window) increment(count int64) {
	b, err := w.readByte(w.cursor)
	if err != nil && err != io.EOF {
		return
	}
	w.replace(w.cursor, b+byte(max(count, 1)))
	if w.length == 0 {
		w.length++
	}
}

func (w *window) decrement(count int64) {
	b, err := w.readByte(w.cursor)
	if err != nil && err != io.EOF {
		return
	}
	w.replace(w.cursor, b-byte(max(count, 1)))
	if w.length == 0 {
		w.length++
	}
}

func (w *window) shiftLeft(count int64) {
	b, err := w.readByte(w.cursor)
	if err != nil && err != io.EOF {
		return
	}
	w.replace(w.cursor, b<<byte(max(count, 1)))
	if w.length == 0 {
		w.length++
	}
}

func (w *window) shiftRight(count int64) {
	b, err := w.readByte(w.cursor)
	if err != nil && err != io.EOF {
		return
	}
	w.replace(w.cursor, b>>byte(max(count, 1)))
	if w.length == 0 {
		w.length++
	}
}

func (w *window) showBinary() string {
	b, err := w.readByte(w.cursor)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%08b", b)
}

func (w *window) showDecimal() string {
	b, err := w.readByte(w.cursor)
	if err != nil {
		return ""
	}
	return strconv.FormatInt(int64(b), 10)
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
			var buf [4]byte
			n := utf8.EncodeRune(buf[:], ch)
			for i := range n {
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
			if w.visualStart >= 0 && w.replaceByte {
				start, end := w.visualStart, w.cursor
				if start > end {
					start, end = end, start
				}
				w.replaceIn(start, end+1, w.pendingByte|b)
				w.visualStart = -1
				return true
			}
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
	w.cursor = min(start, max(w.length, 1)-1)
	w.updateTick()
	return b
}

func (w *window) paste(e event.Event) int64 {
	count := max(e.Count, 1)
	pos := w.cursor
	if e.Type != event.PastePrev {
		pos = min(w.cursor+1, w.length)
	}
	for range count {
		w.buffer.Paste(pos, e.Buffer)
	}
	l, _ := e.Buffer.Len()
	w.length, _ = w.buffer.Len()
	w.cursor = min(max(pos+l*count-1, 0), max(w.length, 1)-1)
	w.updateTick()
	return l * count
}

func (w *window) search(str string, forward bool) {
	if w.searchTick != w.changedTick {
		w.searcher.Abort()
		w.searcher = searcher.NewSearcher(w.buffer)
		w.searchTick = w.changedTick
	}
	ch := w.searcher.Search(w.cursor, str, forward)
	go func() {
		switch x := (<-ch).(type) {
		case error:
			w.eventCh <- event.Event{Type: event.Info, Error: x}
		case int64:
			w.mu.Lock()
			w.cursor = x
			w.mu.Unlock()
			w.redrawCh <- struct{}{}
		}
	}()
}

func (w *window) abortSearch() {
	if err := w.searcher.Abort(); err != nil {
		w.eventCh <- event.Event{Type: event.Info, Error: err}
	}
}

func (w *window) setPathName(path, name string) {
	w.path, w.name = path, name
}

func (w *window) getName() string {
	return cmp.Or(w.name, "[No Name]")
}
