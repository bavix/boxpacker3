package boxpacker3

// CopyPtr creates a copy of a pointer.
//
// It takes a pointer to a generic type T as an argument and returns a new pointer
// to a copy of the original value. If the original pointer is nil, it returns nil.
func CopyPtr[T any](original *T) *T {
	// If the original pointer is nil, return nil.
	if original == nil {
		return nil
	}

	// Create a copy of the value pointed to by the original pointer.
	copyOfValue := *original

	// Return a new pointer to the copied value.
	return &copyOfValue
}

// CopySlicePtr creates a copy of a slice of pointers.
//
// It takes a slice of pointers as an argument and returns a new slice with the same
// elements, but with each element being a copy of the original.
func CopySlicePtr[T any](data []*T) []*T {
	// Create a new slice with the same length as the original.
	result := make([]*T, len(data))

	// Iterate over the original slice and copy each element to the new slice.
	for i, item := range data {
		result[i] = CopyPtr(item)
	}

	// Return the new slice.
	return result
}
