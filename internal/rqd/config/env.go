package config

import (
	"flag"
	"strings"
	"syscall"
)

func BindEnvVar(value flag.Value, name string) {
	val, found := syscall.Getenv(strings.ToUpper(name))
	if !found {
		return
	}
	if err := value.Set(val); err != nil {
		panic("bind env error: " + err.Error())
	}
}

func EnvStringVar(p *string, name string, value string) {
	BindEnvVar(newStringValue(value, p), name)
}

func EnvInt64Var(p *int64, name string, value int64) {
	BindEnvVar(newInt64Value(value, p), name)
}

func EnvBoolVar(p *bool, name string, value bool) {
	BindEnvVar(newBoolValue(value, p), name)
}
