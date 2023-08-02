package boxpacker3

func copyPtr[T any](original *T) *T {
	copyOfValue := *original

	return &copyOfValue
}

func copySlicePtr[T any](data []*T) []*T {
	result := make([]*T, len(data))
	for i := range data {
		result[i] = copyPtr(data[i])
	}

	return result
}
