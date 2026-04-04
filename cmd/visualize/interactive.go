package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// runInteractiveLoop provides a terminal-based Q&A experience (replaces VisualizeUML.bat)
func runInteractiveLoop() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("  UML Visual Report Generator — Interactive CLI")
		fmt.Println("--------------------------------------------------------")

		solPath := prompt(scanner, "[1/3] SOLUTION file path")
		if solPath == "" {
			fmt.Println("❌ Path cannot be empty.")
			continue
		}
		if _, err := os.Stat(solPath); os.IsNotExist(err) {
			fmt.Printf("❌ File not found: %s\n", solPath)
			continue
		}

		stuPath := prompt(scanner, "[2/3] STUDENT file path")
		if stuPath == "" {
			fmt.Println("❌ Path cannot be empty.")
			continue
		}
		if _, err := os.Stat(stuPath); os.IsNotExist(err) {
			fmt.Printf("❌ File not found: %s\n", stuPath)
			continue
		}

		outPath := prompt(scanner, "[3/3] Output .html name (Enter = auto)")

		fmt.Println("\n🚀 Running analysis...")
		
		// Run the core comparison logic (defined in main.go)
		err := runComparison(solPath, stuPath, outPath, false)
		if err != nil {
			fmt.Printf("\n❌ Analysis failed: %v\n", err)
		}

		fmt.Println("\n--------------------------------------------------------")
		ans := prompt(scanner, "Run another comparison? (Y/N)")
		if strings.ToLower(ans) != "y" {
			fmt.Println("\nGoodbye!")
			break
		}
		fmt.Println()
	}
}

func prompt(scanner *bufio.Scanner, label string) string {
	fmt.Printf("  %s: ", label)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}
