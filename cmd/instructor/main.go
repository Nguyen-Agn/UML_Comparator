package main

import (
	"log"
	"uml_compare/domain"
	"uml_compare/gui/controller"
	"uml_compare/gui/service"
	"uml_compare/gui/view"
	"uml_compare/src/uml_generator"
)

func main() {
	// Initialize Service Layer
	srv := service.NewStandardInstructorService()
	genSrv, err := uml_generator.NewGeneratorService()
	if err != nil {
		log.Println("Warning: generator service failed to init:", err)
	}

	// Initialize View Layer
	v, err := view.NewInstructorView()
	if err != nil {
		log.Fatal("Could not open Instructor GUI: ", err)
	}
	defer v.Close()

	// Initialize and Bind Controllers
	ctrl := controller.NewInstructorController(srv, v)
	v.SetController(ctrl)

	// In view package, we assert v to have SetGeneratorController
	// It's a bit of an interface cast, or we can change view.NewInstructorView return type.
	// We'll use a type assertion.
	if hv, ok := v.(interface {
		SetGeneratorController(domain.IGeneratorController)
	}); ok {
		genView, _ := v.(domain.IGeneratorView)
		genCtrl := controller.NewGeneratorController(genSrv, genView)
		hv.SetGeneratorController(genCtrl)
	}

	// Wait for window to close
	v.Wait()
}
