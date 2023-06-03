package store

const (
	// DefaultNamespace = "RedQueen"
	DefaultNamespace = ""
)

func UnwrapGet(val *Value, err error) ([]byte, error) {
	return val.Data, err
}
