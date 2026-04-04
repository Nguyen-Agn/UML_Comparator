package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"uml_compare/builder"
	"uml_compare/comparator"
	"uml_compare/domain"
	"uml_compare/grader"
	"uml_compare/matcher"
	"uml_compare/parser"
	"uml_compare/prematcher"
	"uml_compare/report"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("📊 UML Batch Grader (Teacher Mode)\n")
		fmt.Printf("   Tip: Drag & drop the folder here for student submissions!\n\n")
		runBatchInteractiveLoop()
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Usage: lecture_cli_parallel.exe <solution.drawio> <student_dir> [report.csv]")
		os.Exit(1)
	}

	solutionPath := os.Args[1]
	studentDir := os.Args[2]
	outputPath := "batch_report.csv"
	if len(os.Args) >= 4 {
		outputPath = os.Args[3]
	}

	if err := runBatchGrading(solutionPath, studentDir, outputPath); err != nil {
		fmt.Printf("\n❌ Batch Error: %v\n", err)
		os.Exit(1)
	}
}

// runBatchGrading executes the grading pipeline for all files in a folder in parallel.
func runBatchGrading(solutionPath, studentDir, outputPath string) error {
	fmt.Printf("⏳ Loading solution from %s...\n", filepath.Base(solutionPath))
	solutionGraph := loadGraph(solutionPath)
	if solutionGraph == nil {
		return fmt.Errorf("failed to load solution graph")
	}

	// Integrity Check
	solErrs := domain.ValidateGraph(solutionGraph, "Solution")
	if len(domain.FilterErrors(solErrs)) > 0 {
		return fmt.Errorf("solution has integrity errors (check logs)")
	}

	// Prepare Toolchain (Immutable components shared across threads)
	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()
	solForMatch, _ := solPM.ProcessSolution(solutionGraph)
	
	fuzzy := matcher.NewLevenshteinMatcher()
	arch := matcher.NewStandardArchAnalyzer()
	entityMatcher := matcher.NewStandardEntityMatcher(fuzzy, arch, 0.8)
	ta := comparator.NewStandardTypeAnalyzer()
	mc := comparator.NewStandardMemberComparator(fuzzy, ta)
	ec := comparator.NewStandardEdgeComparator()
	comp := comparator.NewStandardComparator(fuzzy, ta, mc, ec)
	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{}

	batchResult := &report.BatchGradeResult{
		SolutionPath:   solutionPath,
		StudentResults: make(map[string]*domain.GradeResult),
	}

	// Scan directory
	entries, err := os.ReadDir(studentDir)
	if err != nil {
		return fmt.Errorf("cannot read directory: %w", err)
	}

	var studentFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".drawio") {
			studentFiles = append(studentFiles, e.Name())
		}
	}

	if len(studentFiles) == 0 {
		return fmt.Errorf("no .drawio files found in %s", studentDir)
	}

	fmt.Printf("🚀 Processing %d submissions using Parallel Pipeline...\n", len(studentFiles))
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	startTime := time.Now()

	for _, filename := range studentFiles {
		wg.Add(1)
		go func(fname string) {
			defer wg.Done()
			
			stuPath := filepath.Join(studentDir, fname)
			stuGraph := loadGraph(stuPath)
			if stuGraph == nil {
				fmt.Printf("  [!] %s: Failed to load\n", fname)
				return
			}

			stuProc, _ := stdPM.Process(stuGraph)
			mapping, _ := entityMatcher.Match(solForMatch, stuProc)
			diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)
			res, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)

			mu.Lock()
			batchResult.StudentResults[fname] = res
			mu.Unlock()
			fmt.Printf("  [✓] %s: %.1f%%\n", fname, res.CorrectPercent)
		}(filename)
	}

	wg.Wait()
	duration := time.Since(startTime)
	fmt.Printf("\n✨ Finished grading in %v.\n", duration)

	// Save CSV
	if !strings.HasSuffix(strings.ToLower(outputPath), ".csv") {
		outputPath += ".csv"
	}
	csvRep := report.NewCSVReporter(outputPath)
	if err := csvRep.GenerateReport(batchResult); err != nil {
		return fmt.Errorf("CSV export failed: %w", err)
	}

	fmt.Printf("📁 Report saved to: %s\n", outputPath)
	
	// Open CSV automatically
	absPath, _ := filepath.Abs(outputPath)
	openFile(absPath)

	return nil
}

// openFile uses the OS system call to open a file with its default application
func openFile(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	_ = cmd.Start()
}

func loadGraph(filePath string) *domain.UMLGraph {
	p := parser.NewDrawioParser()
	rawXML, err := p.Parse(filePath)
	if err != nil {
		return nil
	}

	b := builder.NewStandardModelBuilder()
	graph, err := b.Build(rawXML)
	if err != nil {
		return nil
	}
	return graph
}
