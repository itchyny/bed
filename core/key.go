package core

import (
	"strconv"
)

// Key represents one keyboard stroke.
type Key string

type keyEvent struct {
	keys  []Key
	event EventType
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
		if k != ks[i] {
			return keysNeq
		}
	}
	return keysEq
}

// KeyManager holds the key mappings and current key sequence.
type KeyManager struct {
	keys      []Key
	keyEvents []keyEvent
	count     bool
}

// NewKeyManager creates a new KeyManager.
func NewKeyManager(count bool) *KeyManager {
	return &KeyManager{count: count}
}

// Register adds a new key mapping.
func (km *KeyManager) Register(event EventType, keys ...Key) {
	km.keyEvents = append(km.keyEvents, keyEvent{keys, event})
}

// Press checks the new key down event.
func (km *KeyManager) Press(k Key) Event {
	km.keys = append(km.keys, k)
	for i := 0; i < len(km.keys); i++ {
		keys := km.keys[i:]
		var count int64
		if km.count {
			numStr := ""
			for j, k := range keys {
				if len(k) == 1 && ('1' <= k[0] && k[0] <= '9' || k[0] == '0' && j > 0) {
					numStr += string(k)
				} else {
					break
				}
			}
			keys = keys[len(numStr):]
			count, _ = strconv.ParseInt(numStr, 10, 64)
		}
		for _, ke := range km.keyEvents {
			switch ke.cmp(keys) {
			case keysPending:
				return Event{Type: EventNop}
			case keysEq:
				km.keys = nil
				return Event{Type: ke.event, Count: count}
			}
		}
	}
	km.keys = nil
	return Event{Type: EventNop}
}
