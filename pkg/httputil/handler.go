package httputil

import (
	"crypto/subtle"
	"net/http"
)

func WrapE(fc func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fc(w, r)

		if err != nil {
			if e, ok := As(err); ok {
				c := Any(e.StatusCode, e.Code)
				if e.Message != "" {
					c.Message(e.Message)
				}
				c.Ok(w)
				return
			}
			Any(http.StatusInternalServerError, 0).Message(err.Error()).Ok(w)
			return
		}

		return
	}
}

type BasicAuthFunc func(username, password string) bool
type basicAuth struct {
	next   http.Handler
	authFC BasicAuthFunc
}

func (a *basicAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, found := r.BasicAuth()
	if !found {
		w.Header().Add("WWW-Authenticate", `Basic realm="auth failed"`)
		Any(http.StatusUnauthorized, 401).Message("No BasicAuth present").Ok(w)
		return
	}

	if !a.authFC(username, password) {
		w.Header().Add("WWW-Authenticate", `Basic realm="auth failed"`)
		Any(http.StatusUnauthorized, 401).Message("Unauthorized").Ok(w)
		return
	}

	a.next.ServeHTTP(w, r)
}

func NewBasicAuth(next http.Handler, fc BasicAuthFunc) http.Handler {
	return &basicAuth{next: next, authFC: fc}
}

func NewMemoryBasicAuthFunc(users map[string]string) BasicAuthFunc {
	return func(username, password string) bool {
		return subtle.ConstantTimeCompare([]byte(users[username]), []byte(password)) == 1
	}
}
