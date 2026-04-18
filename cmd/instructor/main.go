package main

import (
	"log"

	"uml_compare/gui/controller"
	"uml_compare/gui/view"
	"uml_compare/instructor"
)

func main() {
	// Initialize Service Layer
	srv := instructor.NewStandardInstructorService()

	// Initialize View Layer
	v, err := view.NewInstructorView()
	if err != nil {
		log.Fatal("Could not open Instructor GUI: ", err)
	}
	defer v.Close()

	// Initialize and Bind Controller
	ctrl := controller.NewInstructorController(srv, v)
	v.SetController(ctrl)

	// Keep running
	v.Wait()
}
