package bag

import (
	"testing"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
)

func TestBagBasic(t *testing.T) {
	b := New()
	if b.Len() != 0 {
		t.Fatalf("expected empty")
	}
	b2 := b.Set(keys.KMacro, 5)
	if b.Len() != 0 {
		t.Fatalf("immutability broken")
	}
	vAny, ok := b2.Get(keys.KMacro)
	if !ok || vAny.(int) != 5 {
		t.Fatalf("missing value")
	}
	var out int
	if !b2.GetAs(keys.KMacro, &out) || out != 5 {
		t.Fatalf("GetAs failed")
	}
}
