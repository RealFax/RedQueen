package httputil

import (
	"crypto/rand"
	"encoding/base64"
)

func TraceID() string {
	if !traceID.Load() {
		return ""
	}
	tid := make([]byte, 16)
	_, _ = rand.Read(tid)
	return base64.StdEncoding.EncodeToString(tid)
}
