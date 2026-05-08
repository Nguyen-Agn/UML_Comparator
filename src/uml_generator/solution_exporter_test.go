package uml_generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExport_Success(t *testing.T) {
	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "result.mmd")
	code := "classDiagram\n    A --> B : __1__"

	e := NewSolutionExporter()
	if err := e.Export(code, outPath); err != nil {
		t.Fatalf("Export returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}
	got := string(data)
	if got == "" {
		t.Error("output file is empty")
	}
	if got[len(got)-1] != '\n' {
		t.Error("output file should end with newline")
	}
}

func TestExport_RejectsMissingClassDiagram(t *testing.T) {
	tmp := t.TempDir()
	e := NewSolutionExporter()
	err := e.Export("A --> B", filepath.Join(tmp, "bad.mmd"))
	if err == nil {
		t.Fatal("expected error for code without classDiagram, got nil")
	}
}

func TestExport_RejectsEmptyCode(t *testing.T) {
	tmp := t.TempDir()
	e := NewSolutionExporter()
	err := e.Export("", filepath.Join(tmp, "empty.mmd"))
	if err == nil {
		t.Fatal("expected error for empty code, got nil")
	}
}

func TestExport_RejectsWrongExtension(t *testing.T) {
	tmp := t.TempDir()
	e := NewSolutionExporter()
	err := e.Export("classDiagram\n A-->B", filepath.Join(tmp, "result.txt"))
	if err == nil {
		t.Fatal("expected error for .txt extension, got nil")
	}
}

func TestExport_CreatesParentDir(t *testing.T) {
	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "subdir", "nested", "result.mmd")
	e := NewSolutionExporter()
	if err := e.Export("classDiagram\n A-->B __1__", outPath); err != nil {
		t.Fatalf("expected dir creation to succeed, got: %v", err)
	}
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("output file was not created")
	}
}
