package parser_test

import (
	"strings"
	"testing"
	"uml_compare/src/parser"
)

// TestDrawioParser_Parse_PlainXML tests that the parser correctly reads an uncompressed draw.io file.
// According to check.md:
//   - Must return an mxGraphModel-containing string
//   - Must not return an error on a valid file
func TestDrawioParser_Parse_PlainXML(t *testing.T) {
	p := parser.NewDrawioParser()

	result, _, err := p.Parse("testdata/plain_sample.drawio")
	if err != nil {
		t.Fatalf("Expected no error for plain XML file, got: %v", err)
	}

	if !strings.Contains(string(result), "<mxGraphModel") {
		t.Errorf("Expected result to contain <mxGraphModel>, got: %s", string(result)[:200])
	}

	// Verify the known entity "Animal" is present in the raw XML
	if !strings.Contains(string(result), "Animal") {
		t.Errorf("Expected 'Animal' class to be present in parsed XML")
	}
	if !strings.Contains(string(result), "Dog") {
		t.Errorf("Expected 'Dog' class to be present in parsed XML")
	}
	t.Logf("✔ Plain XML parse successful. Length: %d chars", len(result))
}

// TestDrawioParser_Parse_EmptyPath verifies the guard clause for empty filePath.
// From check.md: must handle empty filePath gracefully.
func TestDrawioParser_Parse_EmptyPath(t *testing.T) {
	p := parser.NewDrawioParser()

	_, _, err := p.Parse("")
	if err == nil {
		t.Fatal("Expected an error when filePath is empty, got nil")
	}
	t.Logf("✔ Empty path correctly rejected with error: %v", err)
}

// TestDrawioParser_Parse_NonExistentFile verifies that missing files return a meaningful error.
func TestDrawioParser_Parse_NonExistentFile(t *testing.T) {
	p := parser.NewDrawioParser()

	_, _, err := p.Parse("testdata/does_not_exist.drawio")
	if err == nil {
		t.Fatal("Expected an error for missing file, got nil")
	}
	t.Logf("✔ Missing file correctly rejected with error: %v", err)
}

// TestDrawioParser_Parse_EdgeCount verifies that the raw XML output contains the expected edges.
// The plain_sample.drawio has 1 edge (Dog --inherits--> Animal).
func TestDrawioParser_Parse_EdgeCount(t *testing.T) {
	p := parser.NewDrawioParser()

	result, _, err := p.Parse("testdata/plain_sample.drawio")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	count := strings.Count(string(result), `edge="1"`)
	if count != 1 {
		t.Errorf("Expected 1 edge in XML, found %d", count)
	}
	t.Logf("✔ Edge count verified: %d edge(s) found", count)
}
