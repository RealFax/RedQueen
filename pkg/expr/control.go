package expr

func IsZero[T comparable](v T) bool {
	return v == Zero[T]()
}

func If[T any](f bool, then, end T) T {
	if f {
		return then
	}
	return end
}
