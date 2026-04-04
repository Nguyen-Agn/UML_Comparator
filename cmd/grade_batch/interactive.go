package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runBatchInteractiveLoop provides a terminal-based Q&A experience for teachers.
func runBatchInteractiveLoop() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("  UML Batch Grader — Lecture Edition (Parallel)")
		fmt.Println("--------------------------------------------------------")

		solPath := promptBatch(scanner, "[1/3] SOLUTION file path")
		if solPath == "" {
			fmt.Println("❌ Path cannot be empty.")
			continue
		}
		if _, err := os.Stat(solPath); os.IsNotExist(err) {
			fmt.Printf("❌ File not found: %s\n", solPath)
			continue
		}

		stuDir := promptBatch(scanner, "[2/3] STUDENT SUBMISSIONS directory")
		if stuDir == "" {
			fmt.Println("❌ Directory path cannot be empty.")
			continue
		}
		info, err := os.Stat(stuDir)
		if os.IsNotExist(err) || !info.IsDir() {
			fmt.Printf("❌ Directory not found or is a file: %s\n", stuDir)
			continue
		}

		outPath := promptBatch(scanner, "[3/3] Output .csv name (Enter = batch_report.csv)")
		if outPath == "" {
			outPath = "batch_report.csv"
		}

		// Quick scan to show count
		entries, _ := os.ReadDir(stuDir)
		count := 0
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".drawio") {
				count++
			}
		}

		fmt.Printf("\n📂 Found %d student files in %s.\n", count, filepath.Base(stuDir))
		fmt.Println("🚀 Initializing Parallel Grading Engine...")
		
		// Run the core batch logic (defined in main.go)
		err = runBatchGrading(solPath, stuDir, outPath)
		if err != nil {
			fmt.Printf("\n❌ Batch grading failed: %v\n", err)
		}

		fmt.Println("\n--------------------------------------------------------")
		ans := promptBatch(scanner, "Run another batch? (Y/N)")
		if strings.ToLower(ans) != "y" {
			fmt.Println("\nGoodbye!")
			break
		}
		fmt.Println()
	}
}

func promptBatch(scanner *bufio.Scanner, label string) string {
	fmt.Printf("  %s: ", label)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}
