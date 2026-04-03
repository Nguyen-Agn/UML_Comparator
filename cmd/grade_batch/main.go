package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run ./cmd/grade_batch/main.go <solution.drawio> <student_dir>")
		fmt.Println("  Example: go run ./cmd/grade_batch/main.go UMLs_testcase/problem1.drawio UMLs_testcase/students")
		os.Exit(1)
	}

	solutionPath := os.Args[1]
	studentDir := os.Args[2]

	fmt.Println("Loading solution file...")
	solutionGraph := loadGraph(solutionPath)
	if solutionGraph == nil {
		fmt.Printf("Failed to load solution from %s\n", solutionPath)
		os.Exit(1)
	}

	// Integrity Check for Solution
	solErrs := domain.ValidateGraph(solutionGraph, "Solution")
	hardErrors := domain.FilterErrors(solErrs)
	if len(hardErrors) > 0 {
		fmt.Println("Integrity errors in solution file:")
		for _, e := range hardErrors {
			fmt.Printf(" - %s\n", e.Error())
		}
		os.Exit(1)
	}

	// Prepare Matcher & Comparator instances
	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()
	
	// Create solution processed graphs for matching and grading
	solForMatch, _ := solPM.ProcessSolution(solutionGraph)
	
	fuzzy := matcher.NewLevenshteinMatcher()
	arch := matcher.NewStandardArchAnalyzer()
	entityMatcher := matcher.NewStandardEntityMatcher(fuzzy, arch, 0.8)
	ta := comparator.NewStandardTypeAnalyzer()
	mc := comparator.NewStandardMemberComparator(fuzzy, ta)
	ec := comparator.NewStandardEdgeComparator()
	comp := comparator.NewStandardComparator(fuzzy, ta, mc, ec)
	
	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{} // Use default rules
	
	batchResult := &report.BatchGradeResult{
		SolutionPath:   solutionPath,
		StudentResults: make(map[string]*domain.GradeResult),
	}

	// Process directory
	fmt.Printf("Scanning directory %s for .drawio files...\n", studentDir)
	entries, err := os.ReadDir(studentDir)
	if err != nil {
		fmt.Printf("Failed to read directory: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".drawio") {
			continue
		}

		studentPath := filepath.Join(studentDir, entry.Name())
		fmt.Printf("Processing %s... ", entry.Name())

		studentGraph := loadGraph(studentPath)
		if studentGraph == nil {
			fmt.Println("❌ Load failed")
			continue
		}
		
		stuErrs := domain.ValidateGraph(studentGraph, "Student")
		if len(domain.FilterErrors(stuErrs)) > 0 {
			fmt.Println("❌ Validation failed")
			// Depending on requirements, we might skip or still grade with warnings
			// For batch grading, we'll continue and maybe it scores poorly.
		}

		stuProc, _ := stdPM.Process(studentGraph)
		mapping, _ := entityMatcher.Match(solForMatch, stuProc)
		diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)
		gradeResult, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)

		batchResult.StudentResults[entry.Name()] = gradeResult
		fmt.Println("✅ Done")
	}

	// Generate CSV Report
	csvRep := report.NewCSVReporter("batch_report.csv")
	err = csvRep.GenerateReport(batchResult)
	if err != nil {
		fmt.Printf("CSV Reporter error: %v\n", err)
	} else {
		fmt.Println("Ghi file batch_report.csv thành công!")
	}
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
