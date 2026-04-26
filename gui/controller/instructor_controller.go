package controller

import (
	"log"

	"uml_compare/domain"
	"uml_compare/gui/service"
)

type instructorController struct {
	service service.InstructorService
	view    domain.InstructorView
}

func NewInstructorController(srv service.InstructorService, v domain.InstructorView) domain.InstructorController {
	return &instructorController{
		service: srv,
		view:    v,
	}
}

func (c *instructorController) OnEncrypt(inputPath, outputPath string) {
	c.view.ShowLoading()
	go func() {
		err := c.service.EncryptSolution(inputPath, outputPath)
		c.view.HideLoading()
		if err != nil {
			c.view.ShowError(err)
		} else {
			c.view.ShowSuccess("File successfully encrypted to " + outputPath)
		}
	}()
}

func (c *instructorController) OnBuildExam(solutionsDir, outputPath string) {
	c.view.ShowLoading()
	go func() {
		err := c.service.BuildExamTool(solutionsDir, outputPath)
		c.view.HideLoading()
		if err != nil {
			c.view.ShowError(err)
		} else {
			c.view.ShowSuccess("Exam Builder created successfully at " + outputPath)
		}
	}()
}

func (c *instructorController) OnGradeBatch(solutionPath, studentDir, outputPath string) {
	c.view.ShowLoading()
	go func() {
		res, err := c.service.GradeBatch(solutionPath, studentDir, outputPath)
		c.view.HideLoading()
		if err != nil {
			c.view.ShowError(err)
		} else {
			log.Printf("Batch graded %d files in %v.\n", res.TotalFiles, res.Duration)
			c.view.ShowSuccess("Batch grading completed! Report saved to " + outputPath)
		}
	}()
}

func (c *instructorController) OnLiveCompare(solutionPath, studentPath string) {
	c.view.ShowLoading()
	go func() {
		res, err := c.service.CompareUML(solutionPath, studentPath, true)
		if err != nil {
			c.view.HideLoading()
			c.view.ShowError(err)
			return
		}
		// In live compare, we hide loading inside ShowLiveCompareResult or let view handle it.
		c.view.ShowLiveCompareResult(res)
	}()
}
