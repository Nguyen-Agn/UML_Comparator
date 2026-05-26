package parser_test

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"uml_compare/cipher"
	"uml_compare/src/parser"
)

func TestSolutionParser_Parse(t *testing.T) {
	// 1. Create a dummy drawio and a dummy mermaid file
	drawioContent := `<mxGraphModel><root><mxCell id="0"/></root></mxGraphModel>`
	mermaidContent := `classDiagram
class A {
    +int a
}
%% This is a comment
class B
A <|-- B`

	// Write them to temp files
	tmpDrawio, _ := os.CreateTemp("", "test_*.drawio")
	defer os.Remove(tmpDrawio.Name())
	tmpDrawio.WriteString(drawioContent)
	tmpDrawio.Close()

	tmpMermaid, _ := os.CreateTemp("", "test_*.mmd")
	defer os.Remove(tmpMermaid.Name())
	tmpMermaid.WriteString(mermaidContent)
	tmpMermaid.Close()

	// 2. Encrypt both using cipher
	cipherTool := cipher.New()
	
	tmpDrawioSol, _ := os.CreateTemp("", "test_*.solution")
	defer os.Remove(tmpDrawioSol.Name())
	tmpDrawioSol.Close()
	err := cipherTool.Encrypt(tmpDrawio.Name(), tmpDrawioSol.Name())
	if err != nil {
		t.Fatalf("Failed to encrypt drawio: %v", err)
	}

	tmpMermaidSol, _ := os.CreateTemp("", "test_*.solution")
	defer os.Remove(tmpMermaidSol.Name())
	tmpMermaidSol.Close()
	err = cipherTool.Encrypt(tmpMermaid.Name(), tmpMermaidSol.Name())
	if err != nil {
		t.Fatalf("Failed to encrypt mermaid: %v", err)
	}

	// 3. Test parsing the solutions
	solParser := parser.NewSolutionParserDefault()

	// Drawio test
	rawDrawio, typeDrawio, err := solParser.Parse(tmpDrawioSol.Name())
	if err != nil {
		t.Fatalf("Failed to parse drawio solution: %v", err)
	}
	if typeDrawio != "drawio" {
		t.Errorf("Expected drawio type, got %s", typeDrawio)
	}
	if string(rawDrawio) != drawioContent {
		t.Errorf("Expected raw XML to be preserved. Got: %s", string(rawDrawio))
	}

	// Mermaid test
	rawMermaid, typeMermaid, err := solParser.Parse(tmpMermaidSol.Name())
	if err != nil {
		t.Fatalf("Failed to parse mermaid solution: %v", err)
	}
	if typeMermaid != "mermaid" {
		t.Errorf("Expected mermaid type, got %s", typeMermaid)
	}
	expectedMermaid := "class A {\n    +int a\n}\nclass B\nA <|-- B"
	if string(rawMermaid) != expectedMermaid {
		t.Errorf("Expected cleaned mermaid. Got:\n%s\nExpected:\n%s", string(rawMermaid), expectedMermaid)
	}
}

func TestSolutionParser_InvalidHeader(t *testing.T) {
	tmp, _ := os.CreateTemp("", "test_*.solution")
	defer os.Remove(tmp.Name())
	tmp.WriteString("INVALID_HEADER\n" + base64.StdEncoding.EncodeToString([]byte("test")))
	tmp.Close()

	solParser := parser.NewSolutionParserDefault()
	_, _, err := solParser.Parse(tmp.Name())
	if err == nil || !strings.Contains(err.Error(), "missing header") {
		t.Errorf("Expected missing header error, got %v", err)
	}
}
