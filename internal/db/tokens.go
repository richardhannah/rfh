package db

import (
	"crypto/sha256"
	"fmt"
)

// CreateToken creates a new API token
func (db *DB) CreateToken(tokenHash string, name *string) (*Token, error) {
	query := `
        INSERT INTO tokens (token_hash, name) 
        VALUES ($1, $2) 
        RETURNING id, token_hash, name, created_at`

	var token Token
	err := db.Get(&token, query, tokenHash, name)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// ValidateToken checks if a token exists and is valid
func (db *DB) ValidateToken(tokenHash string) (*Token, error) {
	query := `SELECT id, token_hash, name, created_at FROM tokens WHERE token_hash = $1`

	var token Token
	err := db.Get(&token, query, tokenHash)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// HashToken creates a SHA256 hash of a token with salt
func HashToken(token string, salt string) string {
	h := sha256.New()
	h.Write([]byte(token + salt))
	return fmt.Sprintf("%x", h.Sum(nil))
}