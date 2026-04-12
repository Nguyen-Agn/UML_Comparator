// cmd/prematch/main.go - Thử nghiệm prematcher: go run ./cmd/prematch/main.go <file.drawio>
package main

import (
	"fmt"
	"os"
	"uml_compare/cmd/share"
	"uml_compare/domain"
	"uml_compare/prematcher"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("%sUsage: go run cmd/prematch/main.go <file.drawio>%s\n", share.Yellow, share.Reset)
		fmt.Printf("Example: go run cmd/prematch/main.go UMLs_testcase/problem1.drawio\n")
		os.Exit(1)
	}

	filePath := os.Args[1]

	share.PrintBanner("UML Prematcher — Detail Extraction")
	fmt.Printf("Input File: %s%s%s\n\n", share.Cyan, filePath, share.Reset)

	processedGraph, err := run(filePath)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	printProcessedGraph(processedGraph)
}

// run thực hiện pipeline Parse → Build → PreMatch và trả về ProcessedUMLGraph.
func run(filePath string) (*domain.ProcessedUMLGraph, error) {
	graph, err := share.LoadGraph(filePath)
	if err != nil {
		return nil, err
	}

	pm := prematcher.NewStandardPreMatcher()
	return pm.Process(graph)
}

// ── Print Layer ───────────────────────────────────────────────────────────────

// printProcessedGraph in toàn bộ kết quả prematcher bao gồm ArchWeight và members.
func printProcessedGraph(g *domain.ProcessedUMLGraph) {
	fmt.Printf("\n%s── [RESULTS] Processed Nodes & ArchWeights ──────────────────%s\n", share.Cyan+share.Bold, share.Reset)

	for _, n := range g.Nodes {
		printProcessedNode(n)
	}

	fmt.Printf("\n%s── Summary: %d nodes processed successfully ─────────────────%s\n",
		share.Green+share.Bold, len(g.Nodes), share.Reset)
}

// printProcessedNode in chi tiết một node đã được xử lý.
func printProcessedNode(n domain.ProcessedNode) {
	fmt.Printf("\n%s● NODE: %s%s (Type: %s%s%s)\n", share.Bold+share.Cyan, n.Name, share.Reset, share.Yellow, n.Type, share.Reset)

	fmt.Printf("  %s├─ ArchWeight:%s %d\n", share.Blue, share.Reset, n.ArchWeight)
	printArchWeightBreakdown(n.ArchWeight)

	if n.Shortcut != 0 {
		fmt.Printf("  %s├─ Shortcuts:%s ", share.Blue, share.Reset)
		if (n.Shortcut & 1) != 0 {
			fmt.Printf("[Getters] ")
		}
		if (n.Shortcut & 2) != 0 {
			fmt.Printf("[Setters] ")
		}
		fmt.Println()
	}

	if len(n.Attributes) > 0 {
		fmt.Printf("  %s├─ Attributes (%d):%s\n", share.Blue, len(n.Attributes), share.Reset)
		for _, a := range n.Attributes {
			fmt.Printf("  │  • %s %s%-15s%s : %s%-10s%s [%s]\n",
				a.Scope, share.Bold, a.Name, share.Reset, share.Yellow, a.Type, share.Reset, a.Kind)
		}
	}

	if len(n.Methods) > 0 {
		fmt.Printf("  %s└─ Methods (%d):%s\n", share.Blue, len(n.Methods), share.Reset)
		for _, m := range n.Methods {
			mColor := share.Reset
			switch m.Type {
			case "getter", "setter":
				mColor = share.Green
			case "constructor":
				mColor = share.Yellow + share.Bold
			}
			kindStr := ""
			if m.Kind != "normal" {
				kindStr = fmt.Sprintf(" [%s]", m.Kind)
			}
			fmt.Printf("     • %s %s%-20s%s : %s%-10s%s (Type: %s%s%s)%s\n",
				m.Scope, mColor, m.Name+"()", share.Reset, share.Yellow, m.Output, share.Reset,
				share.Cyan, m.Type, share.Reset, kindStr)
		}
	}
}

// printArchWeightBreakdown in chi tiết từng vùng bit của ArchWeight.
func printArchWeightBreakdown(weight uint32) {
	// Bit 29-31: Loại Class (3 bit)
	typeVal := (weight >> 29) & 0x7
	typeName := "Unknown"
	switch typeVal {
	case 1:
		typeName = "Class"
	case 2:
		typeName = "Interface"
	case 3:
		typeName = "Abstract"
	case 4:
		typeName = "Enum"
	}

	hasInherit := (weight >> 28) & 0x1 // Bit 28: Thừa kế
	numIntf := (weight >> 24) & 0xF     // Bit 24-27: Interface count
	numMeth := (weight >> 18) & 0x3F    // Bit 18-23: Method count
	numAttr := (weight >> 13) & 0x1F    // Bit 13-17: Attribute count

	fmt.Printf("  %s│  └─ Binary:%s %032b\n", share.Blue, share.Reset, weight)
	fmt.Printf("  %s│     [Bits 29-31] Type: %s%-10s%s | [Bit 28] Inherit: %v\n",
		share.Blue, share.Yellow, typeName, share.Reset, hasInherit != 0)
	fmt.Printf("  %s│     [Bits 24-27] Intf: %-10d | [Bits 18-23] Meth: %-2d\n", share.Blue, numIntf, numMeth)
	fmt.Printf("  %s│     [Bits 13-17] Attr: %-10d\n", share.Blue, numAttr)
}
