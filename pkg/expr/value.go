package expr

func Pointer[T any](v T) *T {
	return &v
}

func Zero[T any]() (zero T) {
	return
}

func MustPointer[T comparable](v T) *T {
	if IsZero(v) {
		return nil
	}
	return &v
}
