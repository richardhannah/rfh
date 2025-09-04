# Admin User Role Management Endpoint Implementation Plan

## Overview

This plan outlines the implementation of a PUT endpoint for changing user roles, accessible only to admin and root users. This endpoint is needed to support the enhanced Cucumber testing World class for creating test users with specific roles.

## Endpoint Specification

### Route
```
PUT /v1/admin/users/{username}/role
```

### Authentication Requirements
- **Required Roles**: `admin` or `root`
- **Authentication**: JWT token via `Authorization: Bearer <token>` header
- **Authorization**: Uses existing `HasPermission("admin")` check

### Request Format
```json
{
  "role": "user|publisher|admin"
}
```

### Response Formats

#### Success (200 OK)
```json
{
  "username": "testuser",
  "role": "publisher", 
  "updated_at": "2025-08-30T20:30:00Z"
}
```

#### Error Responses
```json
// 400 Bad Request - Invalid role
{
  "error": "Invalid role. Must be one of: user, publisher, admin"
}

// 403 Forbidden - Insufficient permissions
{
  "error": "Insufficient permissions"
}

// 404 Not Found - User doesn't exist
{
  "error": "User not found"
}

// 409 Conflict - Cannot change root user role
{
  "error": "Cannot modify root user role"
}
```

## Implementation Plan

### Phase 1: Database Layer Updates

#### 1.1 Add User Role Update Method
**File**: `internal/db/users.go`

```go
// UpdateUserRole updates a user's role (admin/root only)
func (db *DB) UpdateUserRole(username string, newRole UserRole) (*User, error) {
    // Prevent changing root user role
    if username == "root" {
        return nil, errors.New("cannot modify root user role")
    }
    
    // Validate role
    validRoles := []UserRole{RoleUser, RolePublisher, RoleAdmin}
    isValidRole := false
    for _, validRole := range validRoles {
        if newRole == validRole {
            isValidRole = true
            break
        }
    }
    if !isValidRole {
        return nil, errors.New("invalid role")
    }
    
    // Update user role
    query := `
        UPDATE users 
        SET role = $1, updated_at = CURRENT_TIMESTAMP 
        WHERE username = $2 
        RETURNING id, username, email, role, created_at, updated_at, last_login, is_active
    `
    
    var user User
    err := db.QueryRow(query, newRole, username).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.Role,
        &user.CreatedAt,
        &user.UpdatedAt,
        &user.LastLogin,
        &user.IsActive,
    )
    
    if err == sql.ErrNoRows {
        return nil, errors.New("user not found")
    }
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

// GetUserByUsername retrieves a user by username (needed for validation)
func (db *DB) GetUserByUsername(username string) (*User, error) {
    query := `
        SELECT id, username, email, role, created_at, updated_at, last_login, is_active 
        FROM users 
        WHERE username = $1
    `
    
    var user User
    err := db.QueryRow(query, username).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.Role,
        &user.CreatedAt,
        &user.UpdatedAt,
        &user.LastLogin,
        &user.IsActive,
    )
    
    if err == sql.ErrNoRows {
        return nil, errors.New("user not found")
    }
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

### Phase 2: API Handler Implementation

#### 2.1 Create Admin Handlers File
**File**: `internal/api/admin_handlers.go` (new file)

```go
package api

import (
    "encoding/json"
    "net/http"
    "strings"

    "github.com/gorilla/mux"
    "rulestack/internal/db"
)

// UpdateUserRoleRequest represents the request body for updating user roles
type UpdateUserRoleRequest struct {
    Role db.UserRole `json:"role"`
}

// UpdateUserRoleResponse represents the response for role update
type UpdateUserRoleResponse struct {
    Username  string    `json:"username"`
    Role      db.UserRole `json:"role"`
    UpdatedAt string    `json:"updated_at"`
}

