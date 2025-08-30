package util

func Optional[T any](defaultValue T, values ...T) T {
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func OptionalPtr[T any](defaultValue T, values ...T) *T {
	if len(values) > 0 {
		return &values[0]
	}
	return &defaultValue
}

func OptionalNil[T any](values ...T) *T {
	if len(values) > 0 {
		return &values[0]
	}
	return nil
}

func PtrOf[T any](value T) *T {
	return &value
}

func Ternary[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}
