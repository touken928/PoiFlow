package akstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	entries := []Entry{{Name: "开发", Key: "ak_dev"}, {Name: "备用", Key: "ak_backup"}}

	if err := Save(path, entries); err != nil { t.Fatalf("Save failed: %v", err) }

	loaded, err := Load(path)
	if err != nil { t.Fatalf("Load failed: %v", err) }
	if len(loaded) != 2 { t.Fatalf("expected 2, got %d", len(loaded)) }
	if loaded[0].Name != "开发" || loaded[0].Key != "ak_dev" { t.Errorf("unexpected: %+v", loaded[0]) }
	if loaded[1].Name != "备用" || loaded[1].Key != "ak_backup" { t.Errorf("unexpected: %+v", loaded[1]) }
}

func TestLoadNonExistent(t *testing.T) {
	e, err := Load("/nonexistent/config.yaml")
	if err != nil { t.Fatal(err) }
	if e != nil { t.Errorf("expected nil, got %v", e) }
}

func TestSaveEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.yaml")
	if err := Save(path, nil); err != nil { t.Fatal(err) }
	loaded, err := Load(path)
	if err != nil { t.Fatal(err) }
	if len(loaded) != 0 { t.Errorf("expected 0, got %d", len(loaded)) }
}

func TestYAMLFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.yaml")
	_ = Save(path, []Entry{{Name: "主", Key: "ak1"}, {Name: "次", Key: "ak2"}})
	data, _ := os.ReadFile(path)
	if string(data) != "aks:\n    - name: 主\n      key: ak1\n    - name: 次\n      key: ak2\n" {
		t.Errorf("unexpected format:\n%s", string(data))
	}
}
