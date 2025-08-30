# Root User Implementation Plan

## Overview
Implement a root user system that provides superuser access to the RuleStack API, with automatic creation on application startup if no root user exists.

## Requirements

### 1. Root Role Definition
- Add a new `root` role to the existing user role hierarchy
- Root role should have unrestricted access to all API endpoints
- Only ONE root user can exist in the system at any time

### 2. Root User Creation
- **NOT created via database migration** (Flyway)
- Created programmatically when the API server starts
- Only created if no root user exists in the database
- Default credentials (hardcoded for now, configurable in future):
  - Username: `root`
  - Email: `root@rulestack.init`
  - Password: `root1234`

### 3. Permission Hierarchy
Current roles and their permissions:
- `user`: Read-only access (can download packages)
- `publisher`: Can publish packages + user permissions
- `admin`: Can manage users + publisher permissions  
- `root`: Full unrestricted access to everything

## Implementation Steps

### Phase 1: Database Schema Updates

#### 1.1 Add Root Role to Enum
- File: `migrations/rulestack/migrations/V4__add_root_role.sql`
- Add `root` value to the `user_role` enum type
- Add unique constraint to ensure only one root user exists
- Document the role hierarchy in database comments

```sql
-- Add root to user_role enum
ALTER TYPE rulestack.user_role ADD VALUE 'root';

-- Ensure only one root user can exist
CREATE UNIQUE INDEX only_one_root_user 
ON rulestack.users (role) 
WHERE role = 'root';
```

### Phase 2: Update Go Code

#### 2.1 Update User Role Constants
- File: `internal/db/users.go`
- Add `RoleRoot UserRole = "root"` constant
- Ensure proper serialization/deserialization

#### 2.2 Create Root User on Startup
- File: `cmd/api/main.go`
- Add `ensureRootUser()` function
- Call after database connection is established
- Before API routes are registered

```go
func ensureRootUser(database *db.Database) error {
    // 1. Check if root user exists
    // 2. If not, create with default credentials
    // 3. Log the action
    // 4. Handle race conditions gracefully
}
```

#### 2.3 Update Authentication Middleware
- File: `internal/api/middleware/auth.go`
- Update `HasRole()` function to give root users access to everything
- Update `RequireRole()` to always pass for root users

```go
func HasRole(userRole UserRole, requiredRole UserRole) bool {
    // Root has access to everything
    if userRole == db.RoleRoot {
        return true
    }
    // Existing role hierarchy logic
    // ...
}
```

### Phase 3: Admin Endpoints

#### 3.1 User Management Endpoint
- File: `internal/api/admin.go` (new file)
- Endpoint: `PUT /v1/admin/users/{id}/role`
- Only accessible to `admin` and `root` roles
- Allows changing user roles
- Prevents changing/creating another root user
- Prevents demoting the root user

```go
type UpdateUserRoleRequest struct {
    Role UserRole `json:"role"`
}

func updateUserRoleHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Validate caller has admin or root role
    // 2. Parse target user ID and new role
    // 3. Prevent creating/changing root users
    // 4. Update user role in database
    // 5. Return success/error response
}
```

#### 3.2 List Users Endpoint
- Endpoint: `GET /v1/admin/users`
- Only accessible to `admin` and `root` roles
- Returns paginated list of users
- Includes role information

### Phase 4: Security Considerations

#### 4.1 Root User Protection
- Cannot be deleted
- Cannot be deactivated  
- Cannot have role changed (always root)
- Password can only be changed by the root user itself

#### 4.2 Audit Logging
- Log all root user actions
- Log failed root login attempts
- Consider rate limiting for root login attempts

### Phase 5: Testing

#### 5.1 Unit Tests
- Test root user creation on startup
- Test permission checks for root role
- Test admin endpoints

#### 5.2 Integration Tests
- Test root user can access all endpoints
- Test admin users can manage other users (except root)
- Test regular users cannot access admin endpoints

#### 5.3 Cucumber Tests
- Update test setup to use root user for creating test data
- Test package publishing with proper permissions

## Migration Strategy

1. Deploy new code with root user creation
2. On first startup, root user will be created automatically
3. Admin can then use root credentials to set up other users with appropriate roles
4. For testing, use root user to create test publishers

## Future Enhancements

1. **Configurable Root User**
   - Environment variables for initial root credentials
   - Require password change on first login

2. **Root User Recovery**
   - Mechanism to reset root password via environment variable
   - Emergency access procedures

3. **Multiple Admin Users**
   - Allow multiple admin users
   - Separate admin role from root role more clearly

4. **Role-Based Access Control (RBAC)**
   - More granular permissions
   - Custom roles

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Root password exposed in code | Move to environment variables in production |
| Root account compromised | Implement audit logging and monitoring |
| Race condition during creation | Use database constraints and handle conflicts |
| Root user locked out | Implement recovery mechanism |

## Success Criteria

1. Root user is automatically created on API startup
2. Root user has access to all API endpoints
3. Admin endpoints work correctly for user management
4. Cucumber tests pass using root user for test setup
5. Only one root user can exist at any time
6. System remains secure with proper access controls

## Timeline

- Phase 1: 30 minutes (Database schema)
- Phase 2: 1 hour (Core implementation)
- Phase 3: 1 hour (Admin endpoints)
- Phase 4: 30 minutes (Security hardening)
- Phase 5: 1 hour (Testing)

Total: ~4 hours

## Notes

- The root user is a privileged account and should be used sparingly
- In production, the default password MUST be changed immediately
- Consider implementing two-factor authentication for root user in future
- All root user actions should be audited for security compliance