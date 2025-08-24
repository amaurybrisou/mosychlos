package bag

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

//go:generate mockgen -source=shared_bag.go -destination=mocks/shared_bag_mock.go -package=mocks

// SharedBag interface for mutable shared state
type SharedBag interface {
	Get(k Key) (any, bool)
	MustGet(k Key) any
	GetAs(k Key, out any) bool
	Set(k Key, v any)
	Update(k Key, fn func(any) any)
	Has(k Key) bool
	Snapshot() Bag
	MarshalJSON() ([]byte, error)
}

// sharedBag implementation with thread-safe operations
type sharedBag struct {
	data map[Key]any
	mu   sync.RWMutex
}

// NewSharedBag creates a new thread-safe shared bag
func NewSharedBag() SharedBag {
	return &sharedBag{
		data: make(map[Key]any),
	}
}

// NewSharedBagFrom creates a shared bag from an existing immutable bag
func NewSharedBagFrom(b Bag) SharedBag {
	sb := &sharedBag{
		data: make(map[Key]any),
	}

	// Copy data from immutable bag
	for _, key := range b.Keys() {
		if val, ok := b.Get(key); ok {
			sb.data[key] = val
		}
	}

	return sb
}

func (sb *sharedBag) Get(k Key) (any, bool) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	v, ok := sb.data[k]
	return v, ok
}

func (sb *sharedBag) MustGet(k Key) any {
	v, ok := sb.Get(k)
	if !ok {
		panic(fmt.Errorf("key not found: %v", k))
	}
	return v
}

func (sb *sharedBag) GetAs(k Key, out any) bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	v, ok := sb.data[k]
	if !ok {
		return false
	}

	// Use same type assertion logic as original bag
	switch dst := out.(type) {
	case *string:
		vv, ok := v.(string)
		if !ok {
			return false
		}
		*dst = vv
	case *int:
		vv, ok := v.(int)
		if !ok {
			return false
		}
		*dst = vv
	case *float64:
		vv, ok := v.(float64)
		if !ok {
			return false
		}
		*dst = vv
	case *[]string:
		vv, ok := v.([]string)
		if !ok {
			return false
		}
		*dst = append((*dst)[:0], vv...)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return false
		}
		if err := json.Unmarshal(data, &dst); err != nil {
			return false
		}
	}
	return true
}

func (sb *sharedBag) Set(k Key, v any) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.data[k] = v
}

func (sb *sharedBag) Update(k Key, fn func(any) any) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	current, exists := sb.data[k]
	if !exists {
		sb.data[k] = make(map[string]any)
		current = sb.data[k]
	}

	newValue := fn(current)

	slog.Debug("Updating bag key", "key", k, "old", current, "new", newValue)

	sb.data[k] = newValue
}

func (sb *sharedBag) Has(k Key) bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	_, ok := sb.data[k]
	return ok
}

func (sb *sharedBag) Snapshot() Bag {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	// Create immutable snapshot
	return From(sb.data)
}

func (sb *sharedBag) MarshalJSON() ([]byte, error) {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return json.Marshal(sb.data)
}

// LoadSharedBagFromJSON loads a SharedBag from a JSON file (io.Reader)
func LoadSharedBagFromJSON(r io.Reader) (SharedBag, error) {
	var m map[Key]any
	dec := json.NewDecoder(r)
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	sb := &sharedBag{
		data: m,
	}
	return sb, nil
}
