package key

import (
	"strconv"

	"github.com/itchyny/bed/event"
)

// Key represents one keyboard stroke.
type Key string

type keyEvent struct {
	keys  []Key
	event event.Type
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

// Manager holds the key mappings and current key sequence.
type Manager struct {
	keys   []Key
	events []keyEvent
	count  bool
}

// NewManager creates a new Manager.
func NewManager(count bool) *Manager {
	return &Manager{count: count}
}

// Register adds a new key mapping.
func (km *Manager) Register(eventType event.Type, keys ...Key) {
	km.events = append(km.events, keyEvent{keys, eventType})
}

// Press checks the new key down event.
func (km *Manager) Press(k Key) event.Event {
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
		for _, ke := range km.events {
			switch ke.cmp(keys) {
			case keysPending:
				return event.Event{Type: event.Nop}
			case keysEq:
				km.keys = nil
				return event.Event{Type: ke.event, Count: count}
			}
		}
	}
	km.keys = nil
	return event.Event{Type: event.Nop}
}
