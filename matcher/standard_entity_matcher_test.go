package matcher

import (
	"testing"

	"uml_compare/domain"
)

// buildSolGraph is a helper that creates a SolutionProcessedUMLGraph
// from simple (id, name, nodeType, archWeight) tuples for test brevity.
// Names is a single-element slice — no OR variants needed for most tests.
func buildSolGraph(nodes ...domain.SolutionProcessedNode) *domain.SolutionProcessedUMLGraph {
	return &domain.SolutionProcessedUMLGraph{Nodes: nodes}
}

func solNode(id, name, nodeType string, archWeight uint32) domain.SolutionProcessedNode {
	return domain.SolutionProcessedNode{
		ID:         id,
		Name:       name,
		Type:       nodeType,
		ArchWeight: archWeight,
	}
}

func buildStuGraph(nodes ...domain.ProcessedNode) *domain.ProcessedUMLGraph {
	return &domain.ProcessedUMLGraph{Nodes: nodes}
}

func stuNode(id, name, nodeType string, archWeight uint32) domain.ProcessedNode {
	return domain.ProcessedNode{ID: id, Name: name, Type: nodeType, ArchWeight: archWeight}
}

func TestFuzzyMatcher(t *testing.T) {
	fz := NewLevenshteinMatcher()

	score := fz.Compare("Account", "Acount")
	if score < 0.8 {
		t.Errorf("Expected score > 0.8 for typo, got %.2f", score)
	}

	scoreExact := fz.Compare("UserService", "userservice") // case-insensitive
	if scoreExact != 1.0 {
		t.Errorf("Expected score 1.0 for exact, got %.2f", scoreExact)
	}

	scoreDiff := fz.Compare("Animal", "Vehicle")
	if scoreDiff > 0.5 {
		t.Errorf("Expected low score for completely different strings, got %.2f", scoreDiff)
	}
}

func TestStandardEntityMatcher(t *testing.T) {
	matcher := NewStandardEntityMatcher(0.75)

	solGraph := buildSolGraph(
		solNode("Sol1", "User", "Class", 100),
		solNode("Sol2", "Order", "Interface", 200),
		solNode("Sol3", "AccountObject", "Class", 300),
	)

	stuGraph := buildStuGraph(
		stuNode("StuA", "User", "Class", 100),
		stuNode("StuB", "Ordr", "Interface", 200),
		stuNode("StuC", "AccountObjct", "Class", 300),
		stuNode("StuD", "Order", "Class", 900), // distractor: same name as Sol2's Order but wrong type+arch
	)

	mapping, err := matcher.Match(solGraph, stuGraph)
	if err != nil {
		t.Fatalf("Matcher returned error: %v", err)
	}

	if len(mapping) != 3 {
		t.Errorf("Expected 3 mapped elements, got %d", len(mapping))
	}

	expected := map[string]string{
		"Sol1": "StuA",
		"Sol2": "StuB",
		"Sol3": "StuC",
	}
	for k, wantStu := range expected {
		if mapping[k].StudentID != wantStu {
			t.Errorf("Expected %s -> %s, got %s", k, wantStu, mapping[k].StudentID)
		}
	}
}

func TestArchWeightPriority(t *testing.T) {
	matcher := NewStandardEntityMatcher(0.75)

	solGraph := buildSolGraph(solNode("S1", "Manager", "Class", 500))
	stuGraph := buildStuGraph(
		stuNode("Stu1", "Managr", "Class", 100), // same typo, different arch
		stuNode("Stu2", "Managr", "Class", 500), // same typo, same arch
	)

	mapping, _ := matcher.Match(solGraph, stuGraph)
	if mapping["S1"].StudentID != "Stu2" {
		t.Errorf("Expected Stu2 prioritized due to ArchWeight, mapped to %s", mapping["S1"].StudentID)
	}
}

func TestInterfaceArchitectureMatch(t *testing.T) {
	matcher := NewStandardEntityMatcher(0.75)

	solGraph := buildSolGraph(solNode("S1", "IProductRepository", "Interface", 1000))
	stuGraph := buildStuGraph(
		stuNode("Stu1", "IUserRepository", "Interface", 1000),
		stuNode("Stu2", "IProductRepository", "Interface", 1000),
		stuNode("Stu3", "IProductRepo", "Interface", 1000),
	)

	mapping, _ := matcher.Match(solGraph, stuGraph)
	if mapping["S1"].StudentID != "Stu2" {
		t.Errorf("Expected Stu2 (exact name) prioritized, mapped to %s", mapping["S1"].StudentID)
	}
}

func TestToleranceArchitectureMatch(t *testing.T) {
	matcher := NewStandardEntityMatcher(0.75)

	solGraph := buildSolGraph(solNode("S1", "UserService", "Class", 539492352))
	stuGraph := buildStuGraph(
		stuNode("Stu1", "UserServic", "Class", 539492352),  // typo, perfect arch
		stuNode("Stu2", "UserService", "Class", 539230208), // exact name, -1 method (within 15%)
		stuNode("Stu3", "UserService", "Class", 537657344), // exact name, bad arch
	)

	mapping, _ := matcher.Match(solGraph, stuGraph)
	if mapping["S1"].StudentID != "Stu2" {
		t.Errorf("Expected Stu2 (exact name + tolerant arch), mapped to %s", mapping["S1"].StudentID)
	}
}

func TestTwoPassMatching(t *testing.T) {
	matcher := NewStandardEntityMatcher(0.8)

	solGraph := buildSolGraph(solNode("S1", "CruiseShip", "Class", 539492352))
	stuGraph := buildStuGraph(
		stuNode("Stu1", "PPShip", "Class", 539492352), // very different name, same arch
	)

	mapping, _ := matcher.Match(solGraph, stuGraph)

	if val, ok := mapping["S1"]; !ok {
		t.Errorf("Expected S1 to be matched in Pass 2")
	} else if val.StudentID != "Stu1" {
		t.Errorf("Expected S1 -> Stu1, got %s", val.StudentID)
	} else if val.Similarity < 0.80 {
		t.Errorf("Expected Similarity >= 0.80 (Arch=1.0*0.7+Text*0.3), got %.4f", val.Similarity)
	}
}
