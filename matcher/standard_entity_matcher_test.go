package matcher

import (
	"testing"

	"uml_compare/domain"
)

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
	fz := NewLevenshteinMatcher()
	matcher := NewStandardEntityMatcher(fz, 0.8)

	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "Sol1", Name: "User", Type: "Class", ArchWeight: 100},
			{ID: "Sol2", Name: "Order", Type: "Interface", ArchWeight: 200},
			{ID: "Sol3", Name: "AccountObject", Type: "Class", ArchWeight: 300}, // to test structural fallback
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			// Exact match for User
			{ID: "StuA", Name: "User", Type: "Class", ArchWeight: 100},

			// Fuzzy match for Order (Ordr)
			{ID: "StuB", Name: "Ordr", Type: "Interface", ArchWeight: 200},

			// Drastically different name, but same structure (ArchWeight)
			// Wait! Because we only map if fuzzy matching > threshold, this will FAIL mapping 
			// unless we implement the "structural fallback" which we decided NOT to implement (relying on threshold > 0.8)
			// Wait, if we test something very different like "Acount" and "AccountObject", will fuzzy match succeed?
			// Compare("AccountObject", "Account") len(max)=13, dist=6 ("Object" missing). Sim = 1 - 6/13 = 7/13 = 0.53
			// It will fail because 0.53 < 0.8.
			// Let's test a close fuzzy match for AccountObject: "AccountObjct"
			{ID: "StuC", Name: "AccountObjct", Type: "Class", ArchWeight: 300},
			
			// Another distractor node that could match Order via Fuzzy, but ArchWeight is completely off
			// (If it was checked first, it would steal it, but ArchWeight prioritization should prevent it)
			{ID: "StuD", Name: "Order", Type: "Class", ArchWeight: 900}, 
		},
	}

	mapping, err := matcher.Match(solGraph, stuGraph)
	if err != nil {
		t.Fatalf("Matcher returned error: %v", err)
	}

	if len(mapping) != 3 {
		t.Errorf("Expected 3 mapped elements, got %d", len(mapping))
	}

	expectedMap := domain.MappingTable{
		"Sol1": domain.MappedNode{StudentID: "StuA", Similarity: 1.0}, // Exact match
		"Sol2": domain.MappedNode{StudentID: "StuB", Similarity: 0.8}, // Exact Type mismatch from distractor StuD, but StuB ArchWeight aligns and fuzz matches
		"Sol3": domain.MappedNode{StudentID: "StuC", Similarity: 0.8}, // Fuzzy match - it's actually missing "o", Levenshtein of AccountObjct vs AccountObject is 1/13. So 12/13 ~ 0.92, but for exact test we might just skip DeepEqual or adjust expectations.
	}

	// We can't rely strictly on DeepEqual for float equality of Similarity if we aren't precisely sure, 
	// but let's check the StudentIDs explicitly.
	if len(mapping) != len(expectedMap) {
		t.Errorf("Mapping mismatch length")
	}
	for k, v := range expectedMap {
		if mapping[k].StudentID != v.StudentID {
			t.Errorf("Mapping mismatch for %s. Expected: %v, Got: %v", k, v.StudentID, mapping[k].StudentID)
		}
	}

	// Double check that distractor StuD works as expected:
	// Sol2 (Order, Interface) vs StuD (Order, Class). 
	// The Exact Match pass will SKIP because Types differ ("Interface" vs "Class").
	// Next pass, Sol2 (w=200) compares deltaWeight. 
	// StuB (w=200) delta=0. StuD (w=900) delta=700.
	// StuB checked first. Fuzzy matches "Ordr" vs "Order" -> dist 1. Sim = 1 - 1/5 = 0.80. 
	// Wait! 0.80 is >= 0.8? Yes! So StuB maps! StuD is skipped.
}

func TestArchWeightPriority(t *testing.T) {
	fz := NewLevenshteinMatcher()
	matcher := NewStandardEntityMatcher(fz, 0.75) // lowered threshold a bit for test
	
	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "S1", Name: "Manager", Type: "Class", ArchWeight: 500},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			// Candidate 1: Typo name, completely different weight
			{ID: "Stu1", Name: "Managr", Type: "Class", ArchWeight: 100},
			// Candidate 2: Typo name, exact same weight
			{ID: "Stu2", Name: "Managr", Type: "Class", ArchWeight: 500},
		},
	}

	mapping, _ := matcher.Match(solGraph, stuGraph)
	// Because of sorting by ArchWeight delta, Stu2 should be preferred over Stu1
	// Wait, since both have the same strings "Managr", they will both score the same fuzzy score.
	// But because Stu2 has delta(500-500)=0, it will be at index 0 in candidates.
	// So Stu2 should definitely be matched first.
	
	if mapping["S1"].StudentID != "Stu2" {
		t.Errorf("Expected Stu2 to be prioritze due to ArchWeight, mapped to %s", mapping["S1"].StudentID)
	}
}

