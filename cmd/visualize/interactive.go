// cmd/visualize/interactive.go - Interactive terminal loop for visual report generation.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"uml_compare/cmd/share"
)

// runInteractiveLoop cung cấp giao diện nhập liệu terminal (thay thế VisualizeUML.bat).
func runInteractiveLoop() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("  UML Visual Report Generator — Interactive CLI")
		fmt.Println("--------------------------------------------------------")

		solPath := share.Prompt(scanner, "[1/3] SOLUTION file path")
		if solPath == "" {
			fmt.Println("❌ Path cannot be empty.")
			continue
		}
		if _, err := os.Stat(solPath); os.IsNotExist(err) {
			fmt.Printf("❌ File not found: %s\n", solPath)
			continue
		}

		stuPath := share.Prompt(scanner, "[2/3] STUDENT file path")
		if stuPath == "" {
			fmt.Println("❌ Path cannot be empty.")
			continue
		}
		if _, err := os.Stat(stuPath); os.IsNotExist(err) {
			fmt.Printf("❌ File not found: %s\n", stuPath)
			continue
		}

		outPath := share.Prompt(scanner, "[3/3] Output .html name (Enter = auto)")

		fmt.Println("\n🚀 Running analysis...")

		if err := runComparison(solPath, stuPath, outPath, false); err != nil {
			fmt.Printf("\n❌ Analysis failed: %v\n", err)
		}

		fmt.Println("\n--------------------------------------------------------")
		ans := share.Prompt(scanner, "Run another comparison? (Y/N)")
		if strings.ToLower(ans) != "y" {
			fmt.Println("\nGoodbye!")
			break
		}
		fmt.Println()
	}
}
