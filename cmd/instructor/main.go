package instructor

import (
	"log"
	"uml_compare/gui/controller"
	"uml_compare/gui/service"
	"uml_compare/gui/view"
)

func main() {
	// Initialize Service Layer
	srv := service.NewStandardInstructorService()

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
