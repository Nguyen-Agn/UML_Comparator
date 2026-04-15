package main

import (
	"embed"
	"log"

	"uml_compare/gui/controller"
	"uml_compare/gui/service"
	"uml_compare/gui/view"
)

//go:embed embedded_solutions
var solutionsFS embed.FS

func main() {
	// Extract content
	embeddedFiles := make(map[string][]byte)

	entries, err := solutionsFS.ReadDir("embedded_solutions")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				data, err := solutionsFS.ReadFile("embedded_solutions/" + entry.Name())
				if err == nil {
					embeddedFiles[entry.Name()] = data
				}
			}
		}
	} else {
		log.Printf("Warning: Could not read embedded_solutions folder: %v", err)
	}

	// 1. Initialize SOLID layers
	proc := service.NewStandardUMLProcessor()

	// Create Exam Lorca view
	v, err := view.NewExamMainView(embeddedFiles)
	if err != nil {
		log.Fatal(err)
	}
	defer v.Close()

	ctrl := controller.NewMainController(proc, v)

	// 2. Inject dependency
	v.SetController(ctrl)

	// 3. Keep main goroutine running until UI is closed
	v.Wait()
}
