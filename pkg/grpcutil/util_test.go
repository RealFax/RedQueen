package grpcutil_test

import (
	"github.com/RealFax/RedQueen/pkg/grpcutil"
	"testing"
)

func TestParseAuthorization(t *testing.T) {
	testAuthFunc := func(username, password string) bool {
		return true
	}

	tests := []struct {
		name         string
		auth         string
		expectedUser string
		expectedPass string
		want         bool
	}{
		{"ValidAuthorization", "dXNlcm5hbWU6cGFzc3dvcmQ=", "username", "password", true},
		{"InvalidBase64", "invalid_base64", "", "", false},
		{"IncompleteCredentials", "dXNlcm5hbWU=", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpcutil.ParseAuthorization(tt.auth, testAuthFunc)
			if got != tt.want {
				t.Errorf("ParseAuthorization() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试BuildAuthorization函数
func TestBuildAuthorization(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		want     string
	}{
		{"ValidCredentials", "username", "password", "dXNlcm5hbWU6cGFzc3dvcmQ="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpcutil.BuildAuthorization(tt.username, tt.password)
			if got != tt.want {
				t.Errorf("BuildAuthorization() = %v, want %v", got, tt.want)
			}
		})
	}
}
