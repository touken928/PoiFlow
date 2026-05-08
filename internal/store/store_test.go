package store

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadAKs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	entries := []Entry{{Name: "开发", Key: "ak_dev"}, {Name: "备用", Key: "ak_backup"}}

	if err := SaveAKs(path, entries); err != nil { t.Fatalf("SaveAKs: %v", err) }

	loaded, err := LoadAKs(path)
	if err != nil { t.Fatal(err) }
	if len(loaded) != 2 { t.Fatalf("expected 2, got %d", len(loaded)) }
	if loaded[0].Name != "开发" || loaded[0].Key != "ak_dev" { t.Errorf("unexpected: %+v", loaded[0]) }
}

func TestExportConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.yaml")

	ec := ExportConfig{Fields: []ExportField{FieldName, FieldAddress}}
	if err := SaveExportConfig(path, ec); err != nil { t.Fatal(err) }

	loaded, err := LoadExportConfig(path)
	if err != nil { t.Fatal(err) }
	if len(loaded.Fields) != 2 { t.Errorf("expected 2 fields, got %d", len(loaded.Fields)) }
}

func TestDefaultExportConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.yaml")
	// no export config saved → should get defaults
	ec, err := LoadExportConfig(path)
	if err != nil { t.Fatal(err) }
	if len(ec.Fields) != len(AllExportFields) {
		t.Errorf("expected %d defaults, got %d", len(AllExportFields), len(ec.Fields))
	}
}

func TestAKsAndExportTogether(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.yaml")
	_ = SaveAKs(path, []Entry{{Name: "x", Key: "ak1"}})
	_ = SaveExportConfig(path, ExportConfig{Fields: []ExportField{FieldName}})

	aks, _ := LoadAKs(path)
	if len(aks) != 1 { t.Error("expected 1 ak") }

	ec, _ := LoadExportConfig(path)
	if len(ec.Fields) != 1 { t.Error("expected 1 export field") }
}

func TestLoadNonExistent(t *testing.T) {
	e, _ := LoadAKs("/nonexistent/c.yaml")
	if e != nil { t.Errorf("expected nil, got %v", e) }
}
