package cache

import (
	"testing"
	"time"
)

func TestTTLCacheBasic(t *testing.T) {
	c := NewTTL(2)
	c.Set("a", []byte("x"), 10*time.Millisecond)
	if b, ok := c.Get("a"); !ok || string(b) != "x" {
		t.Fatalf("expected hit x")
	}
	time.Sleep(12 * time.Millisecond)
	if _, ok := c.Get("a"); ok {
		t.Fatalf("expected expiry")
	}
	c.Set("b", []byte("1"), 0)
	c.Set("c", []byte("2"), 0)
	c.Set("d", []byte("3"), 0) // triggers eviction (cap=2)
	st := c.Stats()
	if st.Size > 2 {
		t.Fatalf("cap exceeded")
	}
	if st.Evicted == 0 {
		t.Fatalf("expected eviction")
	}
}
