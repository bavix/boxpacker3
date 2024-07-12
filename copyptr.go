package boxpacker3

// CopyPtr creates a copy of a pointer.
func CopyPtr[T any](original *T) *T {
	if original == nil {
		return nil
	}

	copyOfValue := *original

	return &copyOfValue
}

// CopySlicePtr creates a copy of a slice of pointers.
func CopySlicePtr[T any](data []*T) []*T {
	result := make([]*T, len(data))

	for i, item := range data {
		result[i] = CopyPtr(item)
	}

	return result
}
