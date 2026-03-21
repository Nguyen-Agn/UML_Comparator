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

	if len(report.MissingMembers) > 0 || len(report.AttributeErrors) > 0 || len(report.MethodErrors) > 0 {
		t.Errorf("Expected perfect match, got errors: %v", report)
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
			{ID: "SolA", Name: "Manager", Methods: []domain.ProcessedMethod{
				{Name: "init", Scope: "+", Inputs: []domain.MethodParam{{Type: "int"}, {Type: "String"}}},
				{Name: "process", Scope: "+", Output: "Task", Inputs: []domain.MethodParam{{Type: "Task"}}},
			}},
			{ID: "SolB", Name: "Task"},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "StuA", Name: "Manager", Methods: []domain.ProcessedMethod{
				// Constructor: unordered params (String, int) vs (int, String) -> Should Pass
				{Name: "Manager", Scope: "+", Inputs: []domain.MethodParam{{Type: "String"}, {Type: "int"}}},
				// Normal Method: mapped type Task_Stu vs Task -> Should Pass
				{Name: "process", Scope: "+", Output: "Task_Stu", Inputs: []domain.MethodParam{{Type: "Task_Stu"}}},
			}},
			{ID: "StuB", Name: "Task_Stu"},
		},
	}

	report, _ := comp.Compare(solGraph, stuGraph, mapping)

	if len(report.MissingMembers) > 0 || len(report.MethodErrors) > 0 {
		t.Errorf("Advanced rules failed. Expected match via TypeMap & Unordered Constructor Params. Report: %+v", report)
	}
}

func TestComparatorScopeAndFuzzy(t *testing.T) {
	fz := matcher.NewLevenshteinMatcher()
	comp := NewStandardComparator(fz)

	mapping := domain.MappingTable{"S1": {StudentID: "T1", Similarity: 1.0}}

	sol := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{{ID: "S1", Name: "Ship", 
			Attributes: []domain.ProcessedAttribute{{Name: "speed", Scope: "-", Type: "float"}},
		}},
	}

	stu := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{{ID: "T1", Name: "Ship",
			// Wrong Scope: + instead of -
			Attributes: []domain.ProcessedAttribute{{Name: "currentSpeed", Scope: "+", Type: "float"}},
		}},
	}

	report, _ := comp.Compare(sol, stu, mapping)
	if len(report.MissingMembers) == 0 {
		t.Errorf("Expected Attribute Missing/Error due to wrong scope")
	}
}
