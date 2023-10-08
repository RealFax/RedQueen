package store

const (
	// DefaultNamespace = "RedQueen"
	DefaultNamespace = ""
)

func UnwrapGet(val *Value, err error) ([]byte, error) {
	if val == nil {
		return nil, err
	}
	return val.Data, err
}