// updateUserRoleHandler handles PUT /v1/admin/users/{username}/role
func (s *Server) updateUserRoleHandler(w http.ResponseWriter, r *http.Request) {
    // Extract username from URL path
    vars := mux.Vars(r)
    username := vars["username"]
    
    if username == "" {
        writeError(w, http.StatusBadRequest, "Username is required")
        return
    }
    
    // Parse request body
    var req UpdateUserRoleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    // Validate role value
    if req.Role == "" {
        writeError(w, http.StatusBadRequest, "Role is required")
        return
    }
    
    // Check if user exists first (better error message)
    existingUser, err := s.DB.GetUserByUsername(username)
    if err != nil {
        if strings.Contains(err.Error(), "user not found") {
            writeError(w, http.StatusNotFound, "User not found")
        } else {
            writeError(w, http.StatusInternalServerError, "Failed to check user")
        }
        return
    }
    
    // Prevent changing root user role
    if existingUser.Role == db.RoleRoot {
        writeError(w, http.StatusConflict, "Cannot modify root user role")
        return
    }
    
    // Update user role
    updatedUser, err := s.DB.UpdateUserRole(username, req.Role)
    if err != nil {
        if strings.Contains(err.Error(), "invalid role") {
            writeError(w, http.StatusBadRequest, "Invalid role. Must be one of: user, publisher, admin")
        } else if strings.Contains(err.Error(), "user not found") {
            writeError(w, http.StatusNotFound, "User not found")
        } else if strings.Contains(err.Error(), "cannot modify root user") {
            writeError(w, http.StatusConflict, "Cannot modify root user role")
        } else {
            writeError(w, http.StatusInternalServerError, "Failed to update user role")
        }
        return
    }
    
    // Return success response
    response := UpdateUserRoleResponse{
        Username:  updatedUser.Username,
        Role:      updatedUser.Role,
        UpdatedAt: updatedUser.UpdatedAt.Format("2006-01-02T15:04:05Z"),
    }
    
    writeJSON(w, http.StatusOK, response)
}
```

### Phase 3: Route Registration

#### 3.1 Update Routes Configuration
**File**: `internal/api/routes.go`

```go
// Add to RegisterRoutes function
func RegisterRoutes(r *mux.Router, database *db.DB, cfg config.Config) {
    // ... existing routes ...
    
    // Admin routes - require admin or root role
    adminRouter := r.PathPrefix("/v1/admin").Subrouter()
    
    // User management endpoints
    adminRouter.HandleFunc("/users/{username}/role", s.updateUserRoleHandler).Methods("PUT")
    
    // ... rest of existing code ...
}
```

#### 3.2 Update Route Registry Metadata
**File**: `internal/api/routes.go` (in the route registration section)

```go
// Add route metadata for admin endpoints
registry.RegisterRoute("/v1/admin/users/{username}/role", "PUT", RouteMetadata{
    RequiresAuthentication: true,
    RequiredRole:          "admin", // admin or root can access
    RateLimit:            10,       // 10 requests per minute
    Description:          "Update user role",
})
```

### Phase 4: Security Enhancements

#### 4.1 Enhanced Role Validation
Add validation to prevent privilege escalation:

```go
// In updateUserRoleHandler, add additional checks
func (s *Server) updateUserRoleHandler(w http.ResponseWriter, r *http.Request) {
    // ... existing code ...
    
    // Get current user from context (authenticated user making the request)
    currentUser := getUserFromContext(r.Context())
    if currentUser == nil {
        writeError(w, http.StatusUnauthorized, "Authentication required")
        return
    }
    
    // Prevent non-root users from creating admin users
    if req.Role == db.RoleAdmin && currentUser.Role != db.RoleRoot {
        writeError(w, http.StatusForbidden, "Only root users can create admin users")
        return
    }
    
    // ... rest of existing code ...
}
```

### Phase 5: Logging and Audit Trail

#### 5.1 Add Audit Logging
```go
// In updateUserRoleHandler, add logging
func (s *Server) updateUserRoleHandler(w http.ResponseWriter, r *http.Request) {
    // ... after successful role update ...
    
    // Log the role change for audit purposes
    log.Printf("User role updated: %s changed %s from %s to %s", 
        currentUser.Username, 
        username, 
        existingUser.Role, 
        updatedUser.Role,
    )
    
    // ... rest of code ...
}
```

### Phase 6: Testing Implementation

#### 6.1 Unit Tests
**File**: `internal/api/admin_handlers_test.go` (new file)

```go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "rulestack/internal/db"
)

