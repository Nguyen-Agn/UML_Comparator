package main

import (
	"encoding/json"
	"fmt"
	"log"
	
	"uml_compare/domain"
	"uml_compare/prematcher"
)

func main() {
	fmt.Println("=== Prematcher Module Demo ===")

	// Setup dummy UMLGraph from Builder
	graph := &domain.UMLGraph{
		ID: "DemoGraph",
		Nodes: []domain.UMLNode{
			{
				ID:   "Node_A",
				Name: "UserService",
				Type: "Class",
				Attributes: []string{
					"- users : List<User>",
					"+ static  Instance : UserService",
				},
				Methods: []string{
					"+ getUser(id: int) : User",
					"- saveUser(u: User) : boolean",
				},
			},
			{
				ID:   "Node_B",
				Name: "IUserService",
				Type: "Interface",
				Attributes: []string{},
				Methods: []string{
					"+ getUser(id: int) : User",
				},
			},
		},
		Edges: []domain.UMLEdge{
			{
				SourceID:     "Node_A",
				TargetID:     "Node_B",
				RelationType: "Realization", // UserService implements IUserService
			},
		},
	}

	pm := prematcher.NewStandardPreMatcher()
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
		fmt.Printf("- Node: %s (Type: %s)\n", n.Name, n.Type)
		fmt.Printf("  -> ArchWeight: %d (Binary: %032b)\n", n.ArchWeight, n.ArchWeight)
	}
}
