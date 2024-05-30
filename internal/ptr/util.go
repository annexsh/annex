package ptr

// Get returns a pointer to a value.
func Get[T any](val T) *T {
	return &val
}

// Copy copies a pointer's value into a new pointer of the same type.
func Copy[T any](t *T) *T {
	var ret T
	ret = *t
	return &ret
}
