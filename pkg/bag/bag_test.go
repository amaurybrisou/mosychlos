package bag

import (
	"testing"
)

func TestBagBasic(t *testing.T) {
	b := New()
	if b.Len() != 0 {
		t.Fatalf("expected empty")
	}
	b2 := b.Set(KMacro, 5)
	if b.Len() != 0 {
		t.Fatalf("immutability broken")
	}
	vAny, ok := b2.Get(KMacro)
	if !ok || vAny.(int) != 5 {
		t.Fatalf("missing value")
	}
	var out int
	if !b2.GetAs(KMacro, &out) || out != 5 {
		t.Fatalf("GetAs failed")
	}
}
