package controller

import (
	"context"
	"fmt"
	"log"

	"uml_compare/domain"
)

// generatorController implements domain.IGeneratorController.
// It dispatches view events to the generator service and feeds results back to the view.
type generatorController struct {
	service domain.IUMLGeneratorService
	view    domain.IGeneratorView
}

// NewGeneratorController wires the service and view together.
func NewGeneratorController(svc domain.IUMLGeneratorService, v domain.IGeneratorView) domain.IGeneratorController {
	return &generatorController{service: svc, view: v}
}

// OnGenerate is triggered when the user clicks "Tạo mẫu UML".
// Runs the AI call in a goroutine to keep the UI responsive.
func (c *generatorController) OnGenerate(problemText string) {
	go func() {
		result, err := c.service.Generate(context.Background(), problemText)
		if err != nil {
			log.Printf("generator_controller: Generate error: %v\n", err)
			c.view.ShowError(fmt.Errorf("Không thể tạo UML: %v", err))
			return
		}
		c.view.ShowGeneratedUML(result.MermaidCode)
	}()
}

// OnSave is triggered when the user clicks "Lưu file" and confirms a path.
func (c *generatorController) OnSave(mermaidCode string, outputPath string) {
	go func() {
		// Import lazily to avoid circular package dependency; exporter is in a sibling package.
		// We call the exporter via its exported function directly.
		if err := exportMermaid(mermaidCode, outputPath); err != nil {
			log.Printf("generator_controller: Save error: %v\n", err)
			c.view.ShowError(fmt.Errorf("Lỗi lưu file: %v", err))
			return
		}
		c.view.ShowSuccess("Đã lưu file: " + outputPath)
	}()
}

// OnGetConfig returns the current provider config to populate the settings UI.
func (c *generatorController) OnGetConfig() domain.GeneratorConfig {
	return c.service.GetConfig()
}

// OnSaveConfig persists new provider config and returns any error.
func (c *generatorController) OnSaveConfig(cfg domain.GeneratorConfig) error {
	return c.service.SaveConfig(cfg)
}
