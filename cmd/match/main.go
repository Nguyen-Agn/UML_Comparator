// cmd/match/main.go - Thử nghiệm matcher: go run ./cmd/match/main.go <solution.drawio> <student.drawio>
package main

import (
	"fmt"
	"log"
	"os"
	"uml_compare/cmd/share"
	"uml_compare/domain"
	"uml_compare/matcher"
	"uml_compare/prematcher"
)

// matchResult chứa kết quả của pipeline match để truyền vào print layer.
type matchResult struct {
	Mapping      domain.MappingTable
	SolProcessed *domain.SolutionProcessedUMLGraph
	StuProcessed *domain.ProcessedUMLGraph
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/match/main.go <solution.drawio> <student.drawio>")
		os.Exit(1)
	}

	solPath := os.Args[1]
	stuPath := os.Args[2]

	share.PrintBanner("Matcher Module Pipeline")
	fmt.Printf("Solution File: %s\n", solPath)
	fmt.Printf("Student File:  %s\n\n", stuPath)

	result, err := run(solPath, stuPath)
	if err != nil {
		log.Fatalf("❌ Pipeline error: %v\n", err)
	}

	printMatchResult(result)
}

// run thực hiện toàn bộ pipeline Parse → Build → PreMatch → Match.
// Trả về mapping table và cả hai processed graph để in kết quả.
func run(solPath, stuPath string) (*matchResult, error) {
	// 1. Load graphs
	solGraph, err := share.LoadGraph(solPath)
	if err != nil {
		return nil, fmt.Errorf("load solution: %w", err)
	}
	stuGraph, err := share.LoadGraph(stuPath)
	if err != nil {
		return nil, fmt.Errorf("load student: %w", err)
	}

	// 2. PreMatch
	stdPreMatcher := prematcher.NewStandardPreMatcher()
	solPreMatcher := prematcher.NewUMLSolutionPreMatcher()

	solProcessed, err := solPreMatcher.ProcessSolution(solGraph)
	if err != nil {
		return nil, fmt.Errorf("process solution: %w", err)
	}
	stuProcessed, err := stdPreMatcher.Process(stuGraph)
	if err != nil {
		return nil, fmt.Errorf("process student: %w", err)
	}

	// 3. Match
	entityMatcher := matcher.NewStandardEntityMatcher(0.8)
	mapping, err := entityMatcher.Match(solProcessed, stuProcessed)
	if err != nil {
		return nil, fmt.Errorf("entity match: %w", err)
	}

	return &matchResult{
		Mapping:      mapping,
		SolProcessed: solProcessed,
		StuProcessed: stuProcessed,
	}, nil
}

// ── Print Layer ───────────────────────────────────────────────────────────────

// printMatchResult in kết quả mapping solution ↔ student.
func printMatchResult(r *matchResult) {
	// Build name lookup maps
	solNames := make(map[string]string, len(r.SolProcessed.Nodes))
	for _, n := range r.SolProcessed.Nodes {
		solNames[n.ID] = n.Name
	}
	stuNames := make(map[string]string, len(r.StuProcessed.Nodes))
	for _, n := range r.StuProcessed.Nodes {
		stuNames[n.ID] = n.Name
	}

	fmt.Println("\n=== Mapping Results ===")
	for solID, mappedNode := range r.Mapping {
		fmt.Printf("  Solution '%s'  ==>  Student '%s'  (Similarity: %.2f)\n",
			solNames[solID],
			stuNames[mappedNode.StudentID],
			mappedNode.Similarity)
	}

	// Unmatched solution nodes
	for _, n := range r.SolProcessed.Nodes {
		if _, ok := r.Mapping[n.ID]; !ok {
			fmt.Printf("  Solution '%s'  ==>  [NO MATCH]\n", n.Name)
		}
	}

	fmt.Printf("\nMatched %d / %d solution nodes.\n", len(r.Mapping), len(r.SolProcessed.Nodes))
	fmt.Println("\nDone.")
}
