// cmd/visualize/main.go
// Usage: go run ./cmd/visualize/main.go [--admin] <solution.drawio> <student.drawio> [output.html]
// Runs the full pipeline (Parse → Build → Validate → PreMatch → Match → Compare → Grade → Visualize)
// and exports a self-contained HTML report, then auto-opens it in the default browser.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"uml_compare/builder"
	"uml_compare/cmd/share"
	"uml_compare/comparator"
	"uml_compare/domain"
	"uml_compare/grader"
	"uml_compare/matcher"
	"uml_compare/parser"
	"uml_compare/prematcher"
	"uml_compare/visualizer"
)

func main() {
	if len(os.Args) <= 1 {
		share.PrintBanner("UML Visual Report Generator — Interactive Mode")
		fmt.Printf("   Tip: You can also drag & drop files here!\n\n")
		runInteractiveLoop()
		return
	}

	isAdmin, args := parseFlags(os.Args)

	if len(args) < 3 {
		fmt.Println("Usage: visualize_cli.exe [--admin] <solution.drawio> <student.drawio> [output.html]")
		os.Exit(1)
	}

	solutionPath := args[1]
	studentPath := args[2]
	outputPath := ""
	if len(args) >= 4 {
		outputPath = args[3]
	}

	if err := runComparison(solutionPath, studentPath, outputPath, isAdmin); err != nil {
		fmt.Printf("\n❌ Error: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags tách flag --admin khỏi danh sách args còn lại.
func parseFlags(rawArgs []string) (isAdmin bool, args []string) {
	for _, arg := range rawArgs {
		if arg == "--admin" {
			isAdmin = true
		} else {
			args = append(args, arg)
		}
	}
	return
}

// runComparison thực hiện toàn bộ pipeline và xuất HTML report.
func runComparison(solutionPath, studentPath, outputPath string, isAdmin bool) error {
	// ── 1. Parse ──────────────────────────────────────────────────────────
	solRaw, stuRaw, solType, stuType, err := parseBothFiles(solutionPath, studentPath)
	if err != nil {
		return err
	}

	// ── 2. Build ──────────────────────────────────────────────────────────
	b := builder.NewStandardModelBuilder()
	solGraph, err := b.Build(solRaw, solType)
	if err != nil {
		return fmt.Errorf("build solution: %w", err)
	}
	stuGraph, err := b.Build(stuRaw, stuType)
	if err != nil {
		return fmt.Errorf("build student: %w", err)
	}
	printLoadStatus(solGraph, stuGraph)

	// ── 3. Validate ───────────────────────────────────────────────────────
	allIssues := append(
		domain.ValidateGraph(solGraph, "Solution"),
		domain.ValidateGraph(stuGraph, "Student")...,
	)
	if hardErrors := domain.FilterErrors(allIssues); len(hardErrors) > 0 {
		fmt.Printf("❌ Integrity errors — pipeline halted:\n")
		for _, e := range hardErrors {
			fmt.Printf("   • %s\n", e.Error())
		}
		return fmt.Errorf("integrity check failed")
	}

	// ── 4. PreMatch ───────────────────────────────────────────────────────
	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()
	stuProc, _ := stdPM.Process(stuGraph)
	solForMatch, _ := solPM.ProcessSolution(solGraph)

	// ── 5. Match ──────────────────────────────────────────────────────────
	entityMatcher := matcher.NewStandardEntityMatcher(0.8)
	mapping, _ := entityMatcher.Match(solForMatch, stuProc)

	// ── 6. Compare ────────────────────────────────────────────────────────
	comp := comparator.NewStandardComparator()
	diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)

	// ── 7. Grade ──────────────────────────────────────────────────────────
	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{}
	gradeResult, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)
	fmt.Printf("   📈 Score: %.2f / %.2f (%.1f%%)\n\n", gradeResult.TotalScore, gradeResult.MaxScore, gradeResult.CorrectPercent)

	// ── 8. Visualize ──────────────────────────────────────────────────────
	return exportReports(gradeResult, studentPath, outputPath, isAdmin)
}

// ── Print / IO Layer ─────────────────────────────────────────────────────────

// parseBothFiles parse cả hai file drawio/solution, trả về raw bytes.
func parseBothFiles(solutionPath, studentPath string) (domain.RawModelData, domain.RawModelData, string, string, error) {
	p, err := parser.GetParser(solutionPath)
	if err != nil {
		return "", "", "", "", fmt.Errorf("parser factory: %w", err)
	}
	solRaw, solType, err := p.Parse(solutionPath)
	if err != nil {
		return "", "", "", "", fmt.Errorf("parse solution: %w", err)
	}
	stuRaw, stuType, err := p.Parse(studentPath)
	if err != nil {
		return "", "", "", "", fmt.Errorf("parse student: %w", err)
	}
	return solRaw, stuRaw, solType, stuType, nil
}

// printLoadStatus in số node/edge sau khi build thành công.
func printLoadStatus(solGraph, stuGraph *domain.UMLGraph) {
	fmt.Printf("   ✅ Loaded solution: %d nodes, %d edges\n", len(solGraph.Nodes), len(solGraph.Edges))
	fmt.Printf("   ✅ Loaded student:  %d nodes, %d edges\n", len(stuGraph.Nodes), len(stuGraph.Edges))
}

// exportReports xuất HTML report cho grader và student, rồi auto-open.
func exportReports(gradeResult *domain.GradeResult, studentPath, outputPath string, isAdmin bool) error {
	vis := visualizer.NewHTMLVisualizer()

	// Resolve output path
	if outputPath == "" {
		baseName := strings.TrimSuffix(filepath.Base(studentPath), filepath.Ext(studentPath))
		outputPath = fmt.Sprintf("report_%s.html", baseName)
	}
	if !strings.HasSuffix(strings.ToLower(outputPath), ".html") {
		outputPath += ".html"
	}

	// Full grader report
	if err := vis.ExportHTML(gradeResult, outputPath); err != nil {
		return fmt.Errorf("export grader report: %w", err)
	}
	fmt.Printf("✅ Grader report:  %s\n", outputPath)

	// Student feedback report
	stuBaseName := strings.TrimSuffix(filepath.Base(studentPath), filepath.Ext(studentPath))
	studentOutputPath := fmt.Sprintf("feedback_%s.html", stuBaseName)
	if err := vis.ExportStudentHTML(gradeResult, studentOutputPath); err != nil {
		return fmt.Errorf("export student report: %w", err)
	}
	fmt.Printf("✅ Student report: %s\n", studentOutputPath)

	// Auto-open
	targetToOpen := studentOutputPath
	if isAdmin {
		targetToOpen = outputPath
	}
	absPath, _ := filepath.Abs(targetToOpen)
	share.OpenFile(absPath)

	return nil
}
