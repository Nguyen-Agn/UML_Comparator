package uml_generator

import (
	"strings"
	"testing"
)

func TestBuildMessages_ReturnsTwoMessages(t *testing.T) {
	pb := newPromptBuilder()
	msgs := pb.buildMessages("Design a library system")

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "system" {
		t.Errorf("first message role: want 'system', got %q", msgs[0].Role)
	}
	if msgs[1].Role != "user" {
		t.Errorf("second message role: want 'user', got %q", msgs[1].Role)
	}
}

func TestBuildMessages_UserContentMatchesInput(t *testing.T) {
	pb := newPromptBuilder()
	problem := "Students enroll in Courses taught by Professors"
	msgs := pb.buildMessages(problem)
	if msgs[1].Content != problem {
		t.Errorf("user content: want %q, got %q", problem, msgs[1].Content)
	}
}

func TestBuildSystemPrompt_ContainsSyntaxKeywords(t *testing.T) {
	pb := newPromptBuilder()
	system := pb.systemPrompt

	keywords := []string{"classDiagram", "__1__", "classDiagram", "|", "~T~", "score"}
	for _, kw := range keywords {
		if !strings.Contains(system, kw) {
			t.Errorf("system prompt missing keyword %q", kw)
		}
	}
}

func TestStripCodeFence_RemovesFence(t *testing.T) {
	input := "```mermaid\nclassDiagram\n    A --> B\n```\nSome note"
	got := stripCodeFence(input)
	if strings.Contains(got, "```") {
		t.Errorf("expected code fences to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "classDiagram") {
		t.Errorf("expected classDiagram to be preserved, got:\n%s", got)
	}
}

func TestStripCodeFence_NoFence_Unchanged(t *testing.T) {
	input := "classDiagram\n    A --> B"
	got := stripCodeFence(input)
	if !strings.Contains(got, "classDiagram") {
		t.Errorf("content should be preserved when no fence, got:\n%s", got)
	}
}
