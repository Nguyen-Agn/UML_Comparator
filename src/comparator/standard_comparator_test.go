package comparator

import (
	"strings"
	"testing"
	"uml_compare/domain"
)

// mockFuzzy is a simple matcher for testing
type mockFuzzy struct{}

func (m *mockFuzzy) Compare(s1, s2 string) float64 {
	if strings.EqualFold(s1, s2) {
		return 1.0
	}
	return 0.0
}

func TestStandardComparator(t *testing.T) {
	comp := NewStandardComparator()

	mapping := domain.MappingTable{
		"S1": {StudentID: "Stu1", Similarity: 1.0},
	}

	solGraph := &domain.SolutionProcessedUMLGraph{
		Nodes: []domain.SolutionProcessedNode{
			{
				ID:   "S1",
				Name: "Employee",
				Type: "Class",
				Attributes: []domain.SolutionProcessedAttribute{
					{Names: []string{"empId"}, Scope: "-", Types: []string{"int"}},
				},
				Methods: []domain.SolutionProcessedMethod{
					{Names: []string{"work"}, Scope: "+", Outputs: []string{"void"}},
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
					{Name: "work", Scope: "+", Output: "void"},
				},
			},
		},
	}

	report, _ := comp.Compare(solGraph, stuGraph, mapping)

	if len(report.CorrectDetail.Attribute) != 1 || len(report.CorrectDetail.Method) != 1 {
		t.Errorf("Expected perfect match, got: %+v", report)
	}
}

func TestComparatorGenerics(t *testing.T) {
	c := NewStandardComparator()

	sol := &domain.SolutionProcessedUMLGraph{
		Nodes: []domain.SolutionProcessedNode{
			{
				ID:   "N1",
				Name: "Service",
				Type: "Class",
				Attributes: []domain.SolutionProcessedAttribute{
					{Names: []string{"users"}, Types: []string{"List<User>"}, Scope: "-", Kind: "normal"},
				},
			},
			{ID: "N2", Name: "User", Type: "Class"},
		},
	}

	stu := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{
				ID:   "S1",
				Name: "ServiceImp",
				Type: "Class",
				Attributes: []domain.ProcessedAttribute{
					{Name: "users", Type: "ArrayList<NguoiDung>", Scope: "-", Kind: "normal"},
				},
			},
			{ID: "S2", Name: "NguoiDung", Type: "Class"},
		},
	}

	mapping := domain.MappingTable{
		"N1": {StudentID: "S1", Similarity: 1.0},
		"N2": {StudentID: "S2", Similarity: 1.0},
	}

	report, _ := c.Compare(sol, stu, mapping)

	if len(report.CorrectDetail.Attribute) == 0 {
		t.Errorf("Expected attribute match for List<User> vs ArrayList<NguoiDung>")
	}
}

func TestComparatorPointers(t *testing.T) {
	c := NewStandardComparator()

	sol := &domain.SolutionProcessedUMLGraph{
		Nodes: []domain.SolutionProcessedNode{
			{ID: "S1", Name: "User", Type: "Class", Attributes: []domain.SolutionProcessedAttribute{{Names: []string{"id"}, Types: []string{"int"}}}},
		},
	}
	stu := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "Stu1", Name: "User", Type: "Interface", Attributes: []domain.ProcessedAttribute{{Name: "id", Type: "string"}}},
		},
	}

	mapping := domain.MappingTable{"S1": {StudentID: "Stu1", Similarity: 1.0}}
	report, _ := c.Compare(sol, stu, mapping)

	// Node mismatch
	if len(report.WrongDetail.Class) == 0 {
		t.Fatal("Expected node type mismatch")
	}
	if report.WrongDetail.Class[0].Sol == nil || report.WrongDetail.Class[0].Stu == nil {
		t.Error("WrongDetail.Class pointers should be non-nil")
	}

	// Attr mismatch
	if len(report.WrongDetail.Attribute) == 0 {
		t.Fatal("Expected attribute type mismatch")
	}
	if report.WrongDetail.Attribute[0].Sol == nil || report.WrongDetail.Attribute[0].Stu == nil {
		t.Error("WrongDetail.Attribute pointers should be non-nil")
	}
}

func TestComparatorOptionalParams(t *testing.T) {
	c := NewStandardComparator()

	sol := &domain.SolutionProcessedUMLGraph{
		Nodes: []domain.SolutionProcessedNode{
			{ID: "S1", Name: "User", Type: "Class", Attributes: []domain.SolutionProcessedAttribute{{Names: []string{"id", "code", "identify"}, Types: []string{"int", "Interge", "String"}}}},
		},
	}
	stu := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "Stu1", Name: "User", Type: "Class", Attributes: []domain.ProcessedAttribute{{Name: "id", Type: "String"}}},
		},
	}

	mapping := domain.MappingTable{"S1": {StudentID: "Stu1", Similarity: 1.0}}
	report, _ := c.Compare(sol, stu, mapping)

	// Wrong Detail?
	if len(report.WrongDetail.Class) > 0 {
		t.Fatal("Expected Pass")
	}
	if len(report.WrongDetail.Attribute) > 0 {
		t.Fatal("Expected Pass")
	}
	if len(report.WrongDetail.Method) > 0 {
		t.Fatal("Expected Pass")
	}

	if len(report.MissingDetail.Class) > 0 || len(report.MissingDetail.Attribute) > 0 || len(report.MissingDetail.Method) > 0 {
		t.Fatal("Expected Pass")
	}

}
