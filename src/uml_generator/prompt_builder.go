package uml_generator

import (
	"fmt"
	"strings"
)

// promptBuilder builds chat messages for the AI generation call.
// The system message is derived from systemPromptTemplate defined in templates.go.
type promptBuilder struct {
	systemPrompt string
}

// newPromptBuilder constructs a promptBuilder using the embedded template.
func newPromptBuilder() *promptBuilder {
	return &promptBuilder{systemPrompt: buildSystemPrompt(systemPromptTemplate)}
}

// buildMessages returns the [system, user] message pair for a single-shot LLM call.
// system: project syntax rules from template.md
// user:   the raw problem description provided by the instructor
func (b *promptBuilder) buildMessages(problemText string) []llmMessage {
	return []llmMessage{
		{Role: "system", Content: b.systemPrompt},
		{Role: "user", Content: problemText},
	}
}

// buildSystemPrompt extracts the instructional content from template.md and wraps it
// in a directive that constrains the model's output format.
func buildSystemPrompt(tmpl string) string {
	// Strip markdown code fences if present so the AI sees the rules, not the rendering
	cleaned := stripCodeFence(tmpl)
	return fmt.Sprintf(
		"You are a UML class diagram expert for a Vietnamese university grading system.\n"+
			"Generate a Mermaid classDiagram following EXACTLY the syntax format below.\n\n"+
			"--- FORMAT REFERENCE ---\n%s\n--- END FORMAT REFERENCE ---\n\n"+
			"STRICT RULES:\n"+
			"1. Return ONLY the Mermaid code block (```mermaid ... ```). No explanation, no prose.\n"+
			"2. Every class, attribute, method, and relationship MUST have a score tag __1__ (default score).\n"+
			"3. Use ~T~ instead of <T> for generics.\n"+
			"4. Use | to express alternative names or types where multiple answers are acceptable.\n"+
			"5. Infer all classes, attributes, methods, and relationships from the problem description.\n",
		cleaned,
	)
}

// stripCodeFence removes leading/trailing markdown code fences (``` or ```mermaid).
func stripCodeFence(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	inFence := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		result = append(result, line)
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}
