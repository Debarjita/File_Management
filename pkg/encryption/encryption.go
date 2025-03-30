// pkg/encryption/encryption.go
package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"io/ioutil"
)

// Encryption handles encryption/decryption of files
type Encryption struct {
	key []byte
}

// NewEncryption creates a new encryption service with the given key
func NewEncryption(key []byte) (*Encryption, error) {
	if len(key) != 32 {
		return nil, errors.New("encryption key must be 32 bytes for AES-256")
	}
	return &Encryption{key: key}, nil
}

// EncryptFile encrypts a file and returns a reader to the encrypted content
func (e *Encryption) EncryptFile(src io.Reader) (io.Reader, error) {
	// Read the entire file into memory (for simplicity)
	// In a production system, you might want to use streaming encryption
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	// Create a random nonce
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Encrypt the data
	ciphertext := aesgcm.Seal(nil, nonce, data, nil)

	// Prepend nonce to ciphertext
	result := append(nonce, ciphertext...)

	// Return a reader to the encrypted data
	return io.NopCloser(io.MultiReader(bytes.NewReader(result))), nil
}

// DecryptFile decrypts a file and returns a reader to the decrypted content
func (e *Encryption) DecryptFile(src io.Reader) (io.Reader, error) {
	// Read the entire file into memory
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	// Check if file is large enough to contain nonce
	if len(data) < 12 {
		return nil, errors.New("encrypted file too short")
	}

	// Extract nonce and ciphertext
	nonce := data[:12]
	ciphertext := data[12:]

	// Create a new AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt the data
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Return a reader to the decrypted data
	return io.NopCloser(bytes.NewReader(plaintext)), nil
}
