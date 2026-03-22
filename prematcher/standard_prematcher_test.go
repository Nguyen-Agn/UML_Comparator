package prematcher

import (
	"reflect"
	"strings"
	"testing"
	"uml_compare/domain"
)

func TestParseAttribute(t *testing.T) {
	matcher := NewStandardPreMatcher()

	tests := []struct {
		input    string
		expected domain.ProcessedAttribute
	}{
		{"- name : String", domain.ProcessedAttribute{Scope: "-", Name: "name", Type: "String", Kind: "normal"}},
		{"+ age: int = 10", domain.ProcessedAttribute{Scope: "+", Name: "age", Type: "int", Kind: "normal"}},
		{"# isValid : boolean", domain.ProcessedAttribute{Scope: "#", Name: "isValid", Type: "boolean", Kind: "normal"}},
		{"id: UUID", domain.ProcessedAttribute{Scope: "+", Name: "id", Type: "UUID", Kind: "normal"}}, // default scope
		{"{static} - count : int", domain.ProcessedAttribute{Scope: "-", Name: "count", Type: "int", Kind: "static"}},
		{"# const VERSION : string = \"1.0\"", domain.ProcessedAttribute{Scope: "#", Name: "VERSION", Type: "string", Kind: "final"}},
		{"+ static final INSTANCE : App", domain.ProcessedAttribute{Scope: "+", Name: "INSTANCE", Type: "App", Kind: "static-final"}},
	}

	for _, tt := range tests {
		result := matcher.parseAttribute(tt.input)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("parseAttribute(%q) = %+v, want %+v", tt.input, result, tt.expected)
		}
	}
}

func TestParseMethod(t *testing.T) {
	matcher := NewStandardPreMatcher()

	tests := []struct {
		className string
		input    string
		expected domain.ProcessedMethod
	}{
		{
			"Calculator",
			"+ calculateSum(a: int, b: int) : int",
			domain.ProcessedMethod{
				Scope:  "+",
				Name:   "calculateSum",
				Type:   "custom",
				Output: "int",
				Inputs: []domain.MethodParam{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "int"},
				},
				Kind: "normal",
			},
		},
		{
			"System",
			"- init()",
			domain.ProcessedMethod{
				Scope:  "-",
				Name:   "init",
				Type:   "constructor",
				Output: "",
				Inputs: []domain.MethodParam{},
				Kind:   "normal",
			},
		},
		{
			"DataHandler",
			"doSomething(data: string)",
			domain.ProcessedMethod{
				Scope:  "+",
				Name:   "doSomething",
				Type:   "custom",
				Output: "void",
				Inputs: []domain.MethodParam{
					{Name: "data", Type: "string"},
				},
				Kind: "normal",
			},
		},
		{
			"User",
			"+ getName() : string",
			domain.ProcessedMethod{
				Scope:  "+",
				Name:   "getName",
				Type:   "getter",
				Output: "string",
				Inputs: []domain.MethodParam{},
				Kind:   "normal",
			},
		},
		{
			"User",
			"+ setName(name: string)",
			domain.ProcessedMethod{
				Scope:  "+",
				Name:   "setName",
				Type:   "setter",
				Output: "void",
				Inputs: []domain.MethodParam{
					{Name: "name", Type: "string"},
				},
				Kind: "normal",
			},
		},
		{
			"User",
			"+ user(id: int)", // Constructor with args, case insensitive match
			domain.ProcessedMethod{
				Scope:  "+",
				Name:   "user",
				Type:   "constructor",
				Output: "",
				Inputs: []domain.MethodParam{
					{Name: "id", Type: "int"},
				},
				Kind: "normal",
			},
		},
		{
			"Shape",
			"+ {abstract} draw()",
			domain.ProcessedMethod{
				Scope:  "+",
				Name:   "draw",
				Type:   "custom",
				Output: "void",
				Inputs: []domain.MethodParam{},
				Kind:   "abstract",
			},
		},
		{
			"User",
			"+ name : string {getter}",
			domain.ProcessedMethod{
				Scope:  "+",
				Name:   "name",
				Type:   "getter",
				Output: "string",
				Inputs: []domain.MethodParam{},
				Kind:   "normal",
			},
		},
	}

	for _, tt := range tests {
		emptyG := make(map[string]bool)
		emptyS := make(map[string]bool)
		// For TestParseMethod, we'll assume attributes exist for existing getters/setters
		attrs := []domain.ProcessedAttribute{}
		if strings.HasPrefix(tt.expected.Name, "get") || strings.EqualFold(tt.expected.Type, "getter") {
			base := tt.expected.Name
			if len(base) > 3 {
				base = base[3:]
			}
			attrs = append(attrs, domain.ProcessedAttribute{Name: base})
		}
		if strings.HasPrefix(tt.expected.Name, "set") || strings.EqualFold(tt.expected.Type, "setter") {
			base := tt.expected.Name
			if len(base) > 3 {
				base = base[3:]
			}
			attrs = append(attrs, domain.ProcessedAttribute{Name: base})
		}

		result := matcher.parseMethod(tt.input, tt.className, attrs, emptyG, emptyS)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("parseMethod(%q, %q) = %+v, want %+v", tt.input, tt.className, result, tt.expected)
		}
	}
}

