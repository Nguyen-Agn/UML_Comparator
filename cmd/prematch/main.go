package main

import (
	"fmt"
	"log"
	"os"

	"uml_compare/builder"
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
	if len(os.Args) < 2 {
		fmt.Printf("%sUsage: go run cmd/prematch/main.go <file.drawio>%s\n", Yellow, Reset)
		fmt.Printf("Example: go run cmd/prematch/main.go UMLs_testcase/problem1.drawio\n")
		os.Exit(1)
	}

	filePath := os.Args[1]

	fmt.Printf("%s╔══════════════════════════════════════════════════════════╗%s\n", Blue, Reset)
	fmt.Printf("%s║           UML Prematcher — Detail Extraction             ║%s\n", Blue+Bold, Reset)
	fmt.Printf("%s╚══════════════════════════════════════════════════════════╝%s\n", Blue, Reset)
	fmt.Printf("Input File: %s%s%s\n\n", Cyan, filePath, Reset)

	// 1. Initialize Pipeline
	fmt.Printf("%s[1] Initializing Pipeline...%s\n", Blue, Reset)
	p, err := parser.GetParser(filePath)
	if err != nil {
		fmt.Printf("❌ Parser error: %v\n", err)
		os.Exit(1)
	}
	b := builder.NewStandardModelBuilder()
	pm := prematcher.NewStandardPreMatcher()

	// 2. Parse File
	fmt.Printf("%s[2] Parsing .drawio file...%s\n", Blue, Reset)
	rawXML, err := p.Parse(filePath)
	if err != nil {
		log.Fatalf("%s❌ Parser error: %v%s", Red, err, Reset)
	}

	// 3. Build Graph
	fmt.Printf("%s[3] Building UML Graph...%s\n", Blue, Reset)
	graph, err := b.Build(rawXML)
	if err != nil {
		log.Fatalf("%s❌ Builder error: %v%s", Red, err, Reset)
	}

	// 4. Process with Prematcher
	fmt.Printf("%s[4] Running Prematcher (Detailed Extraction & ArchWeight)...%s\n", Blue, Reset)
	processedGraph, err := pm.Process(graph)
	if err != nil {
		log.Fatalf("%s❌ Error processing graph: %v%s", Red, err, Reset)
	}

	fmt.Printf("\n%s── [RESULTS] Processed Nodes & ArchWeights ──────────────────%s\n", Cyan+Bold, Reset)
	for _, n := range processedGraph.Nodes {
		fmt.Printf("\n%s● NODE: %s%s (Type: %s%s%s)\n", Bold+Cyan, n.Name, Reset, Yellow, n.Type, Reset)

		fmt.Printf("  %s├─ ArchWeight:%s %d\n", Blue, Reset, n.ArchWeight)
		printArchWeightBreakdown(n.ArchWeight)

		if n.Shortcut != 0 {
			fmt.Printf("  %s├─ Shortcuts:%s ", Blue, Reset)
			if (n.Shortcut & 1) != 0 {
				fmt.Printf("[Getters] ")
			}
			if (n.Shortcut & 2) != 0 {
				fmt.Printf("[Setters] ")
			}
			fmt.Println()
		}

		if len(n.Attributes) > 0 {
			fmt.Printf("  %s├─ Attributes (%d):%s\n", Blue, len(n.Attributes), Reset)
			for _, a := range n.Attributes {
				fmt.Printf("  │  • %s %s%-15s%s : %s%-10s%s [%s]\n", a.Scope, Bold, a.Name, Reset, Yellow, a.Type, Reset, a.Kind)
			}
		}

		if len(n.Methods) > 0 {
			fmt.Printf("  %s└─ Methods (%d):%s\n", Blue, len(n.Methods), Reset)
			for _, m := range n.Methods {
				mColor := Reset
				if m.Type == "getter" || m.Type == "setter" {
					mColor = Green
				} else if m.Type == "constructor" {
					mColor = Yellow + Bold
				}
				fmt.Printf("     • %s %s%-20s%s : %s%-10s%s (Type: %s%s%s)\n", m.Scope, mColor, m.Name+"()", Reset, Yellow, m.Output, Reset, Cyan, m.Type, Reset)
			}
		}
	}

	fmt.Printf("\n%s── Summary: %d nodes processed successfully ──%s\n", Green+Bold, len(processedGraph.Nodes), Reset)
}

func printArchWeightBreakdown(weight uint32) {
	// Bit 29-31: Loại Class (3 bit - 0: Unknown, 1: Class, 2: Interface, 3: Abstract, 4: Enum)
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

	// Bit 28: Thừa kế
	hasInherit := (weight >> 28) & 0x1

	// Bit 24-27: Interface
	numIntf := (weight >> 24) & 0xF

	// Bit 18-23: Method
	numMeth := (weight >> 18) & 0x3F

	// Bit 13-17: Attribute
	numAttr := (weight >> 13) & 0x1F

	fmt.Printf("  %s│  └─ Binary:%s %032b\n", Blue, Reset, weight)
	fmt.Printf("  %s│     [Bits 29-31] Type: %s%-10s%s | [Bit 28] Inherit: %v\n", Blue, Yellow, typeName, Reset, hasInherit != 0)
	fmt.Printf("  %s│     [Bits 24-27] Intf: %-10d | [Bits 18-23] Meth: %-2d\n", Blue, numIntf, numMeth)
	fmt.Printf("  %s│     [Bits 13-17] Attr: %-10d\n", Blue, numAttr)
}
