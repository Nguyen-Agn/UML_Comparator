package uml_generator

import (
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
