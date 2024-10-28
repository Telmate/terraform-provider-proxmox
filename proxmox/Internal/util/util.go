package util

// Gets inlined by the compiler, so it's not a performance hit
func Pointer[T any](v T) *T {
	return &v
}
