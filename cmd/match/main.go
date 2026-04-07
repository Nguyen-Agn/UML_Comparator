package main

import (
	"fmt"
	"log"
	"os"

	"uml_compare/builder"
	"uml_compare/matcher"
	"uml_compare/parser"
	"uml_compare/prematcher"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/match/main.go <solution.drawio> <student.drawio>")
		os.Exit(1)
	}

	solPath := os.Args[1]
	stuPath := os.Args[2]

	fmt.Println("=== Matcher Module Pipeline ===")
	fmt.Printf("Solution File: %s\n", solPath)
	fmt.Printf("Student File:  %s\n\n", stuPath)

	// 1. Initialize Interfaces
	fmt.Println("[1] Initializing Pipeline Interfaces...")
	fileParser, err := parser.GetParser(solPath)
	if err != nil {
		log.Fatalf("Failed to get parser: %v\n", err)
	}
	var modelBuilder builder.IModelBuilder = builder.NewStandardModelBuilder()
	var stdPreMatcher prematcher.IPreMatcher = prematcher.NewStandardPreMatcher()
	var solPreMatcher prematcher.IUMLSolutionPreMatcher = prematcher.NewUMLSolutionPreMatcher()

	fuzzy := matcher.NewLevenshteinMatcher()
	arch := matcher.NewStandardArchAnalyzer()
	var entityMatcher matcher.IEntityMatcher = matcher.NewStandardEntityMatcher(fuzzy, arch, 0.8)

	// 2. Process Solution
	fmt.Println("[2] Processing Solution Graph...")
	solRaw, err := fileParser.Parse(solPath)
	if err != nil {
		log.Fatalf("Failed to parse solution file: %v\n", err)
	}
	solGraph, err := modelBuilder.Build(solRaw)
	if err != nil {
		log.Fatalf("Failed to build solution graph: %v\n", err)
	}
	solProcessed, err := solPreMatcher.ProcessSolution(solGraph)
	if err != nil {
		log.Fatalf("Failed to process solution graph: %v\n", err)
	}

	// 3. Process Student
	fmt.Println("[3] Processing Student Graph...")
	stuRaw, err := fileParser.Parse(stuPath)
	if err != nil {
		log.Fatalf("Failed to parse student file: %v\n", err)
	}
	stuGraph, err := modelBuilder.Build(stuRaw)
	if err != nil {
		log.Fatalf("Failed to build student graph: %v\n", err)
	}
	stuProcessed, err := stdPreMatcher.Process(stuGraph)
	if err != nil {
		log.Fatalf("Failed to process student graph: %v\n", err)
	}

	// 4. Run Matcher
	fmt.Println("[4] Running Entity Matcher...")
	mapping, err := entityMatcher.Match(solProcessed, stuProcessed)
	if err != nil {
		log.Fatalf("Matcher error: %v\n", err)
	}

	// 5. Output Mapping Results
	fmt.Println("\n=== Mapping Results ===")
	solNames := make(map[string]string)
	for _, n := range solProcessed.Nodes {
		solNames[n.ID] = n.Name
	}
	stuNames := make(map[string]string)
	for _, n := range stuProcessed.Nodes {
		stuNames[n.ID] = n.Name
	}

	for solID, mappedNode := range mapping {
		solName := solNames[solID]
		fmt.Printf("  Solution '%s'  ==>  Student '%s'  (Similarity: %.2f)\n",
			solName,
			stuNames[mappedNode.StudentID],
			mappedNode.Similarity)
	}

	// Unmatched solution nodes
	for _, n := range solProcessed.Nodes {
		if _, ok := mapping[n.ID]; !ok {
			fmt.Printf("  Solution '%s'  ==>  [NO MATCH]\n", n.Name)
		}
	}

	fmt.Printf("\nMatched %d / %d solution nodes.\n", len(mapping), len(solProcessed.Nodes))
	fmt.Println("\nDone.")
}