func TestFuzzyGetterSetter(t *testing.T) {
	p := NewStandardPreMatcher()
	attrs := []domain.ProcessedAttribute{
		{Name: "firstName", Type: "string"},
		{Name: "age", Type: "int"},
	}

	tests := []struct {
		name     string
		input    string
		claimedG map[string]bool
		claimedS map[string]bool
		expected string // Expected method type
	}{
		{"Exact match", "+ getFirstName() : string", map[string]bool{}, map[string]bool{}, "getter"},
		{"Fuzzy match (80%+)", "+ getFirst_Name() : string", map[string]bool{}, map[string]bool{}, "getter"},
		{"Case insensitive", "+ GETFIRSTNAME() : string", map[string]bool{}, map[string]bool{}, "getter"},
		{"No match attribute", "+ getSalary() : int", map[string]bool{}, map[string]bool{}, "custom"},
		{"Already claimed", "+ getFirstName() : string", map[string]bool{"firstName": true}, map[string]bool{}, "custom"},
		{"Setter exact", "+ setAge(a: int)", map[string]bool{}, map[string]bool{}, "setter"},
		{"Setter mismatch", "+ setScore(s: int)", map[string]bool{}, map[string]bool{}, "custom"},
	}

	for _, tt := range tests {
		res := p.parseMethod(tt.input, "User", attrs, tt.claimedG, tt.claimedS)
		if res.Type != tt.expected {
			t.Errorf("Test %s: parseMethod type = %s, want %s", tt.name, res.Type, tt.expected)
		}
	}
}

func TestCalculateArchWeight(t *testing.T) {
	matcher := NewStandardPreMatcher()

	// Bit 29-31: Loại Class (Interface=2) -> 2 << 29
	// Bit 28: Thừa kế (1) -> 1 << 28
	// Bit 24-27: Số Interface (1) -> 1 << 24
	// Bit 18-23: Số Method (2) -> 2 << 18
	// Bit 13-17: Số Attribute (3) -> 3 << 13
	// Bit 9-12: Số Class liên quan (0) -> 0 << 9
	// Bit 6-8: Custom type (0) -> 0 << 6
	// Bit 2-5: Static members (0) -> 0 << 2
	
	weight := matcher.calculateArchWeight("Interface", true, 1, 2, 3, 0, 0, 0)
	
	expectedType := uint32(2) << 29
	expectedInherit := uint32(1) << 28
	expectedImpl := uint32(1) << 24
	expectedMeth := uint32(2) << 18
	expectedAttr := uint32(3) << 13
	expectedTotal := expectedType | expectedInherit | expectedImpl | expectedMeth | expectedAttr

	if weight != expectedTotal {
		t.Errorf("calculateArchWeight() = %v, want %v", weight, expectedTotal)
	}
}

func TestProcessGraph(t *testing.T) {
	matcher := NewStandardPreMatcher()

	graph := &domain.UMLGraph{
		ID: "G1",
		Nodes: []domain.UMLNode{
			{
				ID:   "N1",
				Name: "Animal",
				Type: "Class",
				Attributes: []string{
					"+ age : int",
				},
				Methods: []string{
					"+ makeSound() : void",
				},
			},
			{
				ID:   "N2",
				Name: "Dog",
				Type: "Class",
				Attributes: []string{
					"- breed : String {getter, setter}",
				},
				Methods: []string{
					"+ makeSound() : void",
					"+ fetch(item: String) : boolean",
				},
			},
		},
		Edges: []domain.UMLEdge{
			{
				SourceID:     "N2",
				TargetID:     "N1",
				RelationType: "Inheritance",
			},
		},
	}

	processed, err := matcher.Process(graph)
	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}

	if len(processed.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(processed.Nodes))
	}

	// Verify N2 (Dog)
	var dogNode *domain.ProcessedNode
	for _, n := range processed.Nodes {
		if n.ID == "N2" {
			dogNode = &n
			break
		}
	}

	if dogNode == nil {
		t.Fatal("Could not find N2 (Dog)")
	}

	if dogNode.Inherits != "N1" {
		t.Errorf("Expected N2 to inherit from N1, got %s", dogNode.Inherits)
	}

	if len(dogNode.Attributes) != 1 || dogNode.Attributes[0].Name != "breed" {
		t.Errorf("Expected 1 attribute 'breed', got %+v", dogNode.Attributes)
	}

	if len(dogNode.Methods) != 4 {
		t.Errorf("Expected 4 methods (makeSound, fetch, getBreed, setBreed), got %d", len(dogNode.Methods))
		for _, m := range dogNode.Methods {
			t.Logf("Method: %s (%s)", m.Name, m.Type)
		}
	}

	// Verify getBreed and setBreed were generated
	var hasGetBreed, hasSetBreed bool
	for _, m := range dogNode.Methods {
		if m.Name == "getBreed" && m.Type == "getter" && m.Output == "String" {
			hasGetBreed = true
		}
		if m.Name == "setBreed" && m.Type == "setter" && m.Output == "void" {
			hasSetBreed = true
		}
	}
	if !hasGetBreed {
		t.Error("Missing generated getBreed method")
	}
	if !hasSetBreed {
		t.Error("Missing generated setBreed method")
	}

	// ArchWeight check
	// Class(1)<<29 | Inherit(1)<<28 | Interfaces(0)<<24 | Methods(2)<<18 | Attributes(1)<<13
	// Note: Methods count is 2 (makeSound, fetch) because getBreed and setBreed are ignored!
	expectedWeight := (uint32(1) << 29) | (uint32(1) << 28) | (uint32(2) << 18) | (uint32(1) << 13)
	if dogNode.ArchWeight != expectedWeight {
		t.Errorf("Expected ArchWeight %d, got %d", expectedWeight, dogNode.ArchWeight)
	}
}
