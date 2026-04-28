package main

import (
	"fmt"
	"os"
	"uml_compare/domain"
	"uml_compare/similarity"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: similarity_tool <sample> <value1> [value2] ... [valueN]")
		os.Exit(1)
	}

	sample := os.Args[1]
	candidates := os.Args[2:]

	matcher, err := similarity.GetHybridMatcher(domain.DefaultConfig)
	if err != nil {
		fmt.Printf("Error initializing matcher: %v\n", err)
		os.Exit(1)
	}
	defer matcher.Close()

	fmt.Printf("Comparing with sample: '%s'\n", sample)
	fmt.Println("--------------------------------------------------")
	for _, cand := range candidates {
		score := matcher.Compare(sample, cand)
		fmt.Printf("Target: '%-20s' | Similarity: %6.2f%%\n", cand, score*100)
	}
}
