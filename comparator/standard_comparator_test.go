package comparator

import (
	"testing"
	"uml_compare/domain"
	"uml_compare/matcher"
)

func TestStandardComparator(t *testing.T) {
	fz := matcher.NewLevenshteinMatcher()
	comp := NewStandardComparator(fz)

	// Setup Mapping: Solution [S1] -> Student [Stu1]
	mapping := domain.MappingTable{
		"S1": {StudentID: "Stu1", Similarity: 1.0},
	}

	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{
				ID:   "S1",
				Name: "Employee",
				Type: "Class",
				Attributes: []domain.ProcessedAttribute{
					{Name: "empId", Scope: "-", Type: "int"},
				},
				Methods: []domain.ProcessedMethod{
					{Name: "Employee", Scope: "+", Inputs: []domain.MethodParam{{Name: "id", Type: "int"}}}, // Constructor
					{Name: "work", Scope: "+", Output: "void"},
				},
			},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{
				ID:   "Stu1",
				Name: "Employee",
				Type: "Class",
				Attributes: []domain.ProcessedAttribute{
					{Name: "empId", Scope: "-", Type: "int"},
				},
				Methods: []domain.ProcessedMethod{
					{Name: "Employee", Scope: "+", Inputs: []domain.MethodParam{{Name: "id", Type: "int"}}},
					{Name: "work", Scope: "+", Output: "void"},
				},
			},
		},
	}

	report, _ := comp.Compare(solGraph, stuGraph, mapping)

	if len(report.MissingDetail.Attribute) > 0 || len(report.MissingDetail.Method) > 0 || 
	   len(report.WrongDetail.Attribute) > 0 || len(report.WrongDetail.Method) > 0 {
		t.Errorf("Expected perfect match, got errors: %+v", report)
	}
}

func TestComparatorAdvancedRules(t *testing.T) {
	fz := matcher.NewLevenshteinMatcher()
	comp := NewStandardComparator(fz)

	// Scenario:
	// Sol: Node "A" has method with param of type "B".
	// Stu: Node "A_Stu" has method with param of type "B_Stu".
	// Mapping: A -> A_Stu, B -> B_Stu.
	mapping := domain.MappingTable{
		"SolA": {StudentID: "StuA", Similarity: 1.0},
		"SolB": {StudentID: "StuB", Similarity: 1.0},
	}

	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "SolA", Name: "Manager", Type: "Class", Methods: []domain.ProcessedMethod{
				{Name: "init", Scope: "+", Inputs: []domain.MethodParam{{Type: "int"}, {Type: "String"}}},
				{Name: "process", Scope: "+", Output: "Task", Inputs: []domain.MethodParam{{Type: "Task"}}},
			}},
			{ID: "SolB", Name: "Task", Type: "Class"},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "StuA", Name: "Manager", Type: "Class", Methods: []domain.ProcessedMethod{
				// Constructor: unordered params (String, int) vs (int, String) -> Should Pass match logic but maybe show wrong details if not handled as ctor
				{Name: "Manager", Scope: "+", Inputs: []domain.MethodParam{{Type: "String"}, {Type: "int"}}},
				// Normal Method: mapped type Task_Stu vs Task -> Should Pass
				{Name: "process", Scope: "+", Output: "Task_Stu", Inputs: []domain.MethodParam{{Type: "Task_Stu"}}},
			}},
			{ID: "StuB", Name: "Task_Stu", Type: "Class"},
		},
	}

	report, _ := comp.Compare(solGraph, stuGraph, mapping)

	if len(report.MissingDetail.Method) > 0 {
		t.Errorf("Advanced rules failed. Expected match via TypeMap & Constructor Params. Report: %+v", report)
	}
}

func TestComparatorScopeMismatch(t *testing.T) {
	fz := matcher.NewLevenshteinMatcher()
	comp := NewStandardComparator(fz)

	mapping := domain.MappingTable{"S1": {StudentID: "Stu1", Similarity: 1.0}}

	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{
				ID:   "S1",
				Name: "Employee",
				Type: "Class",
				Attributes: []domain.ProcessedAttribute{
					{Name: "empId", Scope: "-", Type: "int"},
				},
				Methods: []domain.ProcessedMethod{
					{Name: "work", Scope: "+", Output: "void"},
				},
			},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{
				ID:   "Stu1",
				Name: "Employee",
				Type: "Class",
				Attributes: []domain.ProcessedAttribute{
					{Name: "empId", Scope: "+", Type: "int"}, // WRONG SCOPE
				},
				Methods: []domain.ProcessedMethod{
					{Name: "work", Scope: "-", Output: "void"}, // WRONG SCOPE
				},
			},
		},
	}

	report, _ := comp.Compare(solGraph, stuGraph, mapping)

	if len(report.WrongDetail.Attribute) != 1 {
		t.Errorf("Expected 1 WrongDetail.Attribute due to scope mismatch, got %d", len(report.WrongDetail.Attribute))
	}
	if len(report.WrongDetail.Method) != 1 {
		t.Errorf("Expected 1 WrongDetail.Method due to scope mismatch, got %d", len(report.WrongDetail.Method))
	}
}
func TestComparatorPointers(t *testing.T) {
	fz := matcher.NewLevenshteinMatcher()
	comp := NewStandardComparator(fz)

	mapping := domain.MappingTable{"S1": {StudentID: "Stu1", Similarity: 1.0}}

	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "S1", Name: "User", Type: "Class", Attributes: []domain.ProcessedAttribute{{Name: "id", Type: "string"}}},
		},
	}
	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "Stu1", Name: "User", Type: "Interface"}, // Type mismatch
		},
	}

	report, _ := comp.Compare(solGraph, stuGraph, mapping)

	// Check Node mismatch
	if len(report.WrongDetail.Class) == 0 {
		t.Fatal("Expected node type mismatch")
	}
	diff := report.WrongDetail.Class[0]
	if diff.Sol == nil || diff.Stu == nil {
		t.Errorf("Expected both Sol and Stu pointers for WrongDetail, got Sol=%v, Stu=%v", diff.Sol, diff.Stu)
	}
	if diff.Sol.Type != "Class" || diff.Stu.Type != "Interface" {
		t.Errorf("Pointer data incorrect: Sol=%s, Stu=%s", diff.Sol.Type, diff.Stu.Type)
	}

	// Check missing attribute
	if len(report.MissingDetail.Attribute) == 0 {
		t.Fatal("Expected missing attribute")
	}
	aDiff := report.MissingDetail.Attribute[0]
	if aDiff.Sol == nil || aDiff.Stu != nil {
		t.Errorf("Expected Sol pointer and nil Stu pointer for MissingDetail, got Sol=%v, Stu=%v", aDiff.Sol, aDiff.Stu)
	}
}
