package uml_generator

// promptBuilder builds chat messages for the AI generation call.
// The system message is the complete systemPromptTemplate from templates.go.
type promptBuilder struct {
	systemPrompt string
}

// newPromptBuilder constructs a promptBuilder using the embedded template.
func newPromptBuilder() *promptBuilder {
	return &promptBuilder{systemPrompt: systemPromptTemplate}
}

// buildMessages returns the [system, user] message pair for a single-shot LLM call.
// system: syntax rules and output constraints from templates.go
// user:   the raw problem description provided by the instructor
func (b *promptBuilder) buildMessages(problemText string) []llmMessage {
	return []llmMessage{
		{Role: "system", Content: b.systemPrompt},
		{Role: "user", Content: problemText},
	}
}
