package boxpacker3

func copyPtr[T any](original *T) *T {
	copyOfValue := *original

	return &copyOfValue
}

func copySlicePtr[T any](data []*T) []*T {
	result := make([]*T, len(data))

	for i := range data {
		val := *data[i]
		result[i] = &val
	}

	return result
}
