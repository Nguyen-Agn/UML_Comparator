package controller

import (
	"uml_compare/domain"
)

type defaultMainController struct {
	processor  domain.UMLProcessor
	view       domain.MainView
	lastResult *domain.GradeResult
}

// NewMainController creates a new MainController
func NewMainController(processor domain.UMLProcessor, view domain.MainView) domain.MainController {
	ctrl := &defaultMainController{
		processor: processor,
		view:      view,
	}
	// Sync AI Status
	view.ShowAIStatus(processor.IsAIAvailable())
	return ctrl
}

// OnSubmit handles the file submission from the view
func (c *defaultMainController) OnSubmit(solutionPath string, assignmentPath string) {
	if solutionPath == "" || assignmentPath == "" {
		return
	}
	c.view.ShowLoading()

	// Process in a goroutine to avoid blocking the Fyne UI thread
	go func() {
		res, err := c.processor.Process(solutionPath, assignmentPath)
		if err != nil {
			c.view.ShowError(err)
			return
		}

		c.lastResult = res
		c.view.ShowResult(res)
		c.view.EnableExport()
	}()
}

// OnExport handles the instruction to export an HTML file
func (c *defaultMainController) OnExport(saveFilePath string) {
	if c.lastResult == nil || saveFilePath == "" {
		return
	}

	err := c.processor.ExportHTML(c.lastResult, saveFilePath)
	if err != nil {
		c.view.ShowError(err)
	}
}
