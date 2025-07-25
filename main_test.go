package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindGoFiles(t *testing.T) {
	dir := t.TempDir()
	files := []string{"a.go", "b.go", "c.txt"}
	for _, f := range files {
		os.WriteFile(filepath.Join(dir, f), []byte("package test"), 0644)
	}
	found, err := findGoFiles(dir)
	if err != nil {
		t.Fatalf("ошибка поиска: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("ожидалось 2 go-файла, найдено %d", len(found))
	}
}

func TestParseImports(t *testing.T) {
	dir := t.TempDir()
	code := `package test
import (
	"fmt"
	"os"
)`
	file := filepath.Join(dir, "a.go")
	os.WriteFile(file, []byte(code), 0644)
	imports, err := parseImports(file)
	if err != nil {
		t.Fatalf("ошибка парсинга: %v", err)
	}
	if len(imports) != 2 || imports[0] != "fmt" || imports[1] != "os" {
		t.Errorf("неверные импорты: %v", imports)
	}
}

func TestBuildImportToFileMap(t *testing.T) {
	dir := t.TempDir()
	fileA := filepath.Join(dir, "a.go")
	fileB := filepath.Join(dir, "b.go")
	os.WriteFile(fileA, []byte("package alpha"), 0644)
	os.WriteFile(fileB, []byte("package beta"), 0644)
	m := buildImportToFileMap([]string{fileA, fileB})
	if m["alpha"] != fileA || m["beta"] != fileB {
		t.Errorf("неверная карта: %v", m)
	}
}

func TestDetectCycles(t *testing.T) {
	depMap := DependencyMap{
		"a.go": {"b.go"},
		"b.go": {"a.go"},
		"c.go": {"d.go"},
		"d.go": {},
	}
	cycles := detectCycles(depMap)
	if len(cycles["a.go"]) == 0 || len(cycles["b.go"]) == 0 {
		t.Errorf("цикл не найден: %v", cycles)
	}
	if len(cycles["c.go"]) > 0 {
		t.Errorf("ложный цикл: %v", cycles["c.go"])
	}
}
