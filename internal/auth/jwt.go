package auth

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"rulestack/internal/db"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	Role     db.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token creation and validation
type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

// GenerateToken generates a new JWT token for a user
func (j *JWTManager) GenerateToken(user *db.User) (string, string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(j.tokenDuration)

	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   fmt.Sprintf("%d", user.ID),
			Issuer:    "rulestack-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Create hash for database storage
	tokenHash := j.hashToken(tokenString)

	return tokenString, tokenHash, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// HashToken creates a SHA256 hash of the token for database storage
func (j *JWTManager) hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// GetTokenHash returns the hash of a token string
func (j *JWTManager) GetTokenHash(tokenString string) string {
	return j.hashToken(tokenString)
}

// DefaultTokenDuration is the default token expiration time
const DefaultTokenDuration = 24 * time.Hour

// DevelopmentTokenDuration is used during development (effectively no expiration)
const DevelopmentTokenDuration = 365 * 24 * time.Hour // 1 year