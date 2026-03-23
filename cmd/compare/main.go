// cmd/compare/main.go
// Chạy: go run ./cmd/compare/main.go <solution.drawio> <student.drawio>
// So sánh UML mẫu vs bài sinh viên, hiển thị trực quan side-by-side
package main

import (
	"fmt"
	"os"
	"strings"
	"uml_compare/builder"
	"uml_compare/comparator"
	"uml_compare/domain"
	"uml_compare/matcher"
	"uml_compare/parser"
	"uml_compare/prematcher"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run ./cmd/compare/main.go <solution.drawio> <student.drawio>")
		fmt.Println("  Example: go run ./cmd/compare/main.go UMLs_testcase/problem1.drawio UMLs_testcase/problem_1.drawio")
		os.Exit(1)
	}

	solutionPath := os.Args[1]
	studentPath := os.Args[2]

	fmt.Printf("%s╔══════════════════════════════════════════════════════════╗%s\n", Blue, Reset)
	fmt.Printf("%s║          UML Compare — Data Flow Integrity Check         ║%s\n", Blue+Bold, Reset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════╝%s\n", Blue, Reset)

	solutionGraph := loadGraph(solutionPath, "📘 Solution")
	studentGraph := loadGraph(studentPath, "📄 Student")

	// ── Integrity Validation Gate ─────────────────────────────────
	fmt.Printf("\n%s── [GATE] Data Integrity Check ──────────────────────────────%s\n", Cyan+Bold, Reset)
	solErrs := domain.ValidateGraph(solutionGraph, "Solution")
	stuErrs := domain.ValidateGraph(studentGraph, "Student")
	allIssues := append(solErrs, stuErrs...)

	hardErrors := domain.FilterErrors(allIssues)
	rawWarnings := domain.FilterWarns(allIssues)
	warnings := []domain.IntegrityError{}
	for _, w := range rawWarnings {
		// Suppress warnings for pure shortcuts like "+ getters / setters"
		lowerMsg := strings.ToLower(w.Message)
		if w.Code == "INCOMPLETE_ATTRIBUTE" && (strings.Contains(lowerMsg, "getter") || strings.Contains(lowerMsg, "setter")) {
			continue
		}
		warnings = append(warnings, w)
	}

	if len(warnings) > 0 {
		fmt.Printf("%s⚠️  UML Quality Warnings (%d) — comparison continues:%s\n", Yellow+Bold, len(warnings), Reset)
		for _, w := range warnings {
			fmt.Printf("   • %s\n", w.Error())
		}
	}
	if len(hardErrors) > 0 {
		fmt.Printf("%s❌ INTEGRITY ERRORS — Pipeline halted:%s\n", Red+Bold, Reset)
		for _, e := range hardErrors {
			fmt.Printf("   • %s\n", e.Error())
		}
		os.Exit(1)
	}
	if len(warnings) == 0 {
		fmt.Printf("%s✅ Both graphs pass — proceeding to comparison%s\n", Green, Reset)
	} else {
		fmt.Printf("%s⚠️  Continuing with warnings — results may be partially inaccurate%s\n", Yellow, Reset)
	}

	// ── AI Matcher Integration ───────────────────────────────────
	preMatcher := prematcher.NewStandardPreMatcher()
	solProc, _ := preMatcher.Process(solutionGraph)
	stuProc, _ := preMatcher.Process(studentGraph)

	fuzzy := matcher.NewLevenshteinMatcher()
	arch := matcher.NewStandardArchAnalyzer()
	entityMatcher := matcher.NewStandardEntityMatcher(fuzzy, arch, 0.8)
	mapping, _ := entityMatcher.Match(solProc, stuProc)

	// ── Advanced Comparator ──────────────────────────────────────
	ta := comparator.NewStandardTypeAnalyzer()
	mc := comparator.NewStandardMemberComparator(fuzzy, ta)
	ec := comparator.NewStandardEdgeComparator()
	comp := comparator.NewStandardComparator(fuzzy, ta, mc, ec)
	diffReport, _ := comp.Compare(solProc, stuProc, mapping)

	// ── Side-by-side Node Comparison ─────────────────────────────
	fmt.Printf("\n%s── [COMPARE] Nodes Side-by-Side ────────────────────────────%s\n", Cyan+Bold, Reset)
	printSideBySideNodes(solProc, stuProc, mapping)

	// ── Edge Comparison ──────────────────────────────────────────
	fmt.Printf("\n%s── [COMPARE] Edges (Relations) ──────────────────────────────%s\n", Cyan+Bold, Reset)
	printEdgeComparison(solProc, stuProc, mapping)

	// ── Detailed Report ──────────────────────────────────────────
	fmt.Printf("\n%s── [REPORT] Detailed Diff ─────────────────────────────────%s\n", Cyan+Bold, Reset)
	printDiffReport(diffReport)

	// ── Summary ──────────────────────────────────────────────────
	fmt.Printf("\n%s── [SUMMARY] Quick Stats ────────────────────────────────────%s\n", Cyan+Bold, Reset)
	printSummary(solProc, stuProc, mapping)
}

