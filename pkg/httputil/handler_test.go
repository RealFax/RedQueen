package httputil_test

import (
	"github.com/RealFax/RedQueen/pkg/httputil"
	"net/http"
	"testing"
)

func TestNewBasicAuth(t *testing.T) {
	users := map[string]string{
		"root":  "toor",
		"admin": "P@ssw0rd",
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Authed"))
	})

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: httputil.NewBasicAuth(mux, httputil.NewMemoryBasicAuthFunc(users)),
	}

	srv.ListenAndServe()
}
