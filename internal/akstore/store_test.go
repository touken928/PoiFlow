package akstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "aks.yaml")
	aks := []string{"ak_test_1", "ak_test_2"}

	if err := Save(path, aks); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil { t.Fatalf("Load failed: %v", err) }

	if len(loaded) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded))
	}
	if loaded[0] != "ak_test_1" || loaded[1] != "ak_test_2" {
		t.Errorf("unexpected entries: %v", loaded)
	}
}

func TestLoadNonExistent(t *testing.T) {
	entries, err := Load("/nonexistent/path/aks.yaml")
	if err != nil { t.Fatalf("Load for non-existent file should not error: %v", err) }
	if entries != nil { t.Errorf("expected nil for non-existent, got %v", entries) }
}

func TestSaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "aks.yaml")
	if err := Save(path, []string{"test_ak"}); err != nil {
		t.Fatalf("Save with nested dir failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) { t.Error("file was not created") }
}

func TestSaveEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := Save(path, nil); err != nil { t.Fatalf("Save nil entries failed: %v", err) }
	loaded, err := Load(path)
	if err != nil { t.Fatalf("Load failed: %v", err) }
	if len(loaded) != 0 { t.Errorf("expected 0 entries, got %d", len(loaded)) }
}

func TestYAMLFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := Save(path, []string{"ak1", "ak2"}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil { t.Fatalf("ReadFile failed: %v", err) }
	expected := "aks:\n    - ak1\n    - ak2\n"
	if string(data) != expected {
		t.Errorf("unexpected YAML format:\ngot:\n%s\nwant:\n%s", string(data), expected)
	}
}
