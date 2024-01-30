package httputil

import (
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
)

func BindJSON(v any, r io.Reader) error {
	err := jsoniter.ConfigFastest.NewDecoder(r).Decode(v)
	if err != nil {
		return NewStatus(http.StatusBadRequest, 0, "Bad Request")
	}
	return nil
}

func XBindJSON[T any](r io.Reader) (T, error) {
	var dst T
	return dst, BindJSON(&dst, r)
}
