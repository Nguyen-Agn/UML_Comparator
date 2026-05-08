package domain

import "context"

// IUMLGeneratorService orchestrates the AI-based UML generation pipeline.
type IUMLGeneratorService interface {
	// Generate sends the problem description to an LLM and returns the
	// generated Mermaid code formatted according to the project's scoring
	// syntax (__d__ tags, | polymorphism). One HTTP call is made per invocation.
	Generate(ctx context.Context, problemText string) (*GeneratedUML, error)

	// GetConfig returns the currently persisted LLM provider configuration.
	GetConfig() GeneratorConfig

	// SaveConfig persists the LLM provider configuration to local storage.
	SaveConfig(cfg GeneratorConfig) error
}

// IGeneratorController dispatches user actions from the view to the service.
type IGeneratorController interface {
	// OnGenerate is called when the user clicks "Tạo mẫu UML".
	OnGenerate(problemText string)

	// OnSave is called when the user clicks "Lưu file".
	// mermaidCode is the (possibly edited) content; outputPath is the destination .mmd file.
	OnSave(mermaidCode string, outputPath string)

	// OnGetConfig returns the current generator configuration to populate the settings UI.
	OnGetConfig() GeneratorConfig

	// OnSaveConfig persists new configuration entered by the user.
	OnSaveConfig(cfg GeneratorConfig) error
}

// IGeneratorView abstracts the Lorca-based GUI for the UML generator window.
type IGeneratorView interface {
	// SetGeneratorController injects the controller dependency.
	SetGeneratorController(c IGeneratorController)

	// ShowLoading shows a loading overlay while the AI is processing.
	ShowLoading()

	// HideLoading removes the loading overlay.
	HideLoading()

	// ShowError displays an error notification to the user.
	ShowError(err error)

	// ShowGeneratedUML passes the Mermaid source to the JS editor and live preview.
	ShowGeneratedUML(mermaidCode string)

	// ShowSuccess displays a success notification.
	ShowSuccess(msg string)

	// Wait blocks until the window is closed.
	Wait()

	// Close terminates the window.
	Close()
}

// GeneratedUML holds the result of a single AI generation call.
type GeneratedUML struct {
	// MermaidCode is the raw Mermaid classDiagram text with __d__ score tags.
	MermaidCode string
}

// GeneratorConfig holds the LLM provider settings persisted locally.
type GeneratorConfig struct {
	// APIKey is the authentication key for the provider (empty = unauthenticated, e.g. Ollama).
	APIKey string `json:"api_key"`

	// APIEndpoint is the base URL of an OpenAI-compatible endpoint.
	// Examples:
	//   https://api.openai.com/v1        (OpenAI)
	//   http://localhost:11434/v1         (Ollama)
	//   https://api.groq.com/openai/v1   (Groq)
	APIEndpoint string `json:"api_endpoint"`

	// Model is the model identifier to use for generation.
	// Examples: gpt-4o-mini, gemma3, llama3.2
	Model string `json:"model"`
}
