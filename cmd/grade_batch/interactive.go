// cmd/grade_batch/interactive.go - Interactive terminal loop for batch grading.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"uml_compare/cmd/share"
)

// runBatchInteractiveLoop cung cấp giao diện nhập liệu terminal cho giáo viên.
func runBatchInteractiveLoop() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("--------------------------------------------------------")
		fmt.Println("  UML Batch Grader — Lecture Edition (Parallel)")
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

		stuDir := share.Prompt(scanner, "[2/3] STUDENT SUBMISSIONS directory")
		if stuDir == "" {
			fmt.Println("❌ Directory path cannot be empty.")
			continue
		}
		info, err := os.Stat(stuDir)
		if os.IsNotExist(err) || !info.IsDir() {
			fmt.Printf("❌ Directory not found or is a file: %s\n", stuDir)
			continue
		}

		outPath := share.Prompt(scanner, "[3/3] Output .csv name (Enter = batch_report.csv)")
		if outPath == "" {
			outPath = "batch_report.csv"
		}

		// Quick scan to show count before starting
		entries, _ := os.ReadDir(stuDir)
		count := 0
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".drawio") {
				count++
			}
		}

		fmt.Printf("\n📂 Found %d student files in %s.\n", count, filepath.Base(stuDir))
		fmt.Println("🚀 Initializing Parallel Grading Engine...")

		result, err := runBatchGrading(solPath, stuDir)
		if err != nil {
			fmt.Printf("\n❌ Batch grading failed: %v\n", err)
		} else if err := saveBatchReport(result, outPath); err != nil {
			fmt.Printf("\n❌ Save failed: %v\n", err)
		}

		fmt.Println("\n--------------------------------------------------------")
		ans := share.Prompt(scanner, "Run another batch? (Y/N)")
		if strings.ToLower(ans) != "y" {
			fmt.Println("\nGoodbye!")
			break
		}
		fmt.Println()
	}
}
