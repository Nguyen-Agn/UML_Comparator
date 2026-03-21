package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"uml_compare/builder"
	"uml_compare/parser"
	"uml_compare/prematcher"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/prematch/main.go <file.drawio>")
		fmt.Println("Example: go run cmd/prematch/main.go UMLs_testcase/problem1.drawio")
		os.Exit(1)
	}

	filePath := os.Args[1]

	fmt.Println("=== Prematcher Module Demo ===")
	fmt.Printf("Input File: %s\n", filePath)

	// 1. Initialize Pipeline
	fmt.Println("[1] Initializing Pipeline...")
	p := parser.NewDrawioParser()
	b := builder.NewStandardModelBuilder()
	pm := prematcher.NewStandardPreMatcher()

	// 2. Parse File
	fmt.Println("[2] Parsing .drawio file...")
	rawXML, err := p.Parse(filePath)
	if err != nil {
		log.Fatalf("Parser error: %v", err)
	}

	// 3. Build Graph
	fmt.Println("[3] Building UML Graph...")
	graph, err := b.Build(rawXML)
	if err != nil {
		log.Fatalf("Builder error: %v", err)
	}

	// 4. Process with Prematcher
	fmt.Println("[4] Running Prematcher (Bóc tách chi tiết & ArchWeight)...")
	processedGraph, err := pm.Process(graph)
	if err != nil {
		log.Fatalf("Error processing graph: %v", err)
	}

	// Pretty print JSON output
	outputBytes, err := json.MarshalIndent(processedGraph, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	fmt.Println("\nProcessed UML Graph Output:")
	fmt.Println(string(outputBytes))

	fmt.Println("\nArchWeight Analysis:")
	for _, n := range processedGraph.Nodes {
		fmt.Printf("- Node: %s (Type: %s, Shortcut: %d)\n", n.Name, n.Type, n.Shortcut)
		fmt.Printf("  -> ArchWeight: %d (Binary: %032b)\n", n.ArchWeight, n.ArchWeight)
		if len(n.Attributes) > 0 {
			fmt.Println("     Attributes:")
			for _, a := range n.Attributes {
				fmt.Printf("       • %s %s : %s [%s]\n", a.Scope, a.Name, a.Type, a.Kind)
			}
		}
		if len(n.Methods) > 0 {
			fmt.Println("     Methods:")
			for _, m := range n.Methods {
				fmt.Printf("       • %s %s() : %s [%s] (Type: %s)\n", m.Scope, m.Name, m.Output, m.Kind, m.Type)
			}
		}
	}
}
