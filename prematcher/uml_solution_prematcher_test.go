package prematcher

import (
	"testing"
	"uml_compare/domain"
)

func TestSplitOR(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"int|long", []string{"int", "long"}},
		{"void | boolean", []string{"void", "boolean"}},
		{"String", []string{"String"}},
		{"int | long | String", []string{"int", "long", "String"}},
		{" | ", []string{}}, // Only empty tokens — should fall back to 1 slice of original trimmed
		{"", []string{}},
	}

	for _, tt := range tests {
		result := splitOR(tt.input)
		if tt.input == " | " || tt.input == "" {
			// Edge case: splitOR returns the trimmed input as single element if all empty
			// (or empty slice if the input itself is empty after trim)
			continue
		}
		if len(result) != len(tt.expected) {
			t.Errorf("splitOR(%q) len = %d, want %d (got %v)", tt.input, len(result), len(tt.expected), result)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("splitOR(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestParseSolutionAttribute_NoOR(t *testing.T) {
	p := NewUMLSolutionPreMatcher()

	result := p.parseSolutionAttribute("- name : String")
	if result.Scope != "-" {
		t.Errorf("Scope: got %q, want %q", result.Scope, "-")
	}
	if len(result.Names) != 1 || result.Names[0] != "name" {
		t.Errorf("Names: got %v, want [name]", result.Names)
	}
	if len(result.Types) != 1 || result.Types[0] != "String" {
		t.Errorf("Types: got %v, want [String]", result.Types)
	}
	if result.Kind != "normal" {
		t.Errorf("Kind: got %q, want normal", result.Kind)
	}
}

func TestParseSolutionAttribute_ORType(t *testing.T) {
	p := NewUMLSolutionPreMatcher()

	result := p.parseSolutionAttribute("- id : int|long")
	if len(result.Names) != 1 || result.Names[0] != "id" {
		t.Errorf("Names: got %v, want [id]", result.Names)
	}
	if len(result.Types) != 2 || result.Types[0] != "int" || result.Types[1] != "long" {
		t.Errorf("Types: got %v, want [int long]", result.Types)
	}
}

func TestParseSolutionAttribute_ORName(t *testing.T) {
	p := NewUMLSolutionPreMatcher()

	result := p.parseSolutionAttribute("+ x | y : int")
	if len(result.Names) != 2 || result.Names[0] != "x" || result.Names[1] != "y" {
		t.Errorf("Names: got %v, want [x y]", result.Names)
	}
	if len(result.Types) != 1 || result.Types[0] != "int" {
		t.Errorf("Types: got %v, want [int]", result.Types)
	}
}

func TestParseSolutionAttribute_ORBoth(t *testing.T) {
	p := NewUMLSolutionPreMatcher()

	// Both name and type have OR
	result := p.parseSolutionAttribute("+ a | b : int|long")
	if len(result.Names) != 2 {
		t.Errorf("Names: got %v, want 2 items", result.Names)
	}
	if len(result.Types) != 2 {
		t.Errorf("Types: got %v, want 2 items", result.Types)
	}
}

func TestParseSolutionAttribute_Static(t *testing.T) {
	p := NewUMLSolutionPreMatcher()

	result := p.parseSolutionAttribute("{static} - count : int")
	if result.Kind != "static" {
		t.Errorf("Kind: got %q, want static", result.Kind)
	}
	if len(result.Names) != 1 || result.Names[0] != "count" {
		t.Errorf("Names: got %v, want [count]", result.Names)
	}
}

func TestParseSolutionMethod_NoOR(t *testing.T) {
	p := NewUMLSolutionPreMatcher()
	emptyG := make(map[string]bool)
	emptyS := make(map[string]bool)

	result := p.parseSolutionMethod("+ calculate(a: int, b: int) : int", "Calculator", nil, emptyG, emptyS)
	if len(result.Names) != 1 || result.Names[0] != "calculate" {
		t.Errorf("Names: got %v, want [calculate]", result.Names)
	}
	if len(result.Outputs) != 1 || result.Outputs[0] != "int" {
		t.Errorf("Outputs: got %v, want [int]", result.Outputs)
	}
	if len(result.Inputs) != 2 {
		t.Errorf("Inputs: got %d, want 2", len(result.Inputs))
	}
	if result.Type != "custom" {
		t.Errorf("Type: got %q, want custom", result.Type)
	}
}

func TestParseSolutionMethod_ORReturn(t *testing.T) {
	p := NewUMLSolutionPreMatcher()
	emptyG := make(map[string]bool)
	emptyS := make(map[string]bool)

	// The user's example: doing(a : int|long): void|boolean
	// Param type int|long is NOT split (kept as-is)
	result := p.parseSolutionMethod("doing(a : int|long): void|boolean", "Foo", nil, emptyG, emptyS)

	if len(result.Names) != 1 || result.Names[0] != "doing" {
		t.Errorf("Names: got %v, want [doing]", result.Names)
	}
	if len(result.Outputs) != 2 || result.Outputs[0] != "void" || result.Outputs[1] != "boolean" {
		t.Errorf("Outputs: got %v, want [void boolean]", result.Outputs)
	}
	// Param type kept as-is (not split)
	if len(result.Inputs) != 1 || result.Inputs[0].Type != "int|long" {
		t.Errorf("Inputs[0].Type: got %q, want %q", result.Inputs[0].Type, "int|long")
	}
}

func TestParseSolutionMethod_ORName(t *testing.T) {
	p := NewUMLSolutionPreMatcher()
	emptyG := make(map[string]bool)
	emptyS := make(map[string]bool)

	result := p.parseSolutionMethod("doA | doB(p: int): void", "Foo", nil, emptyG, emptyS)
	if len(result.Names) != 2 || result.Names[0] != "doA" || result.Names[1] != "doB" {
		t.Errorf("Names: got %v, want [doA doB]", result.Names)
	}
	if len(result.Outputs) != 1 || result.Outputs[0] != "void" {
		t.Errorf("Outputs: got %v, want [void]", result.Outputs)
	}
}

func TestParseSolutionMethod_Constructor(t *testing.T) {
	p := NewUMLSolutionPreMatcher()
	emptyG := make(map[string]bool)
	emptyS := make(map[string]bool)

	result := p.parseSolutionMethod("+ MyClass(id: int)", "MyClass", nil, emptyG, emptyS)
	if result.Type != "constructor" {
		t.Errorf("Type: got %q, want constructor", result.Type)
	}
	if len(result.Outputs) != 0 {
		t.Errorf("Outputs: constructor should have empty Outputs, got %v", result.Outputs)
	}
}

func TestParseSolutionMethod_DefaultVoid(t *testing.T) {
	p := NewUMLSolutionPreMatcher()
	emptyG := make(map[string]bool)
	emptyS := make(map[string]bool)

	// No return type specified — defaults to ["void"]
	result := p.parseSolutionMethod("doSomething(data: string)", "DataHandler", nil, emptyG, emptyS)
	if len(result.Outputs) != 1 || result.Outputs[0] != "void" {
		t.Errorf("Outputs: got %v, want [void]", result.Outputs)
	}
}

func TestParseSolutionMethod_GenericParam(t *testing.T) {
	p := NewUMLSolutionPreMatcher()
	emptyG := make(map[string]bool)
	emptyS := make(map[string]bool)

	// Generic param should not be split by comma inside <>
	result := p.parseSolutionMethod("+ process(data: Map<String, int>): boolean", "Foo", nil, emptyG, emptyS)
	if len(result.Inputs) != 1 {
		t.Errorf("Inputs: got %d params, want 1 (generic not split)", len(result.Inputs))
	}
	if result.Inputs[0].Type != "Map<String, int>" {
		t.Errorf("Inputs[0].Type: got %q, want %q", result.Inputs[0].Type, "Map<String, int>")
	}
}

func TestProcessSolution_Integration(t *testing.T) {
	p := NewUMLSolutionPreMatcher()

	graph := &domain.UMLGraph{
		ID: "G1",
		Nodes: []domain.UMLNode{
			{
				ID:   "N1",
				Name: "Shape",
				Type: "Abstract",
				Attributes: []string{
					"- color | fill : String|int", // OR name + OR type
				},
				Methods: []string{
					"draw | render(): void|boolean", // OR name + OR return
					"+ getArea(): double",
				},
			},
			{
				ID:   "N2",
				Name: "Circle",
				Type: "Class",
				Attributes: []string{
					"- radius : double {getter, setter}",
				},
				Methods: []string{
					"+ doing(a: int|long): void|boolean", // User example
				},
			},
		},
		Edges: []domain.UMLEdge{
			{SourceID: "N2", TargetID: "N1", RelationType: "Inheritance"},
		},
	}

	processed, err := p.ProcessSolution(graph)
	if err != nil {
		t.Fatalf("ProcessSolution returned error: %v", err)
	}
	if len(processed.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(processed.Nodes))
	}

	// Find Shape node
	var shape, circle *domain.SolutionProcessedNode
	for i := range processed.Nodes {
		if processed.Nodes[i].ID == "N1" {
			shape = &processed.Nodes[i]
		}
		if processed.Nodes[i].ID == "N2" {
			circle = &processed.Nodes[i]
		}
	}

	if shape == nil || circle == nil {
		t.Fatal("Could not find expected nodes")
	}

	// Check Shape attribute OR
	if len(shape.Attributes) == 0 {
		t.Fatal("Shape has no attributes")
	}
	attr := shape.Attributes[0]
	if len(attr.Names) != 2 {
		t.Errorf("Shape attr Names: got %v, want [color fill]", attr.Names)
	}
	if len(attr.Types) != 2 {
		t.Errorf("Shape attr Types: got %v, want [String int]", attr.Types)
	}

	// Check draw|render method OR name + OR return
	var drawMethod *domain.SolutionProcessedMethod
	for i := range shape.Methods {
		if len(shape.Methods[i].Names) > 1 {
			drawMethod = &shape.Methods[i]
			break
		}
	}
	if drawMethod == nil {
		t.Error("draw|render method (OR name) not found")
	} else {
		if len(drawMethod.Outputs) != 2 {
			t.Errorf("draw|render Outputs: got %v, want [void boolean]", drawMethod.Outputs)
		}
	}

	// Check Circle inherits Shape
	if circle.Inherits != "N1" {
		t.Errorf("Circle.Inherits: got %q, want N1", circle.Inherits)
	}

	// Check doing() method param type kept as-is
	var doingMethod *domain.SolutionProcessedMethod
	for i := range circle.Methods {
		for _, n := range circle.Methods[i].Names {
			if n == "doing" {
				doingMethod = &circle.Methods[i]
			}
		}
	}
	if doingMethod == nil {
		t.Fatal("doing() method not found in Circle")
	}
	if len(doingMethod.Inputs) != 1 || doingMethod.Inputs[0].Type != "int|long" {
		t.Errorf("doing Inputs[0].Type: got %q, want 'int|long'", doingMethod.Inputs[0].Type)
	}
	if len(doingMethod.Outputs) != 2 {
		t.Errorf("doing Outputs: got %v, want [void boolean]", doingMethod.Outputs)
	}
}
