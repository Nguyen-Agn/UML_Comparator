package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

// aesEncryptor is a private AES-256-GCM implementation used only by this package.
// Binary layout: [ salt(16) | nonce(12) | ciphertext+GCM-tag ]
type aesEncryptor struct {
	secret []byte
}

func newAESEncryptor(secret []byte) iAESEncryptor {
	return &aesEncryptor{secret: secret}
}

func (e *aesEncryptor) encrypt(plaintext []byte) ([]byte, error) {
	// Random salt — makes every encrypted output unique even for the same input
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	key := e.deriveKey(salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Pack: salt || nonce || ciphertext
	out := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

func (e *aesEncryptor) decrypt(data []byte) ([]byte, error) {
	const saltSize, nonceSize = 16, 12
	if len(data) < saltSize+nonceSize+1 {
		return nil, fmt.Errorf("data too short")
	}

	salt := data[:saltSize]
	nonce := data[saltSize : saltSize+nonceSize]
	ciphertext := data[saltSize+nonceSize:]

	key := e.deriveKey(salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("authentication failed — wrong key or corrupted file: %w", err)
	}
	return plaintext, nil
}

func (e *aesEncryptor) deriveKey(salt []byte) []byte {
	h := sha256.New()
	h.Write(e.secret)
	h.Write(salt)
	return h.Sum(nil)
}