func loadGraph(filePath, label string) *domain.UMLGraph {
	fmt.Printf("\n%s: %s\n", label, filePath)
	p := parser.NewDrawioParser()
	rawXML, err := p.Parse(filePath)
	if err != nil {
		fmt.Printf("  ❌ Parser error: %v\n", err)
		os.Exit(1)
	}

	b := builder.NewStandardModelBuilder()
	graph, err := b.Build(rawXML)
	if err != nil {
		fmt.Printf("  ❌ Builder error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  %s✅ Loaded: %d nodes, %d edges%s\n", Green, len(graph.Nodes), len(graph.Edges), Reset)
	return graph
}

func printSideBySideNodes(sol, stu *domain.ProcessedUMLGraph, mapping domain.MappingTable) {
	const col = 60
	header := fmt.Sprintf("%s  %-*s│  %-*s%s", Bold+Blue, col, "SOLUTION (đáp án)", col, "STUDENT (bài nộp)", Reset)
	fmt.Println(strings.Repeat("─", col*2+4))
	fmt.Println(header)
	fmt.Println(strings.Repeat("─", col*2+4))

	mappedStu := make(map[string]bool)

	for i := range sol.Nodes {
		solNode := &sol.Nodes[i]
		solPart := fmt.Sprintf("[%s] %s (%dA/%dM)", solNode.Type[:1], solNode.Name, len(solNode.Attributes), len(solNode.Methods))

		var stuNode *domain.ProcessedNode
		stuPart := ""
		matchMark := "✗ "

		if mapped, ok := mapping[solNode.ID]; ok {
			for j := range stu.Nodes {
				if stu.Nodes[j].ID == mapped.StudentID {
					stuNode = &stu.Nodes[j]
					mappedStu[stuNode.ID] = true
					break
				}
			}
			if stuNode != nil {
				stuPart = fmt.Sprintf("[%s] %s (%dA/%dM)", stuNode.Type[:1], stuNode.Name, len(stuNode.Attributes), len(stuNode.Methods))
				if mapped.Similarity == 1.0 {
					matchMark = "✔ "
				} else {
					matchMark = "≈ "
				}
			}
		}

		color := Reset
		switch matchMark {
		case "✔ ":
			color = Green
		case "≈ ":
			color = Yellow
		default:
			color = Red
		}

		fmt.Printf("%-*s│ %s%s%-*s%s\n", col, "  "+solPart, color, matchMark, col, stuPart, Reset)

		// Print mismatched members side-by-side
		if stuNode != nil {
			solAttrs := []string{}
			for _, a := range solNode.Attributes {
				solAttrs = append(solAttrs, fmt.Sprintf("%s %s %s", a.Scope, a.Type, a.Name))
			}
			stuAttrs := []string{}
			for _, a := range stuNode.Attributes {
				stuAttrs = append(stuAttrs, fmt.Sprintf("%s %s %s", a.Scope, a.Type, a.Name))
			}
			printAllMembers("Attr", solAttrs, stuAttrs, col)

			solMeths := []string{}
			for _, m := range solNode.Methods {
				if m.Type == "getter" || m.Type == "setter" {
					continue
				}

				params := []string{}
				for _, p := range m.Inputs {
					params = append(params, p.Type)
				}
				solMeths = append(solMeths, fmt.Sprintf("%s %s(%s): %s", m.Scope, m.Name, strings.Join(params, ", "), m.Output))
			}
			stuMeths := []string{}
			for _, m := range stuNode.Methods {
				if m.Type == "getter" || m.Type == "setter" {
					continue
				}

				params := []string{}
				for _, p := range m.Inputs {
					params = append(params, p.Type)
				}
				stuMeths = append(stuMeths, fmt.Sprintf("%s %s(%s): %s", m.Scope, m.Name, strings.Join(params, ", "), m.Output))
			}
			printAllMembers("Meth", solMeths, stuMeths, col)
		}
	}

	// Print remaining student nodes
	for i := range stu.Nodes {
		stuNode := &stu.Nodes[i]
		if !mappedStu[stuNode.ID] {
			stuPart := fmt.Sprintf("[%s] %s (%dA/%dM)", stuNode.Type[:1], stuNode.Name, len(stuNode.Attributes), len(stuNode.Methods))
			fmt.Printf("%-*s│ %s%-*s%s\n", col, "  ", Red+"✗ ", col, stuPart, Reset)
		}
	}

	fmt.Println(strings.Repeat("─", col*2+4))
	fmt.Printf("  %sLegend:%s [C]=Class [I]=Interface [A]=Actor  (A=attrs M=methods)\n", Bold, Reset)
	fmt.Printf("          %s✔%s exact/perfect match  %s≈%s fuzzy/arch match  %s✗%s missing in one side\n", Green, Reset, Yellow, Reset, Red, Reset)
}

func printEdgeComparison(sol, stu *domain.ProcessedUMLGraph, mapping domain.MappingTable) {
	// Create lookup for node names
	solNames := make(map[string]string)
	for _, n := range sol.Nodes {
		solNames[n.ID] = n.Name
	}
	stuNames := make(map[string]string)
	for _, n := range stu.Nodes {
		stuNames[n.ID] = n.Name
	}

	fmt.Println("  Solution edges (mapped to student names if available):")
	for _, se := range sol.Edges {
		status := "  "
		mappedSrc := ""
		mappedTgt := ""

		// Find if this edge exists in student graph via mapping
		if mSrc, ok1 := mapping[se.SourceID]; ok1 {
			if mTgt, ok2 := mapping[se.TargetID]; ok2 {
				mappedSrc = mSrc.StudentID
				mappedTgt = mTgt.StudentID

				for _, ste := range stu.Edges {
					if ste.SourceID == mappedSrc && ste.TargetID == mappedTgt && ste.RelationType == se.RelationType {
						status = "✔ "
						break
					}
				}
			}
		}

		// Display using Student names if matched, otherwise Solution names
		srcName := solNames[se.SourceID]
		tgtName := solNames[se.TargetID]
		color := Reset
		if status == "✔ " {
			color = Green
		}
		fmt.Printf("    %s%s%s -[%s]-> %s%s\n", color, status, srcName, se.RelationType, tgtName, Reset)
	}

	fmt.Println("  Extra/different edges in Student:")
	for _, ste := range stu.Edges {
		found := false
		// reverse mapping check
		for solID, m := range mapping {
			if m.StudentID == ste.SourceID {
				// check target
				for solTgtID, m2 := range mapping {
					if m2.StudentID == ste.TargetID {
						// check if this specific relation exists in solution
						for _, se := range sol.Edges {
							if se.SourceID == solID && se.TargetID == solTgtID && se.RelationType == ste.RelationType {
								found = true
								break
							}
						}
					}
					if found {
						break
					}
				}
			}
			if found {
				break
			}
		}

		if !found {
			srcName := stuNames[ste.SourceID]
			tgtName := stuNames[ste.TargetID]
			fmt.Printf("    %s✗ %s -[%s]-> %s (not in solution)%s\n", Red, srcName, ste.RelationType, tgtName, Reset)
		}
	}
}

func printSummary(sol, stu *domain.ProcessedUMLGraph, mapping domain.MappingTable) {
	nodeHit := len(mapping)
	edgeHit := 0
	for _, se := range sol.Edges {
		if mSrc, ok1 := mapping[se.SourceID]; ok1 {
			if mTgt, ok2 := mapping[se.TargetID]; ok2 {
				for _, ste := range stu.Edges {
					if ste.SourceID == mSrc.StudentID && ste.TargetID == mTgt.StudentID && ste.RelationType == se.RelationType {
						edgeHit++
						break
					}
				}
			}
		}
	}

	nodePct := 100.0
	if len(sol.Nodes) > 0 {
		nodePct = float64(nodeHit) / float64(len(sol.Nodes)) * 100
	}
	edgePct := 100.0
	if len(sol.Edges) > 0 {
		edgePct = float64(edgeHit) / float64(len(sol.Edges)) * 100
	}

	overall := (nodePct + edgePct) / 2
	color := Cyan
	if overall >= 90 {
		color = Green + Bold
	} else if overall >= 60 {
		color = Yellow
	} else {
		color = Red
	}

	fmt.Printf("  Stats: Nodes %d/%d (%.0f%%), Edges %d/%d (%.0f%%) → %sOverall: %.1f%%%s\n",
		nodeHit, len(sol.Nodes), nodePct, edgeHit, len(sol.Edges), edgePct, color, overall, Reset)
}

func printDiffReport(report *domain.DiffReport) {
	printDetailSection("🚨 [MISSING DETAILS]", report.MissingDetail, Red+Bold)
	printDetailSection("⚠️ [WRONG/MISMATCHED DETAILS]", report.WrongDetail, Yellow+Bold)
	printDetailSection("➕ [EXTRA DETAILS]", report.ExtraDetail, Cyan+Bold)
	printDetailSection("✅ [CORRECT DETAILS]", report.CorrectDetail, Green+Bold)

	hasIssues := len(report.MissingDetail.Class) > 0 || len(report.MissingDetail.Method) > 0 || len(report.MissingDetail.Attribute) > 0 || len(report.MissingDetail.Edge) > 0 ||
		len(report.WrongDetail.Class) > 0 || len(report.WrongDetail.Method) > 0 || len(report.WrongDetail.Attribute) > 0 || len(report.WrongDetail.Edge) > 0

	if !hasIssues {
		fmt.Printf("\n%s✅ NO ISSUES FOUND — Perfect structural match!%s\n", Green+Bold, Reset)
	}
}

func printDetailSection(title string, detail domain.DetailError, color string) {
	isEmpty := len(detail.Class) == 0 && len(detail.Method) == 0 && len(detail.Attribute) == 0 && len(detail.Edge) == 0
	if isEmpty {
		return
	}

	fmt.Printf("\n%s%s%s\n", color, title, Reset)

	if len(detail.Class) > 0 {
		fmt.Printf("   %sNodes:%s\n", Bold, Reset)
		for _, d := range detail.Class {
			solName := "?"
			if d.Sol != nil {
				solName = d.Sol.Name
			}
			stuName := "?"
			if d.Stu != nil {
				stuName = d.Stu.Name
			}
			fmt.Printf("    • [%s vs %s] %s\n", solName, stuName, d.Description)
		}
	}
	if len(detail.Attribute) > 0 {
		fmt.Printf("   %sAttributes:%s\n", Bold, Reset)
		for _, d := range detail.Attribute {
			solV := "?"
			if d.Sol != nil {
				solV = attrString(d.Sol)
			}
			stuV := "?"
			if d.Stu != nil {
				stuV = attrString(d.Stu)
			}
			fmt.Printf("    • [%s] %s vs %s → %s\n", d.ParentClassName, solV, stuV, d.Description)
		}
	}
	if len(detail.Method) > 0 {
		fmt.Printf("   %sMethods:%s\n", Bold, Reset)
		for _, d := range detail.Method {
			solV := "?"
			if d.Sol != nil {
				solV = methString(d.Sol)
			}
			stuV := "?"
			if d.Stu != nil {
				stuV = methString(d.Stu)
			}
			fmt.Printf("    • [%s] %s vs %s → %s\n", d.ParentClassName, solV, stuV, d.Description)
		}
	}
	if len(detail.Edge) > 0 {
		fmt.Printf("   %sRelationships:%s\n", Bold, Reset)
		for _, d := range detail.Edge {
			solV := "?"
			if d.Sol != nil {
				solV = d.Sol.RelationType
			}
			stuV := "?"
			if d.Stu != nil {
				stuV = d.Stu.RelationType
			}
			fmt.Printf("    • [%s vs %s] %s\n", solV, stuV, d.Description)
		}
	}
}

func attrString(a *domain.ProcessedAttribute) string {
	return fmt.Sprintf("%s %s %s", a.Scope, a.Type, a.Name)
}

func methString(m *domain.ProcessedMethod) string {
	params := []string{}
	for _, p := range m.Inputs {
		params = append(params, p.Type)
	}
	return fmt.Sprintf("%s %s(%s): %s", m.Scope, m.Name, strings.Join(params, ", "), m.Output)
}

func nodeNames(g *domain.UMLGraph) []string {
	names := make([]string, len(g.Nodes))
	for i, n := range g.Nodes {
		names[i] = n.Name
	}
	return names
}

func edgeSummaries(g *domain.UMLGraph) []string {
	byID := make(map[string]string, len(g.Nodes))
	for _, n := range g.Nodes {
		byID[n.ID] = n.Name
	}
	sums := make([]string, len(g.Edges))
	for i, e := range g.Edges {
		src := byID[e.SourceID]
		tgt := byID[e.TargetID]
		if src == "" {
			src = e.SourceID
		}
		if tgt == "" {
			tgt = e.TargetID
		}
		sums[i] = fmt.Sprintf("%s -[%s]-> %s", src, e.RelationType, tgt)
	}
	return sums
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func printAllMembers(label string, solList, stuList []string, col int) {
	// Create a map to track which student members have been matched
	matchedStu := make(map[int]bool)

	// Interleave matched and mismatched from solution perspective
	for _, s := range solList {
		foundIdx := -1
		for i, st := range stuList {
			if matchedStu[i] {
				continue
			}
			// Exact match (including scope)
			if strings.TrimSpace(s) == strings.TrimSpace(st) {
				foundIdx = i
				break
			}
		}

		if foundIdx != -1 {
			// Matched: Both Green
			matchedStu[foundIdx] = true
			fmt.Printf("%s%-*s%s│ %s%-*s%s\n", Green, col, "    ✔ ["+label+"] "+s, Reset, Green, col, " ✔ ["+label+"] "+s, Reset)
		} else {
			// Not in student: Left Red
			solPart := "    ✗ [" + label + "] " + s
			if len(solPart) > col-1 {
				solPart = solPart[:col-4] + "..."
			}
			fmt.Printf("%s%-*s%s│  %-*s\n", Red, col, solPart, Reset, col, "")
		}
	}

	// Print remaining student members (extra ones)
	for i, st := range stuList {
		if !matchedStu[i] {
			fmt.Printf("%-*s│ %s%-*s%s\n", col, "", Red, col, " ✗ ["+label+"] "+st, Reset)
		}
	}
}
