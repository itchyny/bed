package core

import "testing"

func TestKeyManagerPress(t *testing.T) {
	km := NewKeyManager(true)
	km.Register(EventCursorUp, "k")
	e := km.Press("k")
	if e.Type != EventCursorUp {
		t.Errorf("pressing k should emit EventCursorUp but got: %d", e.Type)
	}
	e = km.Press("j")
	if e.Type != EventNop {
		t.Errorf("pressing j should be nop but got: %d", e.Type)
	}
}

func TestKeyManagerPressMulti(t *testing.T) {
	km := NewKeyManager(true)
	km.Register(EventCursorUp, "k", "k", "j")
	km.Register(EventCursorDown, "k", "j", "j")
	km.Register(EventCursorDown, "j", "k", "k")
	e := km.Press("k")
	if e.Type != EventNop {
		t.Errorf("pressing k should be nop but got: %d", e.Type)
	}
	e = km.Press("k")
	if e.Type != EventNop {
		t.Errorf("pressing k twice should be nop but got: %d", e.Type)
	}
	e = km.Press("k")
	if e.Type != EventNop {
		t.Errorf("pressing k three times should be nop but got: %d", e.Type)
	}
	e = km.Press("j")
	if e.Type != EventCursorUp {
		t.Errorf("pressing kkj should emit EventCursorUp but got: %d", e.Type)
	}
}

func TestKeyManagerPressCount(t *testing.T) {
	km := NewKeyManager(true)
	km.Register(EventCursorUp, "k", "j")
	e := km.Press("k")
	if e.Type != EventNop {
		t.Errorf("pressing k should be nop but got: %d", e.Type)
	}
	e = km.Press("3")
	if e.Type != EventNop {
		t.Errorf("pressing 3 should be nop but got: %d", e.Type)
	}
	e = km.Press("7")
	if e.Type != EventNop {
		t.Errorf("pressing 7 should be nop but got: %d", e.Type)
	}
	e = km.Press("k")
	if e.Type != EventNop {
		t.Errorf("pressing k should be nop but got: %d", e.Type)
	}
	e = km.Press("j")
	if e.Type != EventCursorUp {
		t.Errorf("pressing 37kj should emit EventCursorUp but got: %d", e.Type)
	}
	if e.Count != 37 {
		t.Errorf("pressing 37kj should emit EventCursorUp with count 37 but got: %d", e.Count)
	}
}