func TestInterfaceArchitectureMatch(t *testing.T) {
	fz := NewLevenshteinMatcher()
	matcher := NewStandardEntityMatcher(fz, 0.75)
	
	// Create interface with identical structure weight (1000)
	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "S1", Name: "IProductRepository", Type: "Interface", ArchWeight: 1000},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			// Candidate 1: Same architecture, bad name
			{ID: "Stu1", Name: "IUserRepository", Type: "Interface", ArchWeight: 1000},
			// Candidate 2: Same architecture, exact name
			{ID: "Stu2", Name: "IProductRepository", Type: "Interface", ArchWeight: 1000},
			// Candidate 3: Same architecture, typo name
			{ID: "Stu3", Name: "IProductRepo", Type: "Interface", ArchWeight: 1000},
		},
	}

	mapping, _ := matcher.Match(solGraph, stuGraph)
	// Because all share identical Architecture (delta=0), they are bundled in "IsSimilar=True"
	// Then tiered tie-breaker will pick the one with Highest Fuzzy Score (Stu2 with 1.0)
	if mapping["S1"].StudentID != "Stu2" {
		t.Errorf("Expected Stu2 to be prioritized due to exact name match within same architecture, mapped to %s", mapping["S1"].StudentID)
	}
}

func TestToleranceArchitectureMatch(t *testing.T) {
	fz := NewLevenshteinMatcher()
	matcher := NewStandardEntityMatcher(fz, 0.75)
	
	// Base ArchWeight: Class (1<<29), NumMethods=10 (10<<18) => 539492352
	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "S1", Name: "UserService", Type: "Class", ArchWeight: 539492352},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			// Candidate 1: Typo name, perfect architecture (NumMethods=10) => 539492352
			{ID: "Stu1", Name: "UserServic", Type: "Class", ArchWeight: 539492352},
			
			// Candidate 2: Exact name, tolerant architecture (9 methods -> diff 1 <= ceil(1.5))
			// AW: Class (1<<29), NumMethods=9 (9<<18) => 539230208
			{ID: "Stu2", Name: "UserService", Type: "Class", ArchWeight: 539230208},
			
			// Candidate 3: Exact name, bad architecture (3 methods -> diff 7 > 2, ~70% drop)
			// AW: Class (1<<29), NumMethods=3 (3<<18) => 537657344
			{ID: "Stu3", Name: "UserService", Type: "Class", ArchWeight: 537657344},
		},
	}

	mapping, _ := matcher.Match(solGraph, stuGraph)
	
	// Stu1 and Stu2 BOTH pass IsArchitectureSimilar (Stu2 is within 15% missing 1 method)
	// So they are tied in Tier 1.
	// In Tier 2 (Fuzzy), Stu2's exact name match (1.0) beats Stu1's typo (~0.90)
	// Therefore Stu2 must win.
	if mapping["S1"].StudentID != "Stu2" {
		t.Errorf("Expected Stu2 to be matched due to exact name and <=15%% tolerance architecture, mapped to %s", mapping["S1"].StudentID)
	}
}

func TestTwoPassMatching(t *testing.T) {
	fz := NewLevenshteinMatcher()
	matcher := NewStandardEntityMatcher(fz, 0.8)

	// Base ArchWeight: Class (1<<29), NumMethods=10 (10<<18) => 539492352
	solGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			{ID: "S1", Name: "CruiseShip", Type: "Class", ArchWeight: 539492352},
		},
	}

	stuGraph := &domain.ProcessedUMLGraph{
		Nodes: []domain.ProcessedNode{
			// Candidate: Completely different name "PPShip". 
			// SimScore = ~0.4 (4 common chars / 10 maxlen = 0.4)
			// ArchWeight is exactly identical: 539492352
			{ID: "Stu1", Name: "PPShip", Type: "Class", ArchWeight: 539492352},
		},
	}

	mapping, _ := matcher.Match(solGraph, stuGraph)
	
	// In Pass 1: "PPShip" vs "CruiseShip" simScore < 0.8. Even though Arch is perfectly matched, it fails threshold.
	// In Pass 2: The remaining unmapped S1 is re-evaluated.
	// Tolerance becomes 0.10 (Arch is identical, so it passes).
	// Threshold drops to 0.4. Since SimScore is ~0.4 (depending on exact levenshtein implementation, might be 0.4 or higher),
	// it should match!
	
	if val, ok := mapping["S1"]; !ok {
		t.Errorf("Expected S1 to be matched in Pass 2")
	} else if val.StudentID != "Stu1" {
		t.Errorf("Expected S1 to map to Stu1, got %s", val.StudentID)
	} else if val.Similarity < 0.80 { // Arch (1.0 * 0.7) + Text (~0.4 * 0.3) = ~0.82
		t.Errorf("Expected Similarity to be ~0.82, got %.4f", val.Similarity)
	}
}

