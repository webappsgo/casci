package users

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// AuthManager handles authentication and JWT tokens
type AuthManager struct {
	jwtSecret []byte
}

// NewAuthManager creates a new auth manager
func NewAuthManager() *AuthManager {
	// Generate a random JWT secret
	secret := make([]byte, 32)
	rand.Read(secret)

	return &AuthManager{
		jwtSecret: secret,
	}
}

// NewAuthManagerWithSecret creates a new auth manager with a specific secret
func NewAuthManagerWithSecret(secret string) *AuthManager {
	return &AuthManager{
		jwtSecret: []byte(secret),
	}
}

// GenerateToken generates a JWT token for a user
func (am *AuthManager) GenerateToken(user *User) (string, error) {
	// Token expires in 24 hours
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "casci",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns claims
func (am *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check expiration
		if claims.ExpiresAt.Before(time.Now()) {
			return nil, ErrTokenExpired
		}
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshToken generates a new token from an existing valid token
func (am *AuthManager) RefreshToken(tokenString string) (string, error) {
	claims, err := am.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Create a new token with updated expiration
	newClaims := &Claims{
		UserID:   claims.UserID,
		Username: claims.Username,
		IsAdmin:  claims.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "casci",
			Subject:   claims.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err = token.SignedString(am.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateSessionID generates a random session ID
func (am *AuthManager) GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
