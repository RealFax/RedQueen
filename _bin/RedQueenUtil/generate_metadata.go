package main

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

type generateMetadata struct {
	cfg GenerateMetadataConfig
}

func (m *generateMetadata) configEntity() config {
	return &m.cfg
}

func (m *generateMetadata) exec() error {
	if m.cfg.Endpoints == "" {
		return errors.New("empty endpoints")
	}

	f, err := os.OpenFile(metadataFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(m.cfg)
}

func newGenerateMetadata() *generateMetadata {
	return &generateMetadata{}
}
