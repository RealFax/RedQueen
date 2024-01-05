package tlsutil_test

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"github.com/RealFax/RedQueen/pkg/tlsutil"
	"net/http"
	"testing"
)

func TestGenX509KeyPair(t *testing.T) {
	cert, err := tlsutil.GenX509KeyPair(pkix.Name{
		CommonName:         "localhost",
		Country:            []string{"Russian"},
		Organization:       []string{"RealFax"},
		OrganizationalUnit: []string{"OpenSource"},
	})
	if err != nil {
		t.Fatal(err)
	}

	m := &http.ServeMux{}
	m.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Hello, TLS"))
	})

	cfg := &tls.Config{
		NextProtos:         []string{"http/1.1"},
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	srv := &http.Server{
		Addr:      "localhost:8080",
		Handler:   m,
		TLSConfig: cfg,
	}

	listener, err := tls.Listen("tcp", "localhost:8080", cfg)
	if err != nil {
		t.Fatal(err)
	}

	srv.Serve(listener)
}
