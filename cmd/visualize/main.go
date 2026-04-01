// cmd/visualize/main.go
// Usage: go run ./cmd/visualize/main.go <solution.drawio> <student.drawio> [output.html]
// Runs the full pipeline (Parse → Build → Validate → PreMatch → Match → Compare → Grade → Visualize)
// and exports a self-contained HTML report, then auto-opens it in the default browser.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"uml_compare/builder"
	"uml_compare/comparator"
	"uml_compare/domain"
	"uml_compare/grader"
	"uml_compare/matcher"
	"uml_compare/parser"
	"uml_compare/prematcher"
	"uml_compare/visualizer"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run ./cmd/visualize/main.go <solution.drawio> <student.drawio> [output.html]")
		fmt.Println("  Example: go run ./cmd/visualize/main.go UMLs_testcase/problem1.drawio UMLs_testcase/problem_1.drawio")
		os.Exit(1)
	}

	solutionPath := os.Args[1]
	studentPath := os.Args[2]

	// Determine output path
	outputPath := ""
	if len(os.Args) >= 4 {
		outputPath = os.Args[3]
	} else {
		baseName := strings.TrimSuffix(filepath.Base(studentPath), filepath.Ext(studentPath))
		outputPath = fmt.Sprintf("report_%s.html", baseName)
	}

	fmt.Printf("📊 UML Visual Report Generator\n")
	fmt.Printf("   Solution: %s\n", solutionPath)
	fmt.Printf("   Student:  %s\n", studentPath)
	fmt.Printf("   Output:   %s\n\n", outputPath)

	// ── 1. Parse ─────────────────────────────────────────────────────────
	p := parser.NewDrawioParser()

	solRaw, err := p.Parse(solutionPath)
	if err != nil {
		fmt.Printf("❌ Parse solution error: %v\n", err)
		os.Exit(1)
	}

	stuRaw, err := p.Parse(studentPath)
	if err != nil {
		fmt.Printf("❌ Parse student error: %v\n", err)
		os.Exit(1)
	}

	// ── 2. Build ─────────────────────────────────────────────────────────
	b := builder.NewStandardModelBuilder()

	solGraph, err := b.Build(solRaw)
	if err != nil {
		fmt.Printf("❌ Build solution error: %v\n", err)
		os.Exit(1)
	}

	stuGraph, err := b.Build(stuRaw)
	if err != nil {
		fmt.Printf("❌ Build student error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("   ✅ Loaded solution: %d nodes, %d edges\n", len(solGraph.Nodes), len(solGraph.Edges))
	fmt.Printf("   ✅ Loaded student:  %d nodes, %d edges\n", len(stuGraph.Nodes), len(stuGraph.Edges))

	// ── 3. Validate ──────────────────────────────────────────────────────
	allIssues := append(
		domain.ValidateGraph(solGraph, "Solution"),
		domain.ValidateGraph(stuGraph, "Student")...,
	)
	hardErrors := domain.FilterErrors(allIssues)
	if len(hardErrors) > 0 {
		fmt.Printf("❌ Integrity errors — pipeline halted:\n")
		for _, e := range hardErrors {
			fmt.Printf("   • %s\n", e.Error())
		}
		os.Exit(1)
	}

	// ── 4. PreMatch ──────────────────────────────────────────────────────
	stdPM := prematcher.NewStandardPreMatcher()
	solPM := prematcher.NewUMLSolutionPreMatcher()

	stuProc, _ := stdPM.Process(stuGraph)
	solForMatch, _ := solPM.ProcessSolution(solGraph)

	// ── 5. Match ─────────────────────────────────────────────────────────
	fuzzy := matcher.NewLevenshteinMatcher()
	arch := matcher.NewStandardArchAnalyzer()
	entityMatcher := matcher.NewStandardEntityMatcher(fuzzy, arch, 0.8)
	mapping, _ := entityMatcher.Match(solForMatch, stuProc)

	// ── 6. Compare ───────────────────────────────────────────────────────
	ta := comparator.NewStandardTypeAnalyzer()
	mc := comparator.NewStandardMemberComparator(fuzzy, ta)
	ec := comparator.NewStandardEdgeComparator()
	comp := comparator.NewStandardComparator(fuzzy, ta, mc, ec)
	diffReport, _ := comp.Compare(solForMatch, stuProc, mapping)

	// ── 7. Grade ─────────────────────────────────────────────────────────
	gr := grader.NewStandardGrader()
	rules := &grader.GradingRules{}
	gradeResult, _ := gr.Grade(diffReport, solForMatch, stuProc, rules)

	fmt.Printf("   📈 Score: %.2f / %.2f (%.1f%%)\n\n", gradeResult.TotalScore, gradeResult.MaxScore, gradeResult.CorrectPercent)

	// ── 8. Visualize ─────────────────────────────────────────────────────
	vis := visualizer.NewHTMLVisualizer()
	if err := vis.ExportHTML(gradeResult, outputPath); err != nil {
		fmt.Printf("❌ Export error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Report exported: %s\n", outputPath)

	// ── 9. Auto-open in browser ──────────────────────────────────────────
	absPath, _ := filepath.Abs(outputPath)
	openBrowser(absPath)
}

// openBrowser opens the given file path in the default browser.
func openBrowser(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // linux
		cmd = exec.Command("xdg-open", path)
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("⚠️  Could not auto-open browser: %v\n", err)
		fmt.Printf("   Open manually: %s\n", path)
	}
}
