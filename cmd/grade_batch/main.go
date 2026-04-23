// cmd/grade_batch/main.go - Teacher batch grading CLI
// Usage: lecture_cli_parallel.exe <solution.drawio|.mmd> <student_dir> [report.csv]
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"uml_compare/cmd/share"
	"uml_compare/comparator"
	"uml_compare/domain"
	"uml_compare/grader"
	"uml_compare/matcher"
	"uml_compare/prematcher"
	"uml_compare/report"
)

func main() {
	if len(os.Args) <= 1 {
		share.PrintBanner("UML Batch Grader — Lecture Edition (Parallel)")
		fmt.Printf("   Tip: Drag & drop the folder here for student submissions!\n\n")
		runBatchInteractiveLoop()
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Usage: lecture_cli_parallel.exe <solution.drawio|.mmd> <student_dir> [report.csv]")
		os.Exit(1)
	}

	solutionPath := os.Args[1]
	studentDir := os.Args[2]
	outputPath := "batch_report.csv"
	if len(os.Args) >= 4 {
		outputPath = os.Args[3]
	}

	result, err := runBatchGrading(solutionPath, studentDir)
	if err != nil {
		fmt.Printf("\n❌ Batch Error: %v\n", err)
		os.Exit(1)
	}

	if err := saveBatchReport(result, outputPath); err != nil {
		fmt.Printf("\n❌ Report Error: %v\n", err)
		os.Exit(1)
	}
}

// batchRunResult chứa kết quả grading để truyền vào save/print layer.
type batchRunResult struct {
	BatchResult *report.BatchGradeResult
	Duration    time.Duration
	TotalFiles  int
}

// runBatchGrading thực hiện pipeline grading song song cho tất cả file trong thư mục.
// Trả về batchRunResult để caller tự quyết định cách lưu/hiển thị.
func runBatchGrading(solutionPath, studentDir string) (*batchRunResult, error) {
	fmt.Printf("⏳ Loading solution from %s...\n", filepath.Base(solutionPath))

	solutionGraph, err := share.LoadGraph(solutionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load solution: %w", err)
	}

	// Integrity Check
	if errs := domain.FilterErrors(domain.ValidateGraph(solutionGraph, "Solution")); len(errs) > 0 {
		return nil, fmt.Errorf("solution has integrity errors (check logs)")
	}

	// Build shared toolchain (immutable, safe for concurrent use)
	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()
	solForMatch, _ := solPM.ProcessSolution(solutionGraph)
	entityMatcher := matcher.NewStandardEntityMatcher(0.8)
	comp := comparator.NewStandardComparator()
	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{}

	// Scan student files
	entries, err := os.ReadDir(studentDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}
	var studentFiles []string
	for _, e := range entries {
		if !e.IsDir() && (strings.HasSuffix(e.Name(), ".drawio") || strings.HasSuffix(e.Name(), ".xml") || strings.HasSuffix(e.Name(), ".mmd") || strings.HasSuffix(e.Name(), ".mermaid")) {
			studentFiles = append(studentFiles, e.Name())
		}
	}
	if len(studentFiles) == 0 {
		return nil, fmt.Errorf("no UML files (.drawio, .mmd, .mermaid) found in %s", studentDir)
	}

	fmt.Printf("🚀 Processing %d submissions using Parallel Pipeline...\n", len(studentFiles))

	batchResult := &report.BatchGradeResult{
		SolutionPath:   solutionPath,
		StudentResults: make(map[string]*domain.GradeResult),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	startTime := time.Now()

	for _, filename := range studentFiles {
		wg.Add(1)
		go func(fname string) {
			defer wg.Done()

			stuGraph, err := share.LoadGraph(filepath.Join(studentDir, fname))
			if err != nil {
				fmt.Printf("  [!] %s: Failed to load — %v\n", fname, err)
				return
			}

			stuProc, _ := stdPM.Process(stuGraph)
			mapping, _ := entityMatcher.Match(solForMatch, stuProc)
			diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)
			res, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)

			mu.Lock()
			// NAME without extension (handle multiple dots)
			sname := strings.TrimSuffix(fname, filepath.Ext(fname))
			batchResult.StudentResults[sname] = res
			mu.Unlock()

			fmt.Printf("  [✓] %s: %.1f%%\n", fname, res.CorrectPercent)
		}(filename)
	}

	wg.Wait()

	return &batchRunResult{
		BatchResult: batchResult,
		Duration:    time.Since(startTime),
		TotalFiles:  len(studentFiles),
	}, nil
}

// saveBatchReport lưu kết quả ra file CSV và in tóm tắt.
func saveBatchReport(result *batchRunResult, outputPath string) error {
	fmt.Printf("\n✨ Finished grading in %v.\n", result.Duration)

	if !strings.HasSuffix(strings.ToLower(outputPath), ".csv") {
		outputPath += ".csv"
	}

	csvRep := report.NewCSVReporter(outputPath)
	if err := csvRep.GenerateReport(result.BatchResult); err != nil {
		return fmt.Errorf("CSV export failed: %w", err)
	}

	fmt.Printf("📁 Report saved to: %s\n", outputPath)

	absPath, _ := filepath.Abs(outputPath)
	share.OpenFile(absPath)

	return nil
}
