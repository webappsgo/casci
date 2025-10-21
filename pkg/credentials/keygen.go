package credentials

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/ssh"
)

// KeyGenerator generates cryptographic keys
type KeyGenerator struct{}

// NewKeyGenerator creates a new key generator
func NewKeyGenerator() *KeyGenerator {
	return &KeyGenerator{}
}

// GenerateGPGKey generates a GPG key pair (simulated using RSA)
func (kg *KeyGenerator) GenerateGPGKey(name, email string, keySize int) (*GPGKey, error) {
	if keySize == 0 {
		keySize = 4096
	}

	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Calculate fingerprint
	fingerprint := kg.calculateFingerprint(publicKeyBytes)

	return &GPGKey{
		PublicKey:   string(publicKeyPEM),
		PrivateKey:  string(privateKeyPEM),
		Fingerprint: fingerprint,
		KeyID:       fingerprint[:16],
		UserID:      fmt.Sprintf("%s <%s>", name, email),
	}, nil
}

// GenerateSSHKey generates an SSH key pair
func (kg *KeyGenerator) GenerateSSHKey(keyType string) (*SSHKey, error) {
	switch keyType {
	case "ed25519", "":
		return kg.generateED25519Key()
	case "rsa":
		return kg.generateRSASSHKey(4096)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// generateED25519Key generates an Ed25519 SSH key pair
func (kg *KeyGenerator) generateED25519Key() (*SSHKey, error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
	}

	// Convert to SSH format
	sshPublicKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert public key: %w", err)
	}

	// Encode private key
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: kg.marshalED25519PrivateKey(privateKey),
	})

	// Format public key
	publicKeyString := string(ssh.MarshalAuthorizedKey(sshPublicKey))

	// Calculate fingerprint
	fingerprint := ssh.FingerprintSHA256(sshPublicKey)

	return &SSHKey{
		PublicKey:   publicKeyString,
		PrivateKey:  string(privateKeyPEM),
		Fingerprint: fingerprint,
		KeyType:     "ed25519",
	}, nil
}

// generateRSASSHKey generates an RSA SSH key pair
func (kg *KeyGenerator) generateRSASSHKey(keySize int) (*SSHKey, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Convert to SSH format
	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert public key: %w", err)
	}

	// Encode private key
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Format public key
	publicKeyString := string(ssh.MarshalAuthorizedKey(sshPublicKey))

	// Calculate fingerprint
	fingerprint := ssh.FingerprintSHA256(sshPublicKey)

	return &SSHKey{
		PublicKey:   publicKeyString,
		PrivateKey:  string(privateKeyPEM),
		Fingerprint: fingerprint,
		KeyType:     "rsa",
	}, nil
}

// GenerateSigningCertificate generates a self-signed code signing certificate
func (kg *KeyGenerator) GenerateSigningCertificate(name, organization string) (*SigningCertificate, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // 1 year validity

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{organization},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// Create self-signed certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// Encode private key
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Calculate fingerprint
	fingerprint := kg.calculateFingerprint(certBytes)

	return &SigningCertificate{
		Certificate: string(certPEM),
		PrivateKey:  string(privateKeyPEM),
		Chain:       []string{},
		Fingerprint: fingerprint,
		Subject:     name,
		Issuer:      name, // Self-signed
		ValidFrom:   notBefore,
		ValidTo:     notAfter,
	}, nil
}

// calculateFingerprint calculates SHA-256 fingerprint
func (kg *KeyGenerator) calculateFingerprint(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// marshalED25519PrivateKey marshals an Ed25519 private key
func (kg *KeyGenerator) marshalED25519PrivateKey(privateKey ed25519.PrivateKey) []byte {
	// Simple marshaling (in production, use proper OpenSSH format)
	return privateKey
}

// ParseSSHPublicKey parses an SSH public key and returns its fingerprint
func ParseSSHPublicKey(publicKey string) (string, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse SSH public key: %w", err)
	}

	return ssh.FingerprintSHA256(key), nil
}

// ParseGPGPublicKey parses a GPG public key and returns its fingerprint
func ParseGPGPublicKey(publicKey string) (string, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	hash := sha256.Sum256(block.Bytes)
	return fmt.Sprintf("%x", hash), nil
}
