package credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Encryption parameters
	saltSize   = 32
	keySize    = 32 // AES-256
	iterations = 100000
)

// Encryptor handles credential encryption and decryption
type Encryptor struct {
	masterKey []byte
}

// NewEncryptor creates a new encryptor with a master key
func NewEncryptor(masterKey string) *Encryptor {
	// Derive a proper key from the master key
	derived := pbkdf2.Key([]byte(masterKey), []byte("casci-credentials"), iterations, keySize, sha256.New)
	return &Encryptor{
		masterKey: derived,
	}
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create cipher
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", err
	}

	// Use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and append nonce
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// Create cipher
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", err
	}

	// Use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptWithUserKey encrypts data with a user-specific key
func (e *Encryptor) EncryptWithUserKey(plaintext string, userID int) (string, error) {
	// Derive user-specific key
	userKey := e.deriveUserKey(userID)
	userEncryptor := &Encryptor{masterKey: userKey}
	return userEncryptor.Encrypt(plaintext)
}

// DecryptWithUserKey decrypts data with a user-specific key
func (e *Encryptor) DecryptWithUserKey(ciphertext string, userID int) (string, error) {
	// Derive user-specific key
	userKey := e.deriveUserKey(userID)
	userEncryptor := &Encryptor{masterKey: userKey}
	return userEncryptor.Decrypt(ciphertext)
}

// deriveUserKey derives a user-specific encryption key
func (e *Encryptor) deriveUserKey(userID int) []byte {
	// Combine master key with user ID
	userSalt := []byte("user-" + string(rune(userID)))
	return pbkdf2.Key(e.masterKey, userSalt, iterations, keySize, sha256.New)
}

// GenerateMasterKey generates a random master key
func GenerateMasterKey() (string, error) {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// HashPassword hashes a password for secure storage
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}
