package main

import (
	"fmt"
	"os"

	"uml_compare/gui/controller"
	"uml_compare/gui/view"
	"uml_compare/src/uml_generator"
)

func main() {
	// 1. Initialise the generator service (reads config from disk)
	svc, err := uml_generator.NewGeneratorService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialise generator service: %v\n", err)
		os.Exit(1)
	}

	// 2. Create the Lorca GUI window
	v, err := view.NewGeneratorView()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open generator window: %v\n", err)
		os.Exit(1)
	}
	defer v.Close()

	// 3. Wire controller
	ctrl := controller.NewGeneratorController(svc, v)
	v.SetGeneratorController(ctrl)

	// 4. Block until window is closed
	v.Wait()
}
