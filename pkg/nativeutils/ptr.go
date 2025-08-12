package nativeutils

// Ptr returns a pointer to the given value.
// This is a generic helper function to create pointers for any type.
func Ptr[T any](v T) *T {
	return &v
}
