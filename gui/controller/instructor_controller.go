package controller

import (
	"fmt"
	"log"

	"uml_compare/domain"
	"uml_compare/gui/service"
)

type instructorController struct {
	service service.InstructorService
	view    domain.InstructorView
}

func NewInstructorController(srv service.InstructorService, v domain.InstructorView) domain.InstructorController {
	c := &instructorController{
		service: srv,
		view:    v,
	}
	// Initial sync
	th, ai, avail := srv.GetConfig()
	v.UpdateConfigUI(th, ai, avail)
	return c
}

func (c *instructorController) OnUpdateConfig(threshold float64, useAI bool) {
	c.service.UpdateConfig(threshold, useAI)
	// No notification needed for silent background update, or we could show success
	c.view.ShowSuccess(fmt.Sprintf("Global configuration updated: Threshold=%.1f, UseAI=%v", threshold, useAI))
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
