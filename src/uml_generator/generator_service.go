package uml_generator

import (
	"context"
	"fmt"
	"log"
	"strings"

	"uml_compare/domain"
)

// generatorService implements domain.IUMLGeneratorService.
// It orchestrates: PromptBuilder → ILLMProvider → return GeneratedUML.
// Parsing into UMLGraph is intentionally skipped here — the HTML view renders
// Mermaid directly via Mermaid.js, and the parser is only invoked at save time
// by the solution exporter.
type generatorService struct {
	provider      iLLMProvider
	promptBuilder *promptBuilder
	store         *configStore
}

// Compile-time check that generatorService satisfies the domain interface.
var _ domain.IUMLGeneratorService = (*generatorService)(nil)

// NewGeneratorService constructs a generatorService wired to the persisted config.
// Call this once at startup; it reads config from disk and builds the provider.
func NewGeneratorService() (domain.IUMLGeneratorService, error) {
	store, err := newConfigStore()
	if err != nil {
		return nil, fmt.Errorf("generator_service: init config store: %w", err)
	}
	cfg, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("generator_service: load config: %w", err)
	}
	return &generatorService{
		provider:      newOpenAICompatibleProvider(cfg.APIEndpoint, cfg.APIKey, cfg.Model),
		promptBuilder: newPromptBuilder(),
		store:         store,
	}, nil
}

// Generate calls the LLM with a single HTTP request and returns the Mermaid code.
// The context can be cancelled by the caller (e.g. user clicks "Stop").
func (s *generatorService) Generate(ctx context.Context, problemText string) (*domain.GeneratedUML, error) {
	if strings.TrimSpace(problemText) == "" {
		return nil, fmt.Errorf("generator_service: problem text is empty")
	}

	messages := s.promptBuilder.buildMessages(problemText)
	raw, err := s.provider.ChatComplete(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("generator_service: llm call failed: %w", err)
	}

	mermaid := extractMermaidBlock(raw)
	if mermaid == "" {
		// Fall back to raw response if no code fence found
		mermaid = strings.TrimSpace(raw)
	}

	return &domain.GeneratedUML{MermaidCode: mermaid}, nil
}

// GetConfig returns the currently loaded provider configuration.
func (s *generatorService) GetConfig() domain.GeneratorConfig {
	cfg, _ := s.store.Load()
	return cfg
}

// SaveConfig persists new config and rebuilds the internal provider.
func (s *generatorService) SaveConfig(cfg domain.GeneratorConfig) error {
	cfg.APIEndpoint = strings.TrimSpace(cfg.APIEndpoint)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Model = strings.TrimSpace(cfg.Model)

	log.Printf("[GeneratorService] SaveConfig endpoint=%q (len:%d) model=%q (len:%d)\n",
		cfg.APIEndpoint, len(cfg.APIEndpoint), cfg.Model, len(cfg.Model))

	if err := s.store.Save(cfg); err != nil {
		return err
	}
	// Rebuild the provider with the new settings (hot-reload without restart)
	s.provider = newOpenAICompatibleProvider(cfg.APIEndpoint, cfg.APIKey, cfg.Model)
	return nil
}

// extractMermaidBlock strips the ```mermaid ... ``` fence from the AI response.
// Returns the inner content, or empty string if no fence is found.
func extractMermaidBlock(raw string) string {
	const openFence = "```mermaid"
	const closeFence = "```"

	start := strings.Index(raw, openFence)
	if start == -1 {
		return ""
	}
	inner := raw[start+len(openFence):]
	end := strings.Index(inner, closeFence)
	if end == -1 {
		return strings.TrimSpace(inner)
	}
	return strings.TrimSpace(inner[:end])
}
