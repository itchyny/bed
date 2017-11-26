package core

// Key represents one keyboard stroke.
type Key string

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
}

// NewKeyManager creates a new KeyManager.
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// Register adds a new key mapping.
func (km *KeyManager) Register(event Event, keys ...Key) {
	km.keyEvents = append(km.keyEvents, keyEvent{keys, event})
}

// Press checks the new key down event.
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
