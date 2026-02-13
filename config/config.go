package config

import (
	"bytes"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	data []byte
}

func FromFile(name string) (Config, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return Config{}, err
	}

	return FromBytes(b)
}

func FromBytes(b []byte) (Config, error) {
	var a yaml.MapSlice
	err := yaml.Unmarshal(b, &a) // validate yaml

	return Config{data: b}, err
}

// Unmarshal maps config into a type
func (cfg Config) Unmarshal(target any) error {
	return yaml.Unmarshal(cfg.data, target)
}

// UnmarshalPath reads config at specific path and maps it into a type
// path example: $.store.book[*].author
func (cfg Config) UnmarshalPath(path string, target any) error {
	ps, err := yaml.PathString(path)
	if err != nil {
		return err
	}
	return ps.Read(bytes.NewReader(cfg.data), target)
}
