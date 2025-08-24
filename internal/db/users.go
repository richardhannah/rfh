package db

import (
	"database/sql/driver"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserRole represents user access levels
type UserRole string

const (
	RoleUser      UserRole = "user"
	RolePublisher UserRole = "publisher"
	RoleAdmin     UserRole = "admin"
)

// Value implements the driver.Valuer interface for database storage
func (r UserRole) Value() (driver.Value, error) {
	return string(r), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = RoleUser
		return nil
	}
	if str, ok := value.(string); ok {
		*r = UserRole(str)
		return nil
	}
	return errors.New("cannot scan UserRole")
}

// User represents a user account
type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         UserRole  `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	LastLogin    *time.Time `json:"last_login" db:"last_login"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// UserSession represents a user authentication session
type UserSession struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	LastUsed  time.Time `json:"last_used" db:"last_used"`
	UserAgent *string   `json:"user_agent" db:"user_agent"`
	IPAddress *string   `json:"ip_address" db:"ip_address"`
}

// CreateUserRequest represents user registration data
type CreateUserRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Role     UserRole `json:"role"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest represents password change data
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// CreateUser creates a new user account
func (db *DB) CreateUser(req CreateUserRequest) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, password_hash, role, created_at, updated_at, last_login, is_active`

	var user User
	err = db.Get(&user, query, req.Username, req.Email, string(hashedPassword), req.Role)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (db *DB) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at, last_login, is_active
		FROM users 
		WHERE username = $1 AND is_active = true`

	var user User
	err := db.Get(&user, query, username)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(id int) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at, last_login, is_active
		FROM users 
		WHERE id = $1 AND is_active = true`

	var user User
	err := db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ValidatePassword checks if the provided password matches the user's password
func (db *DB) ValidatePassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// CreateUserSession creates a new user session
func (db *DB) CreateUserSession(userID int, tokenHash string, expiresAt time.Time, userAgent, ipAddress *string) (*UserSession, error) {
	query := `
		INSERT INTO user_sessions (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, token_hash, expires_at, created_at, last_used, user_agent, ip_address`

	var session UserSession
	err := db.Get(&session, query, userID, tokenHash, expiresAt, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// ValidateUserSession validates a session token and returns user info
func (db *DB) ValidateUserSession(tokenHash string) (*User, *UserSession, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.role, u.created_at, u.updated_at, u.last_login, u.is_active,
		       s.id, s.user_id, s.token_hash, s.expires_at, s.created_at, s.last_used, s.user_agent, s.ip_address
		FROM users u
		JOIN user_sessions s ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now() AND u.is_active = true`

	rows, err := db.Query(query, tokenHash)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil, errors.New("invalid or expired session")
	}

	var user User
	var session UserSession

	err = rows.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.IsActive,
		&session.ID, &session.UserID, &session.TokenHash, &session.ExpiresAt,
		&session.CreatedAt, &session.LastUsed, &session.UserAgent, &session.IPAddress,
	)
	if err != nil {
		return nil, nil, err
	}

	return &user, &session, nil
}

// UpdateLastLogin updates the user's last login timestamp
func (db *DB) UpdateLastLogin(userID int) error {
	query := `UPDATE users SET last_login = now() WHERE id = $1`
	_, err := db.Exec(query, userID)
	return err
}

// UpdateSessionLastUsed updates the session's last used timestamp
func (db *DB) UpdateSessionLastUsed(sessionID int) error {
	query := `UPDATE user_sessions SET last_used = now() WHERE id = $1`
	_, err := db.Exec(query, sessionID)
	return err
}

// ChangeUserPassword changes a user's password
func (db *DB) ChangeUserPassword(userID int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`
	_, err = db.Exec(query, string(hashedPassword), userID)
	return err
}

// DeleteUser soft deletes a user account
func (db *DB) DeleteUser(userID int) error {
	// Start transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Soft delete user
	_, err = tx.Exec(`UPDATE users SET is_active = false, updated_at = now() WHERE id = $1`, userID)
	if err != nil {
		return err
	}

	// Delete all user sessions
	_, err = tx.Exec(`DELETE FROM user_sessions WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Delete old API tokens
	_, err = tx.Exec(`DELETE FROM tokens WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CleanupExpiredSessions removes expired sessions from the database
func (db *DB) CleanupExpiredSessions() error {
	query := `DELETE FROM user_sessions WHERE expires_at <= now()`
	_, err := db.Exec(query)
	return err
}

// ListUsers returns all active users (admin function)
func (db *DB) ListUsers(limit, offset int) ([]User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at, last_login, is_active
		FROM users 
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	var users []User
	err := db.Select(&users, query, limit, offset)
	return users, err
}

// HasPermission checks if a user role has permission for a specific action
func (r UserRole) HasPermission(action string) bool {
	switch action {
	case "read":
		return r == RoleUser || r == RolePublisher || r == RoleAdmin
	case "publish":
		return r == RolePublisher || r == RoleAdmin
	case "admin":
		return r == RoleAdmin
	default:
		return false
	}
}