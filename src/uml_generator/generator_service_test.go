package uml_generator

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockProvider is a test double for iLLMProvider.
type mockProvider struct {
	response string
	err      error
}

// ChatComplete returns the preset response or error, ignoring the messages.
func (m *mockProvider) ChatComplete(_ context.Context, _ []llmMessage) (string, error) {
	return m.response, m.err
}

func newTestService(provider iLLMProvider) *generatorService {
	store, _ := newConfigStore()
	return &generatorService{
		provider:      provider,
		promptBuilder: newPromptBuilder(),
		store:         store,
	}
}

func TestGenerate_ReturnsMermaidCode(t *testing.T) {
	expected := "classDiagram\n    A --> B : __1__"
	svc := newTestService(&mockProvider{
		response: "```mermaid\n" + expected + "\n```",
	})

	result, err := svc.Generate(context.Background(), "Design a simple system with A and B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.MermaidCode, "classDiagram") {
		t.Errorf("expected classDiagram in output, got: %q", result.MermaidCode)
	}
}

func TestGenerate_PropagatesProviderError(t *testing.T) {
	provErr := errors.New("connection refused")
	svc := newTestService(&mockProvider{err: provErr})

	_, err := svc.Generate(context.Background(), "some problem")
	if err == nil {
		t.Fatal("expected error to be propagated, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected original error in message, got: %v", err)
	}
}

func TestGenerate_RejectsEmptyProblem(t *testing.T) {
	svc := newTestService(&mockProvider{response: "anything"})

	_, err := svc.Generate(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error for empty problem text, got nil")
	}
}

func TestGenerate_FallbackWhenNoFence(t *testing.T) {
	// AI returns code without a code fence — should still work
	rawCode := "classDiagram\n    X --> Y : __1__"
	svc := newTestService(&mockProvider{response: rawCode})

	result, err := svc.Generate(context.Background(), "design X and Y")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MermaidCode == "" {
		t.Error("expected non-empty mermaid code in fallback case")
	}
}

func TestExtractMermaidBlock(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "with fence",
			input: "Sure!\n```mermaid\nclassDiagram\n    A-->B\n```\nDone.",
			want:  "classDiagram\n    A-->B",
		},
		{
			name:  "no fence",
			input: "classDiagram\n    A-->B",
			want:  "",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractMermaidBlock(tc.input)
			if got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
