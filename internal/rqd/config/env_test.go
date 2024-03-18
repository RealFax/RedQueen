package config_test

import (
	"github.com/RealFax/RedQueen/internal/rqd/config"
	"github.com/stretchr/testify/assert"
	"syscall"
	"testing"
)

func init() {
	syscall.Setenv("BONJOUR", "HELLO")
	syscall.Setenv("TEST_STRING", "IS_STRING")
	syscall.Setenv("TEST_INT64", "2147483648")
	syscall.Setenv("TEST_BOOL", "true")
}

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) String() string { return string(*s) }

func TestBindEnvVar(t *testing.T) {
	var s string
	config.BindEnvVar(newStringValue("default", &s), "bonjour")
	assert.Equal(t, "HELLO", s)
}

func TestEnvStringVar(t *testing.T) {
	var s string
	config.EnvStringVar(&s, "test_string", "none")
	assert.Equal(t, s, "IS_STRING")
}

func TestEnvInt64Var(t *testing.T) {
	var i int64
	config.EnvInt64Var(&i, "test_int64", -1)
	assert.Equal(t, int64(2147483648), i)
}

func TestEnvBoolVar(t *testing.T) {
	var b bool
	config.EnvBoolVar(&b, "test_bool", false)
	assert.Equal(t, true, b)
}
