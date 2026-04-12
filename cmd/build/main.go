// cmd/build/main.go - Chạy bằng: go run ./cmd/build/main.go <file.drawio>
// Hiển thị trực quan cấu trúc UMLGraph sau khi Builder xử lý file .drawio
package main

import (
	"fmt"
	"os"
	"strings"
	"uml_compare/cmd/share"
	"uml_compare/domain"
)

func main() {
	filePath := "parser/testdata/plain_sample.drawio"
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	share.PrintBanner("UML Compare - Builder Visual Output Demo")
	fmt.Printf("📂 Input file : %s\n\n", filePath)

	graph, err := run(filePath)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	printGraph(graph)
}

// run thực hiện toàn bộ pipeline Parse → Build và trả về UMLGraph.
func run(filePath string) (*domain.UMLGraph, error) {
	return share.LoadGraph(filePath)
}

// ── Print Layer ───────────────────────────────────────────────────────────────

// printGraph in tổng quan graph và gọi các hàm in chi tiết.
func printGraph(g *domain.UMLGraph) {
	fmt.Printf("✅ [Builder] UMLGraph built successfully\n")
	fmt.Printf("   ├─ Nodes : %d\n", len(g.Nodes))
	fmt.Printf("   └─ Edges : %d\n\n", len(g.Edges))

	printNodes(g)
	printEdges(g)
	printEdgeDiagram(g)
}

func printNodes(g *domain.UMLGraph) {
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│                     NODES (Classes)                 │")
	fmt.Println("├─────────────────────────────────────────────────────┤")
	for i, n := range g.Nodes {
		fmt.Printf("│ [%d] %-10s  Type: %-12s  ID: %s\n", i+1, n.Name, n.Type, n.ID)
		if len(n.Attributes) > 0 {
			fmt.Println("│      Attributes:")
			for _, a := range n.Attributes {
				fmt.Printf("│        • %s\n", a)
			}
		}
		if len(n.Methods) > 0 {
			fmt.Println("│      Methods:")
			for _, m := range n.Methods {
				fmt.Printf("│        ◆ %s\n", m)
			}
		}
		if i < len(g.Nodes)-1 {
			fmt.Println("├─────────────────────────────────────────────────────┤")
		}
	}
	fmt.Println("└─────────────────────────────────────────────────────┘")
	fmt.Println()
}

func printEdges(g *domain.UMLGraph) {
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│                     EDGES (Relations)               │")
	fmt.Println("├─────────────────────────────────────────────────────┤")
	if len(g.Edges) == 0 {
		fmt.Println("│  (no edges found)                                   │")
	}
	for i, e := range g.Edges {
		arrow := relationArrow(e.RelationType)
		srcName := nodeNameByID(g, e.SourceID)
		tgtName := nodeNameByID(g, e.TargetID)
		fmt.Printf("│ [%d] %-12s %s %-12s  [%s]\n", i+1, srcName, arrow, tgtName, e.RelationType)
	}
	fmt.Println("└─────────────────────────────────────────────────────┘")
	fmt.Println()
}

func printEdgeDiagram(g *domain.UMLGraph) {
	fmt.Println("── ASCII Relation Diagram ────────────────────────────")
	for _, e := range g.Edges {
		src := nodeNameByID(g, e.SourceID)
		tgt := nodeNameByID(g, e.TargetID)
		arrow := relationArrow(e.RelationType)
		pad := strings.Repeat(" ", maxInt(0, 12-len(src)))
		fmt.Printf("  [ %-10s ]%s%s──▶ [ %-10s ]\n", src, pad, arrow, tgt)
	}
	fmt.Println("──────────────────────────────────────────────────────")
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func nodeNameByID(g *domain.UMLGraph, id string) string {
	for _, n := range g.Nodes {
		if n.ID == id {
			return n.Name
		}
	}
	return id
}

func relationArrow(rel string) string {
	switch rel {
	case "Inheritance":
		return "══════▷"
	case "Realization":
		return "- - -▷"
	case "Composition":
		return "◆─────"
	case "Aggregation":
		return "◇─────"
	case "Dependency":
		return "·····▶"
	default:
		return "──────"
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
