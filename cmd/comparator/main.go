package main

import (
	"fmt"
	"log"
	"os"

	"uml_compare/builder"
	"uml_compare/comparator"
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
		fmt.Println("Usage: go run cmd/comparator/main.go <solution.drawio> <student.drawio>")
		os.Exit(1)
	}

	solPath := os.Args[1]
	stuPath := os.Args[2]

	fmt.Printf("%s╔══════════════════════════════════════════════════════════╗%s\n", Blue, Reset)
	fmt.Printf("%s║          Advanced Comparator End-to-End Pipeline         ║%s\n", Blue+Bold, Reset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════╝%s\n", Blue, Reset)
	fmt.Printf("Solution File: %s%s%s\n", Cyan, solPath, Reset)
	fmt.Printf("Student File:  %s%s%s\n\n", Cyan, stuPath, Reset)

	// 1. Initialize Interfaces
	fmt.Printf("%s[1] Initializing Pipeline Interfaces...%s\n", Blue, Reset)
	var fileParser parser.IFileParser = parser.NewDrawioParser()
	var modelBuilder builder.IModelBuilder = builder.NewStandardModelBuilder()
	var preMatcher prematcher.IPreMatcher = prematcher.NewStandardPreMatcher()
	
	fuzzy := matcher.NewLevenshteinMatcher()
	var entityMatcher matcher.IEntityMatcher = matcher.NewStandardEntityMatcher(fuzzy, 0.8) // Threshold 80%

	var comp comparator.IComparator = comparator.NewStandardComparator(fuzzy)

	// 2. Run Pipeline for Solution
	fmt.Printf("%s[2] Processing Solution Graph...%s\n", Blue, Reset)
	solRaw, err := fileParser.Parse(solPath)
	if err != nil {
		log.Fatalf("Failed to parse solution file: %v\n", err)
	}
	solGraph, err := modelBuilder.Build(solRaw)
	if err != nil {
		log.Fatalf("Failed to build solution graph: %v\n", err)
	}
	solProcessed, err := preMatcher.Process(solGraph)
	if err != nil {
		log.Fatalf("Failed to calculate solution ArchWeights: %v\n", err)
	}

	// 3. Run Pipeline for Student
	fmt.Printf("%s[3] Processing Student Graph...%s\n", Blue, Reset)
	stuRaw, err := fileParser.Parse(stuPath)
	if err != nil {
		log.Fatalf("Failed to parse student file: %v\n", err)
	}
	stuGraph, err := modelBuilder.Build(stuRaw)
	if err != nil {
		log.Fatalf("Failed to build student graph: %v\n", err)
	}
	stuProcessed, err := preMatcher.Process(stuGraph)
	if err != nil {
		log.Fatalf("Failed to calculate student ArchWeights: %v\n", err)
	}

	// 4. Run Matcher
	fmt.Printf("%s[4] Running Entity Matcher (Exact -> ArchWeight -> Fuzzy)...%s\n", Blue, Reset)
	mapping, err := entityMatcher.Match(solProcessed, stuProcessed)
	if err != nil {
		log.Fatalf("Matcher error: %v\n", err)
	}

	// 5. Run Comparator
	fmt.Printf("%s[5] Running Advanced Comparator (TypeMap -> Attributes -> Methods -> Edges)...%s\n", Blue, Reset)
	diffReport, err := comp.Compare(solProcessed, stuProcessed, mapping)
	if err != nil {
		log.Fatalf("Comparator error: %v\n", err)
	}

	// 6. Output Review
	fmt.Printf("\n%s===============================================%s\n", Blue, Reset)
	fmt.Printf("%s             CLASS REVIEW REPORT               %s\n", Bold+Blue, Reset)
	fmt.Printf("%s===============================================%s\n", Blue, Reset)

	if len(diffReport.MissedClass) > 0 {
		fmt.Printf("\n%s🚨 [MISSED CLASSES] - Các Lớp bị thiếu hoàn toàn:%s\n", Red+Bold, Reset)
		for _, m := range diffReport.MissedClass { fmt.Printf("   ❌ %s\n", m) }
	}
	if len(diffReport.MissingNodes) > 0 {
		fmt.Printf("\n%s🚨 [MISSED NODES] - Các Đối tượng khác bị thiếu:%s\n", Red+Bold, Reset)
		for _, m := range diffReport.MissingNodes { fmt.Printf("   ❌ %s\n", m) }
	}
	if len(diffReport.MissingEdges) > 0 {
		fmt.Printf("\n%s🚨 [MISSED EDGES] - Các Mối quan hệ bị thiếu:%s\n", Red+Bold, Reset)
		for _, m := range diffReport.MissingEdges { fmt.Printf("   ❌ %s\n", m) }
	}
	if len(diffReport.MissingMembers) > 0 {
		fmt.Printf("\n%s🚨 [MISSED MEMBERS] - Các Thuộc tính/Phương thức bị thiếu:%s\n", Red+Bold, Reset)
		for _, m := range diffReport.MissingMembers { fmt.Printf("   ❌ %s\n", m) }
	}

	if len(diffReport.AttributeErrors) > 0 {
		fmt.Printf("\n%s⚠️ [ATTRIBUTE ERRORS] - Thuộc tính sai kiểu/scope/tên:%s\n", Yellow+Bold, Reset)
		for _, m := range diffReport.AttributeErrors { fmt.Printf("   %s🔸%s %s\n", Yellow, Reset, m) }
	}
	if len(diffReport.MethodErrors) > 0 {
		fmt.Printf("\n%s⚠️ [METHOD ERRORS] - Phương thức sai tham số/scope/lệch đếm:%s\n", Yellow+Bold, Reset)
		for _, m := range diffReport.MethodErrors { fmt.Printf("   %s🔸%s %s\n", Yellow, Reset, m) }
	}
	if len(diffReport.NodeEdgeErrors) > 0 {
		fmt.Printf("\n%s⚠️ [RELATIONSHIP ERRORS] - Mũi tên ngược/Liên kết sai:%s\n", Yellow+Bold, Reset)
		for _, m := range diffReport.NodeEdgeErrors { fmt.Printf("   %s🔸%s %s\n", Yellow, Reset, m) }
	}

	if len(diffReport.MissedClass) == 0 && len(diffReport.MissingNodes) == 0 && len(diffReport.MissingEdges) == 0 && len(diffReport.MissingMembers) == 0 &&
	   len(diffReport.AttributeErrors) == 0 && len(diffReport.MethodErrors) == 0 && len(diffReport.NodeEdgeErrors) == 0 {
		fmt.Printf("\n%s✅ BẠN ĐÃ ĐẠT ĐIỂM TUYỆT ĐỐI! KHÔNG TÌM THẤY LỖI NÀO (PERFECT MATCH)!%s\n", (Green + Bold), Reset)
	}
	
	fmt.Printf("\n%s===============================================%s\n", Blue, Reset)
	fmt.Printf("%sDone.%s\n", Green, Reset)
}
