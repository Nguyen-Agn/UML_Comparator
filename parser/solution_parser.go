package parser

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"uml_compare/domain"
)

const solutionHeader = "SOLUTION_V1\n"

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

	xmlBytes, err := p.decryptor.Decrypt(packed)
	if err != nil {
		return "", "", fmt.Errorf("SolutionParser.Parse: decrypt: %w", err)
	}

	return domain.RawModelData(xmlBytes), "drawio", nil
}
