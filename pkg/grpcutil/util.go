package grpcutil

import (
	"bytes"
	"encoding/base64"
	"github.com/RealFax/RedQueen/pkg/hack"
	"strings"
)

func ParseAuthorization(auth string, fc BasicAuthFunc) bool {
	p, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return false
	}

	xp := strings.Split(hack.Bytes2String(p), ":")
	if len(xp) != 2 {
		return false
	}

	return fc(xp[0], xp[1])
}

func BuildAuthorization(username, password string) string {
	b := bytes.Buffer{}
	b.Grow(len(username) + len(password) + 1)

	b.WriteString(username)
	b.WriteRune(':')
	b.WriteString(password)

	return base64.StdEncoding.EncodeToString(b.Bytes())
}
