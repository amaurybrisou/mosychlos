package nativeutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		val := "test"
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
		assert.Equal(t, &val, ptr) // Should point to different memory
	})

	t.Run("int", func(t *testing.T) {
		val := 42
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("float64", func(t *testing.T) {
		val := 3.14
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("bool", func(t *testing.T) {
		val := true
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
	})

	t.Run("struct", func(t *testing.T) {
		type TestStruct struct {
			Field string
		}
		val := TestStruct{Field: "test"}
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, val, *ptr)
		assert.Equal(t, "test", ptr.Field)
	})

	t.Run("zero values", func(t *testing.T) {
		assert.Equal(t, 0, *Ptr(0))
		assert.Equal(t, "", *Ptr(""))
		assert.Equal(t, false, *Ptr(false))
		assert.Equal(t, 0.0, *Ptr(0.0))
	})
}