func TestUpdateUserRole(t *testing.T) {
    // Test cases
    tests := []struct {
        name           string
        username       string
        requestBody    UpdateUserRoleRequest
        currentUser    *db.User
        expectedStatus int
        expectedError  string
    }{
        {
            name:     "successful role update by admin",
            username: "testuser",
            requestBody: UpdateUserRoleRequest{
                Role: db.RolePublisher,
            },
            currentUser: &db.User{
                Username: "admin",
                Role:     db.RoleAdmin,
            },
            expectedStatus: http.StatusOK,
        },
        {
            name:     "prevent non-admin from changing roles",
            username: "testuser",
            requestBody: UpdateUserRoleRequest{
                Role: db.RolePublisher,
            },
            currentUser: &db.User{
                Username: "user",
                Role:     db.RoleUser,
            },
            expectedStatus: http.StatusForbidden,
            expectedError:  "Insufficient permissions",
        },
        {
            name:     "prevent changing root user role",
            username: "root",
            requestBody: UpdateUserRoleRequest{
                Role: db.RoleUser,
            },
            currentUser: &db.User{
                Username: "admin",
                Role:     db.RoleAdmin,
            },
            expectedStatus: http.StatusConflict,
            expectedError:  "Cannot modify root user role",
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Implementation of individual test cases
            // ... test setup and assertions ...
        })
    }
}
```

#### 6.2 Integration Tests
**File**: Add to existing cucumber tests

```gherkin
Feature: Admin User Management
  As an admin user
  I want to manage user roles
  So that I can control user permissions

  Background:
    Given I am logged in as an admin

  Scenario: Change user role to publisher
    Given user "testuser" exists with role "user"
    When I change user "testuser" role to "publisher"
    Then the user should have role "publisher"
    And I should see "Role updated successfully"

  Scenario: Prevent non-admin from changing roles
    Given I am logged in as a user
    When I try to change user "testuser" role to "publisher"  
    Then I should see "Insufficient permissions"
    And the request should fail with status 403

  Scenario: Prevent changing root user role
    Given I am logged in as an admin
    When I try to change user "root" role to "user"
    Then I should see "Cannot modify root user role"
    And the request should fail with status 409
```

### Phase 7: Documentation

#### 7.1 API Documentation
Add to API documentation:

```markdown
## Admin Endpoints

### Update User Role
Updates a user's role. Requires admin or root authentication.

**Endpoint**: `PUT /v1/admin/users/{username}/role`

**Authentication**: Required (admin or root)

**Request Body**:
```json
{
  "role": "user|publisher|admin"
}
```

**Responses**:
- `200 OK`: Role updated successfully
- `400 Bad Request`: Invalid request or role
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: User not found
- `409 Conflict`: Cannot modify root user
```

## Implementation Timeline

### Week 1: Core Implementation
- [ ] Implement database layer methods (`UpdateUserRole`, `GetUserByUsername`)
- [ ] Create admin handlers file with role update endpoint
- [ ] Add route registration and metadata
- [ ] Basic error handling

### Week 2: Security & Validation  
- [ ] Add privilege escalation protection
- [ ] Implement root user protection
- [ ] Add comprehensive input validation
- [ ] Add audit logging

### Week 3: Testing
- [ ] Write unit tests for database methods
- [ ] Write unit tests for API handlers
- [ ] Add integration tests (Cucumber)
- [ ] Test security scenarios

### Week 4: Documentation & Integration
- [ ] Update API documentation
- [ ] Integration with Enhanced World class
- [ ] End-to-end testing
- [ ] Performance testing

## Security Considerations

### 1. **Authorization Checks**
- Only admin and root users can change roles
- Prevent privilege escalation (only root can create admins)
- Protect root user from role changes

### 2. **Input Validation**
- Validate role values against enum
- Sanitize username input
- Prevent injection attacks

### 3. **Audit Trail**
- Log all role changes with timestamps
- Include actor and target user information
- Monitor for suspicious activity

### 4. **Rate Limiting**
- Limit role change requests per minute
- Prevent abuse of admin endpoints

## Error Handling Strategy

### 1. **Clear Error Messages**
- Specific error messages for different failure types
- Consistent error format across admin endpoints
- Helpful guidance for resolution

### 2. **Proper HTTP Status Codes**
- `400` for client errors (bad input)
- `403` for authorization failures
- `404` for resource not found
- `409` for business rule conflicts

### 3. **Graceful Degradation**
- Handle database connection failures
- Provide meaningful fallback responses
- Log errors for debugging

## Future Enhancements

### 1. **Bulk Operations**
- Update multiple users at once
- Batch role assignments

### 2. **Role History**
- Track role change history
- Provide role change audit trail

### 3. **Advanced Permissions**
- Custom role definitions
- Fine-grained permission management
- Role-based access control (RBAC)

### 4. **User Management Dashboard**
- Web interface for user management
- Visual role management tools

---

This implementation plan provides a secure, well-tested endpoint for user role management that integrates seamlessly with the enhanced Cucumber World class while maintaining proper security boundaries and audit capabilities.