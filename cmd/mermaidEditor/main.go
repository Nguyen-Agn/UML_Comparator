package main

import (
	"log"
	"uml_compare/gui/controller"
	"uml_compare/gui/view"
	"uml_compare/src/uml_generator"
	_ "embed"
)

//go:embed mermaid.min.js
var mermaidJS string

func main() {
	// Initialize the underlying generator service (only for file saving)
	srv, err := uml_generator.NewGeneratorService()

	// Initialize the standalone Editor View
	v, err := view.NewMermaidEditorView(mermaidJS)
	if err != nil {
		log.Fatal("Could not open Mermaid Editor GUI: ", err)
	}
	defer v.Close()

	// Reuse the generator controller
	ctrl := controller.NewGeneratorController(srv, v)
	v.SetGeneratorController(ctrl)

	// Keep running
	v.Wait()
}
