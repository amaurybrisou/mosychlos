// Package bag provides an key/value store with typed accessors.
package bag

import (
	"encoding/json"
	"maps"
)

//go:generate mockgen -source=bag.go -destination=mocks/bag_mock.go -package=mocks

// Bag is an immutable key/value store with typed accessors.
type Bag interface {
	Get(k Key) (any, bool)
	GetAs(k Key, out any) bool // out must be pointer; returns false on type mismatch or missing
	Set(k Key, v any) Bag
	Has(k Key) bool
	Keys() []Key
	Len() int
}

type bag struct{ m map[Key]any }

// New creates an empty Bag.
func New() Bag { return bag{m: make(map[Key]any)} }

// From creates a Bag from an existing map (shallow copied).
func From(src map[Key]any) Bag {
	cp := make(map[Key]any, len(src))
	maps.Copy(cp, src)
	return bag{m: cp}
}

func (b bag) clone() map[Key]any {
	cp := make(map[Key]any, len(b.m))
	maps.Copy(cp, b.m)
	return cp
}

func (b bag) Get(k Key) (any, bool) {
	v, ok := b.m[k]
	return v, ok
}

func (b bag) GetAs(k Key, out any) bool {
	v, ok := b.Get(k)
	if !ok {
		return false
	}
	// expect pointer
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
		// try direct assign via type assertion generically
		// fallback: ignore for now (could use reflection later)
		return false
	}
	return true
}

func (b bag) Set(k Key, v any) Bag {
	cp := b.clone()
	cp[k] = v
	return bag{m: cp}
}

func (b bag) Has(k Key) bool {
	_, ok := b.Get(k)
	return ok
}

func (b bag) Keys() []Key {
	ks := make([]Key, 0, len(b.m))
	for k := range b.m {
		ks = append(ks, k)
	}
	return ks
}

func (b bag) Len() int { return len(b.m) }

func (b bag) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.m)
}
