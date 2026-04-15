package main

import (
	"log"
	"uml_compare/gui/controller"
	"uml_compare/gui/service"
	"uml_compare/gui/view"
)

func main() {
	// 1. Initialize SOLID layers
	proc := service.NewStandardUMLProcessor()

	// Create Lorca view
	v, err := view.NewMainView()
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
