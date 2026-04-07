package parser

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
)

// internalDecryptor is a private AES-256-GCM IDecryptor used only inside
// the parser package. It mirrors the layout produced by cipher/aes.go:
//
//	[ salt(16) | nonce(12) | ciphertext + GCM tag ]
type internalDecryptor struct {
	secret []byte
}

// newInternalDecryptor creates an IDecryptor with the given key bytes.
// Called by NewSolutionParserDefault() — not part of the public API.
func newInternalDecryptor(secret []byte) IDecryptor {
	return &internalDecryptor{secret: secret}
}

// Decrypt unpacks the binary blob and decrypts with AES-256-GCM.
func (d *internalDecryptor) Decrypt(data []byte) ([]byte, error) {
	const saltSize, nonceSize = 16, 12
	const minSize = saltSize + nonceSize + 1

	if len(data) < minSize {
		return nil, fmt.Errorf("internalDecryptor.Decrypt: data too short (%d bytes)", len(data))
	}

	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	key := d.deriveKey(salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("internalDecryptor.Decrypt: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("internalDecryptor.Decrypt: create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("internalDecryptor.Decrypt: auth failed — wrong key or corrupted file: %w", err)
	}

	return plaintext, nil
}

// deriveKey produces a 32-byte AES key via SHA-256(secret + salt).
// Must match the derivation used in cipher/aes.go exactly.
func (d *internalDecryptor) deriveKey(salt []byte) []byte {
	h := sha256.New()
	h.Write(d.secret)
	h.Write(salt)
	return h.Sum(nil)
}
