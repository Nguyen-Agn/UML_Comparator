package parser

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"uml_compare/domain"
)

const solutionHeader = "SOLUTION_V1"

// defaultSolutionKey is the built-in key — same constant as cipher.defaultKey.
// Must stay in sync. Both encrypt (cipher pkg) and decrypt (here) use the same value.
const defaultSolutionKey = "uml_comparator_solution_key_v1"

// IDecryptor is the local Strategy interface for decryption.
// (ISP: single method — callers that only need to read don't get encrypt capabilities)
type IDecryptor interface {
	Decrypt(ciphertext []byte) ([]byte, error)
}

// SolutionParser implements IFileParser for encrypted .solution files.
// It delegates ALL crypto work to the injected IDecryptor strategy (SRP + DIP).
//
// The .solution file format:
//
//	SOLUTION_V1\n
//	<Base64( encrypted-payload )>
type SolutionParser struct {
	decryptor IDecryptor
}

// Compile-time interface guarantee.
var _ IFileParser = (*SolutionParser)(nil)

// NewSolutionParser constructs a SolutionParser with the given IDecryptor.
// Use NewSolutionParserDefault() for the zero-config path.
func NewSolutionParser(decryptor IDecryptor) IFileParser {
	return &SolutionParser{decryptor: decryptor}
}

// NewSolutionParserDefault creates a SolutionParser using the built-in default key.
// Key override priority: SOLUTION_KEY env var → built-in default.
// This is the recommended constructor for most use-cases.
func NewSolutionParserDefault() IFileParser {
	key := resolveDefaultKey()
	return NewSolutionParser(newInternalDecryptor(key))
}

// resolveDefaultKey selects the key using the same priority as cipher.resolveKey.
func resolveDefaultKey() []byte {
	if env := os.Getenv("SOLUTION_KEY"); env != "" {
		return []byte(env)
	}
	return []byte(defaultSolutionKey)
}

// Parse reads a .solution file, base64-decodes the payload, delegates
// decryption to the injected IDecryptor, and returns domain.RawModelData
// along with the source type "drawio" (since decrypted data is Draw.io XML).
func (p *SolutionParser) Parse(filePath string) (domain.RawModelData, string, error) {
	if filePath == "" {
		return "", "", fmt.Errorf("SolutionParser.Parse: filePath cannot be empty")
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("SolutionParser.Parse: read file: %w", err)
	}
	content := string(raw)

	if !strings.HasPrefix(content, solutionHeader) {
		return "", "", fmt.Errorf("SolutionParser.Parse: invalid .solution file — missing header")
	}
	encoded := strings.TrimSpace(strings.TrimPrefix(content, solutionHeader))

	packed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("SolutionParser.Parse: base64 decode: %w", err)
	}

	decryptedBytes, err := p.decryptor.Decrypt(packed)
	if err != nil {
		return "", "", fmt.Errorf("SolutionParser.Parse: decrypt: %w", err)
	}

	decryptedStr := string(decryptedBytes)
	trimmed := strings.TrimSpace(decryptedStr)

	// Detect type: if it starts with XML tag like `<`, it's a drawio/XML file.
	if strings.HasPrefix(trimmed, "<") {
		return domain.RawModelData(decryptedBytes), "drawio", nil
	}

	// Otherwise, it is a mermaid DSL file. We should clean it just like MermaidParser does.
	var cleanedLines []string
	scanner := bufio.NewScanner(strings.NewReader(decryptedStr))
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Filter unnecessary information:
		// 1. Skip empty lines.
		// 2. Skip Mermaid comments (starting with %%).
		// 3. Skip header "classDiagram" as it's redundant for the builder.
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "%%") || strings.EqualFold(trimmedLine, "classDiagram") {
			continue
		}

		cleanedLines = append(cleanedLines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("SolutionParser.Parse: scan decrypted mermaid content: %w", err)
	}

	cleanedContent := strings.Join(cleanedLines, "\n")
	return domain.RawModelData(cleanedContent), "mermaid", nil
}
