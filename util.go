package schema

func optional[T any](defaultValue T, values ...T) T {
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func optionalPtr[T any](defaultValue T, values ...T) *T {
	if len(values) > 0 {
		return &values[0]
	}
	return &defaultValue
}

func optionalNil[T any](values ...T) *T {
	if len(values) > 0 {
		return &values[0]
	}
	return nil
}

func ptrOf[T any](value T) *T {
	return &value
}

func ternary[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}
