package core

import (
	"strings"
)

type Key struct {
	Key   string
	Ctrl  bool
	Shift bool
}

func NewKey(k string) Key {
	var ctrl, shift bool
	if strings.HasPrefix(k, "c-") {
		ctrl = true
		k = k[2:]
	}
	if strings.HasPrefix(k, "s-") {
		shift = true
		k = k[2:]
	}
	l := strings.ToLower(k)
	return Key{Key: l, Ctrl: ctrl, Shift: shift || k != l}
}

func (x Key) eq(y Key) bool {
	return x.Key == y.Key && x.Ctrl == y.Ctrl && x.Shift == y.Shift
}

type keyEvent struct {
	keys  []Key
	event Event
}

const (
	keysEq = iota
	keysPending
	keysNeq
)

func (ke keyEvent) cmp(ks []Key) int {
	if len(ke.keys) < len(ks) {
		return keysNeq
	}
	for i, k := range ke.keys {
		if i >= len(ks) {
			return keysPending
		}
		if !k.eq(ks[i]) {
			return keysNeq
		}
	}
	return keysEq
}

type KeyManager struct {
	keys      []Key
	keyEvents []keyEvent
}

func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

func (km *KeyManager) Register(event Event, keys ...Key) {
	km.keyEvents = append(km.keyEvents, keyEvent{keys, event})
}

func (km *KeyManager) Press(k Key) Event {
	km.keys = append(km.keys, k)
	for i := 0; i < len(km.keys); i++ {
		keys := km.keys[i:]
		for _, ke := range km.keyEvents {
			switch ke.cmp(keys) {
			case keysPending:
				return Event(Nop)
			case keysEq:
				km.keys = nil
				return ke.event
			}
		}
	}
	km.keys = nil
	return Event(Nop)
}
