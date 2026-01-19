package boxpacker3

type cloner interface {
	clone() cloner
}

func clonePtr[T any](original *T) *T {
	if original == nil {
		return nil
	}

	if c, ok := any(original).(cloner); ok {
		if cloned, ok := any(c.clone()).(*T); ok {
			return cloned
		}
	}

	copyOfValue := *original

	return &copyOfValue
}

func cloneSlice[T any](data []*T) []*T {
	if data == nil {
		return nil
	}

	result := make([]*T, len(data))

	for i, item := range data {
		result[i] = clonePtr(item)
	}

	return result
}
