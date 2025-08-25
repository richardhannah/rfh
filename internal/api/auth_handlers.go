package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"rulestack/internal/auth"
	"rulestack/internal/db"
)

// registerHandler handles user registration
func (s *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req db.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Validate username format
	if len(req.Username) < 3 || len(req.Username) > 50 {
		writeError(w, http.StatusBadRequest, "Username must be between 3 and 50 characters")
		return
	}

	// Validate password strength
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	// Default role is user unless specified by admin
	if req.Role == "" {
		req.Role = db.RoleUser
	}

	// Only admins can create accounts with publisher or admin roles
	if req.Role != db.RoleUser {
		user := getUserFromContext(r.Context())
		if user == nil || !user.Role.HasPermission("admin") {
			writeError(w, http.StatusForbidden, "Only admins can create accounts with elevated permissions")
			return
		}
	}

	// Create user
	user, err := s.DB.CreateUser(req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			writeError(w, http.StatusConflict, "Username or email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Return user info (without password hash)
	response := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
	}

	writeJSON(w, http.StatusCreated, response)
}

// loginHandler handles user authentication
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req db.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Get user
	user, err := s.DB.GetUserByUsername(req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Validate password
	if !s.DB.ValidatePassword(user, req.Password) {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token (use development duration for long-lived tokens)
	jwtManager := auth.NewJWTManager(s.Config.JWTSecret, auth.DevelopmentTokenDuration)
	tokenString, tokenHash, expiresAt, err := jwtManager.GenerateToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Store session in database
	userAgent := r.Header.Get("User-Agent")
	ipAddress := getClientIP(r)
	session, err := s.DB.CreateUserSession(user.ID, tokenHash, expiresAt, &userAgent, &ipAddress)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Update last login
	if err := s.DB.UpdateLastLogin(user.ID); err != nil {
		// Log but don't fail
		writeError(w, http.StatusInternalServerError, "Failed to update last login")
		return
	}

	// Return token and user info
	response := map[string]interface{}{
		"token":      tokenString,
		"expires_at": expiresAt,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"session_id": session.ID,
	}

	writeJSON(w, http.StatusOK, response)
}

// logoutHandler handles user logout
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	session := getUserSessionFromContext(r.Context())
	if session != nil {
		// Delete the session
		if _, err := s.DB.Exec(`DELETE FROM user_sessions WHERE id = $1`, session.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to logout")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// changePasswordHandler handles password changes
func (s *Server) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req db.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "Current password and new password are required")
		return
	}

	// Validate password strength
	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "New password must be at least 8 characters long")
		return
	}

	// Verify current password
	if !s.DB.ValidatePassword(user, req.CurrentPassword) {
		writeError(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	// Change password
	if err := s.DB.ChangeUserPassword(user.ID, req.NewPassword); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to change password")
		return
	}

	// Invalidate all existing sessions except current one
	session := getUserSessionFromContext(r.Context())
	query := `DELETE FROM user_sessions WHERE user_id = $1`
	if session != nil {
		query += ` AND id != $2`
		_, _ = s.DB.Exec(query, user.ID, session.ID)
	} else {
		_, _ = s.DB.Exec(query, user.ID)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// deleteAccountHandler handles user account deletion
func (s *Server) deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Prevent admins from deleting themselves
	if user.Role == db.RoleAdmin {
		writeError(w, http.StatusForbidden, "Admin accounts cannot be self-deleted")
		return
	}

	// Delete user account
	if err := s.DB.DeleteUser(user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Account deleted successfully"})
}

// profileHandler returns current user profile
func (s *Server) profileHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	response := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
		"last_login": user.LastLogin,
	}

	writeJSON(w, http.StatusOK, response)
}

// listUsersHandler returns all users (admin only)
func (s *Server) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil || !user.Role.HasPermission("admin") {
		writeError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Parse pagination parameters
	limit := 50 // default
	offset := 0 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get users
	users, err := s.DB.ListUsers(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	// Remove sensitive information
	var response []map[string]interface{}
	for _, u := range users {
		response = append(response, map[string]interface{}{
			"id":         u.ID,
			"username":   u.Username,
			"email":      u.Email,
			"role":       u.Role,
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
			"last_login": u.LastLogin,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

// adminDeleteUserHandler allows admins to delete other users
func (s *Server) adminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil || !user.Role.HasPermission("admin") {
		writeError(w, http.StatusForbidden, "Admin access required")
		return
	}

	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Prevent admins from deleting themselves
	if userID == user.ID {
		writeError(w, http.StatusForbidden, "Cannot delete your own account")
		return
	}

	// Check if target user exists
	targetUser, err := s.DB.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	// Delete user account
	if err := s.DB.DeleteUser(targetUser.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message":      "User deleted successfully",
		"deleted_user": targetUser.Username,
	})
}
