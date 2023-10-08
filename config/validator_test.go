package config_test

import (
	"testing"

	"github.com/RealFax/RedQueen/config"
)

func unexpected(t *testing.T, validator config.Validator) {
	if err := validator.Valid(); err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func expected(t *testing.T, validator config.Validator) {
	if err := validator.Valid(); err != nil {
		t.Log("expected error:", err.Error())
	}
}

func TestEnumStoreBackend_Valid(t *testing.T) {
	unexpected(t, config.StoreBackendNuts)

	expected(t, config.EnumStoreBackend("db"))
}

func TestEnumNutsRWMode_Valid(t *testing.T) {
	for _, val := range []config.EnumNutsRWMode{
		config.NutsRWModeFileIO,
		config.NutsRWModeMMap,
	} {
		unexpected(t, val)
	}

	expected(t, config.EnumNutsRWMode("rwmode"))
}

func TestEnumLogLogger_Valid(t *testing.T) {
	for _, val := range []config.EnumLogLogger{
		config.LogLoggerZap,
		config.LogLoggerInternal,
	} {
		unexpected(t, val)
	}

	expected(t, config.EnumLogLogger("none"))
}

func TestFilePath_Valid(t *testing.T) {
	unexpected(t, config.FilePath("./"))
}
