// Package cipher provides the optional solution-file encryption feature.
//
// Architecture note:
//   - Students use this package PASSIVELY via AutoParser (transparent decryption)
//   - Teachers use it ACTIVELY via ISolutionCipher.Encrypt() to produce .solution files
//   - The feature is completely optional: the rest of the system works identically
//     whether input is a .drawio or a .solution file.
package cipher

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"uml_compare/src/parser"
)

// ─── Default key ──────────────────────────────────────────────────────────────
// A built-in key that works out-of-the-box for both teacher (encrypt) and
// student (decrypt) without any configuration.
//
// Override priority (highest → lowest):
//  1. Key passed explicitly to NewWithKey(key)
//  2. SOLUTION_KEY environment variable
//  3. defaultKey constant below
const defaultKey = "uml_comparator_solution_key_v1"

// solutionHeader is the magic bytes written at the top of every .solution file.
const solutionHeader = "SOLUTION_V1\n"

// ─── Interface ────────────────────────────────────────────────────────────────

// ISolutionCipher is the optional encryption interface for teachers.
//
// Students never call this. The GUI/CLI can expose it as an optional menu item.
// To add a new algorithm, implement this interface and inject it — nothing else changes.
type ISolutionCipher interface {
	// Encrypt reads a .drawio file and writes an AES-encrypted .solution file.
	Encrypt(inputPath, outputPath string) error
}

// ─── Constructor ──────────────────────────────────────────────────────────────

// New returns an ISolutionCipher using the built-in default key.
// This is the zero-config path suitable for most use-cases.
//
//	c := cipher.New()
//	c.Encrypt("exam.drawio", "exam.solution")
func New() ISolutionCipher {
	return NewWithKey(nil)
}

// NewWithKey returns an ISolutionCipher with an explicit key.
// Pass nil or empty slice to fall back to env var then the built-in default.
//
//	c := cipher.NewWithKey([]byte("custom-secret"))
func NewWithKey(key []byte) ISolutionCipher {
	resolved := resolveKey(key)
	return &solutionCipher{
		drawioPaser: parser.NewDrawioParser(),
		encryptor:   newAESEncryptor(resolved),
	}
}

// resolveKey picks the first non-empty value in priority order.
func resolveKey(explicit []byte) []byte {
	if len(explicit) > 0 {
		return explicit
	}
	if env := os.Getenv("SOLUTION_KEY"); env != "" {
		return []byte(env)
	}
	return []byte(defaultKey)
}

// ─── Concrete implementation ──────────────────────────────────────────────────

type solutionCipher struct {
	drawioPaser parser.IFileParser // DIP: interface, not *DrawioParser
	encryptor   iAESEncryptor
}

// iAESEncryptor is an unexported interface — internal use only (ISP).
type iAESEncryptor interface {
	encrypt(plaintext []byte) ([]byte, error)
}

// Encrypt reads inputPath (.drawio), encrypts it, and writes to outputPath (.solution).
// If outputPath is empty, it defaults to <inputPath without .drawio> + ".solution".
func (c *solutionCipher) Encrypt(inputPath, outputPath string) error {
	if inputPath == "" {
		return fmt.Errorf("cipher.Encrypt: inputPath cannot be empty")
	}

	// Auto-derive output path if not specified
	if outputPath == "" {
		outputPath = strings.TrimSuffix(inputPath, ".drawio") + ".solution"
	}

	// Step 1: Parse .drawio → raw XML
	rawXML, _, err := c.drawioPaser.Parse(inputPath)
	if err != nil {
		return fmt.Errorf("cipher.Encrypt: parse drawio: %w", err)
	}
	// RawModelData is a named string type — cast explicitly for strings.TrimSpace
	if strings.TrimSpace(string(rawXML)) == "" {
		return fmt.Errorf("cipher.Encrypt: empty diagram — invalid .drawio file")
	}

	// Step 2: Encrypt using AES-256-GCM
	cipherBytes, err := c.encryptor.encrypt([]byte(rawXML))
	if err != nil {
		return fmt.Errorf("cipher.Encrypt: encrypt: %w", err)
	}

	// Step 3: Write SOLUTION_V1 header + Base64(ciphertext) to output file
	content := solutionHeader + base64.StdEncoding.EncodeToString(cipherBytes)
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("cipher.Encrypt: write output: %w", err)
	}

	return nil
}
