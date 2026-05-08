package akstore

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type config struct {
	AKs []Entry `yaml:"aks"`
}

func Load(path string) ([]Entry, error) {
	abs, err := filepath.Abs(path)
	if err != nil { return nil, err }
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) { return nil, nil }
		return nil, err
	}
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil { return nil, err }
	return cfg.AKs, nil
}

func Save(path string, entries []Entry) error {
	abs, err := filepath.Abs(path)
	if err != nil { return err }
	dir := filepath.Dir(abs)
	if err := os.MkdirAll(dir, 0755); err != nil { return err }
	cfg := config{AKs: entries}
	data, err := yaml.Marshal(&cfg)
	if err != nil { return err }
	return os.WriteFile(abs, data, 0644)
}
