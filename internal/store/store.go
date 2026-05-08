package store

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Entry struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type ExportField string

const (
	FieldName      ExportField = "name"
	FieldAddress   ExportField = "address"
	FieldTelephone ExportField = "telephone"
	FieldProvince  ExportField = "province"
	FieldCity      ExportField = "city"
	FieldArea      ExportField = "area"
	FieldUID       ExportField = "uid"
	FieldQuery     ExportField = "query"
	FieldType      ExportField = "type"
	FieldTaskName  ExportField = "taskName"
	FieldTarget    ExportField = "target"
)

var AllExportFields = []ExportField{
	FieldName, FieldAddress, FieldTelephone,
	FieldProvince, FieldCity, FieldArea,
	FieldUID, FieldQuery, FieldType, FieldTaskName, FieldTarget,
}

var MandatoryFields = []ExportField{}

type ExportConfig struct {
	Fields []ExportField `yaml:"fields"`
}

type Config struct {
	AKs    []Entry      `yaml:"aks"`
	Export ExportConfig `yaml:"export,omitempty"`
}

func DefaultExportConfig() ExportConfig {
	return ExportConfig{Fields: AllExportFields}
}

func readConfig(path string) (*Config, error) {
	abs, err := filepath.Abs(path)
	if err != nil { return nil, err }
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := &Config{}
			cfg.Export = DefaultExportConfig()
			return cfg, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil { return nil, err }
	if cfg.Export.Fields == nil {
		cfg.Export = DefaultExportConfig()
	}
	return &cfg, nil
}

func writeConfig(path string, cfg *Config) error {
	abs, err := filepath.Abs(path)
	if err != nil { return err }
	dir := filepath.Dir(abs)
	if err := os.MkdirAll(dir, 0755); err != nil { return err }
	data, err := yaml.Marshal(cfg)
	if err != nil { return err }
	return os.WriteFile(abs, data, 0644)
}

func LoadAKs(path string) ([]Entry, error) {
	cfg, err := readConfig(path)
	if err != nil { return nil, err }
	return cfg.AKs, nil
}

func SaveAKs(path string, entries []Entry) error {
	cfg, err := readConfig(path)
	if err != nil { return err }
	cfg.AKs = entries
	return writeConfig(path, cfg)
}

func LoadExportConfig(path string) (ExportConfig, error) {
	cfg, err := readConfig(path)
	if err != nil { return ExportConfig{}, err }
	return cfg.Export, nil
}

func SaveExportConfig(path string, ec ExportConfig) error {
	cfg, err := readConfig(path)
	if err != nil { return err }
	cfg.Export = ec
	return writeConfig(path, cfg)
}
