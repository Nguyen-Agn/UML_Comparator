package main

import (
	"fmt"
	"log"
	"os"

	"uml_compare/builder"
	"uml_compare/matcher"
	"uml_compare/parser"
	"uml_compare/prematcher"
	"uml_compare/comparator"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/match/main.go <solution.drawio> <student.drawio>")
		os.Exit(1)
	}

	solPath := os.Args[1]
	stuPath := os.Args[2]

	fmt.Println("=== Matcher Module End-to-End Pipeline ===")
	fmt.Printf("Solution File: %s\n", solPath)
	fmt.Printf("Student File:  %s\n\n", stuPath)

	// 1. Initialize Interfaces
	fmt.Println("[1] Initializing Pipeline Interfaces...")
	var fileParser parser.IFileParser = parser.NewDrawioParser()
	
	var modelBuilder builder.IModelBuilder = builder.NewStandardModelBuilder()
	var preMatcher prematcher.IPreMatcher = prematcher.NewStandardPreMatcher()
	
	fuzzy := matcher.NewLevenshteinMatcher()
	var entityMatcher matcher.IEntityMatcher = matcher.NewStandardEntityMatcher(fuzzy, 0.8) // Threshold 80%

	// 2. Run Pipeline for Solution
	fmt.Println("[2] Processing Solution Graph...")
	solRaw, err := fileParser.Parse(solPath)
	if err != nil {
		log.Fatalf("Failed to parse solution file: %v\n", err)
	}
	solGraph, err := modelBuilder.Build(solRaw)
	if err != nil {
		log.Fatalf("Failed to build solution graph: %v\n", err)
	}
	solProcessed, err := preMatcher.Process(solGraph)
	if err != nil {
		log.Fatalf("Failed to calculate solution ArchWeights: %v\n", err)
	}

	// 3. Run Pipeline for Student
	fmt.Println("[3] Processing Student Graph...")
	stuRaw, err := fileParser.Parse(stuPath)
	if err != nil {
		log.Fatalf("Failed to parse student file: %v\n", err)
	}
	stuGraph, err := modelBuilder.Build(stuRaw)
	if err != nil {
		log.Fatalf("Failed to build student graph: %v\n", err)
	}
	stuProcessed, err := preMatcher.Process(stuGraph)
	if err != nil {
		log.Fatalf("Failed to calculate student ArchWeights: %v\n", err)
	}

	mapping, err := entityMatcher.Match(solProcessed, stuProcessed)
	if err != nil {
		log.Fatalf("Matcher error: %v\n", err)
	}

	// 5. Run Comparator
	fmt.Println("[5] Running Detailed Comparator (TypeMap -> Attributes -> Methods -> Edges)...")
	var comp comparator.IComparator = comparator.NewStandardComparator(fuzzy)
	diffReport, err := comp.Compare(solProcessed, stuProcessed, mapping)
	if err != nil {
		log.Fatalf("Comparator error: %v\n", err)
	}

	// 5. Output
	fmt.Println("\n=== Final Mapping Results ===")
	// Create lookup maps for names
	solNames := make(map[string]string)
	for _, n := range solProcessed.Nodes {
		solNames[n.ID] = n.Name
	}
	stuNames := make(map[string]string)
	for _, n := range stuProcessed.Nodes {
		stuNames[n.ID] = n.Name
	}

	for solID, mappedNode := range mapping {
		fmt.Printf("Solution [%s] '%s'  ==>  Student [%s] '%s' (Similarity: %.2f)\n", 
			solID, solNames[solID], 
			mappedNode.StudentID, stuNames[mappedNode.StudentID],
			mappedNode.Similarity)
	}

	fmt.Println("\n=== Detailed Difference Report ===")
	if len(diffReport.MissedClass) > 0 {
		fmt.Println("[Missed Classes]")
		for _, m := range diffReport.MissedClass { fmt.Printf(" - %s\n", m) }
	}
	if len(diffReport.MissingNodes) > 0 {
		fmt.Println("[Missed Nodes]")
		for _, m := range diffReport.MissingNodes { fmt.Printf(" - %s\n", m) }
	}
	if len(diffReport.MissingEdges) > 0 {
		fmt.Println("[Missed Edges]")
		for _, m := range diffReport.MissingEdges { fmt.Printf(" - %s\n", m) }
	}
	if len(diffReport.MissingMembers) > 0 {
		fmt.Println("[Missed Members]")
		for _, m := range diffReport.MissingMembers { fmt.Printf(" - %s\n", m) }
	}
	if len(diffReport.AttributeErrors) > 0 {
		fmt.Println("[Attribute Mismatches]")
		for _, m := range diffReport.AttributeErrors { fmt.Printf(" - %s\n", m) }
	}
	if len(diffReport.MethodErrors) > 0 {
		fmt.Println("[Method Mismatches]")
		for _, m := range diffReport.MethodErrors { fmt.Printf(" - %s\n", m) }
	}
	if len(diffReport.NodeEdgeErrors) > 0 {
		fmt.Println("[Relationship/Node Errors]")
		for _, m := range diffReport.NodeEdgeErrors { fmt.Printf(" - %s\n", m) }
	}

	if len(diffReport.MissedClass) == 0 && len(diffReport.MissingNodes) == 0 && len(diffReport.MissingEdges) == 0 && len(diffReport.MissingMembers) == 0 &&
	   len(diffReport.AttributeErrors) == 0 && len(diffReport.MethodErrors) == 0 && len(diffReport.NodeEdgeErrors) == 0 {
		fmt.Println("No structural differences found. Perfect match!")
	}
	
	fmt.Println("\nDone.")
}
