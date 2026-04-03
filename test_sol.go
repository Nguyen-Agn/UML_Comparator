package main

import (
	"fmt"
	"uml_compare/parser"
	"uml_compare/builder"
	"uml_compare/prematcher"
)

func main() {
	raw, err := parser.NewDrawioParser().Parse("UMLs_testcase/question7.drawio")
	if err != nil {
		panic(err)
	}
	b, err := builder.GetBuilder("drawio")
	if err != nil {
		panic(err)
	}
	graph, err := b.Build(raw)
	if err != nil {
		panic(err)
	}
	solPre := prematcher.NewUMLSolutionPreMatcher()
	pGraph, err := solPre.ProcessSolution(graph)
	if err != nil {
		panic(err)
	}
	for _, n := range pGraph.Nodes {
		fmt.Printf("Node: %s, Score: %f\n", n.Name, n.Score)
		for _, a := range n.Attributes {
			fmt.Printf("  Attr: %v, Score: %f\n", a.Names, a.Score)
		}
		for _, m := range n.Methods {
			fmt.Printf("  Method: %v, Score: %f\n", m.Names, m.Score)
		}
	}
	fmt.Printf("Grading Config Nodes: %v\n", pGraph.GradingConfig.Nodes)
	fmt.Printf("Grading Config Attrs: %v\n", pGraph.GradingConfig.Attributes)
	fmt.Printf("Grading Config Methods: %v\n", pGraph.GradingConfig.Methods)
	fmt.Printf("Grading Config Edges: %v\n", pGraph.GradingConfig.Edges)
}
