#!/bin/bash

# Setup script for Gitea test repositories
# Creates test repositories with proper RuleStack registry structure

set -e

GITEA_URL="http://localhost:3000"
ADMIN_USER="rfh-admin"
ADMIN_PASSWORD="admin123456"
ADMIN_EMAIL="admin@localhost"
ADMIN_TOKEN=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Wait for Gitea to be ready
wait_for_gitea() {
    log_info "Waiting for Gitea to be ready..."
    for i in {1..30}; do
        if curl -s "${GITEA_URL}/api/v1/version" > /dev/null 2>&1; then
            log_info "Gitea is ready!"
            return 0
        fi
        echo -n "."
        sleep 2
    done
    log_error "Gitea failed to start within 60 seconds"
    exit 1
}

# Create admin user and get access token
setup_admin_user() {
    log_info "Setting up admin user..."
    
    # Create admin user via Gitea CLI (inside container)
    docker exec rulestack-gitea-test gitea admin user create \
        --username "${ADMIN_USER}" \
        --password "${ADMIN_PASSWORD}" \
        --email "${ADMIN_EMAIL}" \
        --admin \
        --must-change-password=false || log_warn "Admin user may already exist"
    
    # Get access token
    ADMIN_TOKEN=$(docker exec rulestack-gitea-test gitea admin user generate-access-token \
        --username "${ADMIN_USER}" \
        --token-name "test-token" \
        --scopes "all" | grep -o 'Access token: [a-z0-9]*' | cut -d' ' -f3)
    
    if [[ -z "$ADMIN_TOKEN" ]]; then
        log_error "Failed to generate admin token"
        exit 1
    fi
    
    log_info "Admin user created with token: ${ADMIN_TOKEN}"
}

# Create a repository via API
create_repository() {
    local repo_name="$1"
    local description="$2"
    local private="$3"
    
    log_info "Creating repository: $repo_name"
    
    curl -X POST "${GITEA_URL}/api/v1/user/repos" \
        -H "Authorization: token ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$repo_name\",
            \"description\": \"$description\",
            \"private\": $private,
            \"auto_init\": true,
            \"default_branch\": \"main\"
        }" > /dev/null
    
    if [[ $? -eq 0 ]]; then
        log_info "Repository $repo_name created successfully"
    else
        log_warn "Repository $repo_name may already exist"
    fi
}

# Clone and populate repository with test data
populate_repository() {
    local repo_name="$1"
    local repo_type="$2"
    
    log_info "Populating repository: $repo_name"
    
    local temp_dir="/tmp/gitea-setup-$$"
    mkdir -p "$temp_dir"
    cd "$temp_dir"
    
    # Clone repository
    git clone "${GITEA_URL}/${ADMIN_USER}/${repo_name}.git"
    cd "$repo_name"
    
    # Configure git for commits
    git config user.name "RFH Test Setup"
    git config user.email "test@localhost"
    
    if [[ "$repo_type" == "valid" ]]; then
        create_valid_registry_structure
    elif [[ "$repo_type" == "invalid" ]]; then
        create_invalid_registry_structure
    fi
    
    # Commit and push
    git add .
    git commit -m "Add test registry structure and data"
    git push origin main
    
    # Cleanup
    cd /
    rm -rf "$temp_dir"
}

# Create valid registry structure with test packages
create_valid_registry_structure() {
    log_info "Creating valid registry structure..."
    
    # Create packages directory structure
    mkdir -p packages/security-rules/versions/{1.0.0,1.1.0}
    mkdir -p packages/example-rules/versions/1.0.0
    
    # Security rules v1.0.0
    cat > packages/security-rules/versions/1.0.0/manifest.json << 'EOF'
{
  "name": "security-rules",
  "version": "1.0.0",
  "description": "Security rules for RuleStack testing",
  "author": "RFH Test Team",
  "license": "MIT",
  "keywords": ["security", "rules", "testing"],
  "created": "2025-01-01T00:00:00Z"
}
EOF

    cat > packages/security-rules/versions/1.0.0/security-rules.mdc << 'EOF'
# Security Rules v1.0.0

## Authentication Rules
- Always validate user input
- Use secure authentication mechanisms
- Never store passwords in plaintext

## Authorization Rules  
- Implement principle of least privilege
- Validate permissions on every request
- Use role-based access control

## Data Protection Rules
- Encrypt sensitive data at rest and in transit
- Sanitize all user inputs
- Use prepared statements for database queries
EOF

    # Security rules v1.1.0
    cat > packages/security-rules/versions/1.1.0/manifest.json << 'EOF'
{
  "name": "security-rules",
  "version": "1.1.0",
  "description": "Enhanced security rules for RuleStack testing",
  "author": "RFH Test Team",
  "license": "MIT",
  "keywords": ["security", "rules", "testing", "enhanced"],
  "created": "2025-01-15T00:00:00Z"
}
EOF

    cat > packages/security-rules/versions/1.1.0/security-rules.mdc << 'EOF'
# Security Rules v1.1.0

## Authentication Rules
- Always validate user input
- Use secure authentication mechanisms (MFA preferred)
- Never store passwords in plaintext
- Implement session timeout mechanisms

## Authorization Rules  
- Implement principle of least privilege
- Validate permissions on every request
- Use role-based access control
- Log all authorization decisions

## Data Protection Rules
- Encrypt sensitive data at rest and in transit
- Sanitize all user inputs
- Use prepared statements for database queries
- Implement data retention policies
EOF

    # Example rules v1.0.0
    cat > packages/example-rules/versions/1.0.0/manifest.json << 'EOF'
{
  "name": "example-rules",
  "version": "1.0.0",
  "description": "Example rules for RuleStack testing and demos",
  "author": "RFH Test Team",
  "license": "MIT",
  "keywords": ["example", "demo", "testing"],
  "created": "2025-01-01T00:00:00Z"
}
EOF

    cat > packages/example-rules/versions/1.0.0/example-rules.mdc << 'EOF'
# Example Rules v1.0.0

## Code Style Rules
- Use consistent indentation (2 or 4 spaces)
- Follow naming conventions for your language
- Write meaningful variable and function names
- Keep functions focused and small

## Documentation Rules
- Document all public APIs
- Include usage examples in comments
- Keep README files up to date
- Use clear and concise language

## Testing Rules
- Write tests for all critical functionality
- Use descriptive test names
- Keep tests independent and isolated
- Maintain good test coverage
EOF

    # Registry index file
    cat > index.json << 'EOF'
{
  "name": "RFH Test Registry",
  "description": "Test registry for RuleStack Git client development",
  "version": "1.0.0",
  "packages": {
    "security-rules": {
      "description": "Security rules for RuleStack testing",
      "versions": ["1.0.0", "1.1.0"],
      "latest": "1.1.0"
    },
    "example-rules": {
      "description": "Example rules for RuleStack testing and demos",
      "versions": ["1.0.0"],
      "latest": "1.0.0"
    }
  },
  "updated": "2025-01-15T00:00:00Z"
}
EOF

    # README
    cat > README.md << 'EOF'
# RFH Test Registry

This is a test registry for RuleStack Git client development and testing.

## Structure

- `packages/` - Contains all rule packages organized by name and version
- `index.json` - Registry metadata and package index
- Each package version has a `manifest.json` with metadata and rule files

## Test Packages

- **security-rules**: Security rules for applications (v1.0.0, v1.1.0)
- **example-rules**: Example rules for demonstrations (v1.0.0)

This registry is automatically populated by the RuleStack test setup scripts.
EOF
}

# Create invalid registry structure (missing required directories/files)
create_invalid_registry_structure() {
    log_info "Creating invalid registry structure..."
    
    # Create some files but not the expected registry structure
    cat > README.md << 'EOF'
# Invalid Registry

This repository does not contain the proper RuleStack registry structure.
It is used for testing error handling in the Git client.
EOF

    cat > some-other-file.txt << 'EOF'
This is not a valid registry structure.
No packages/ directory or index.json file.
EOF
}

# Main execution
main() {
    log_info "Starting Gitea test repository setup..."
    
    # Check if Gitea container is running
    if ! docker ps | grep -q rulestack-gitea-test; then
        log_error "Gitea container is not running. Please start with: docker-compose -f docker-compose.test.yml up -d gitea-test"
        exit 1
    fi
    
    wait_for_gitea
    setup_admin_user
    
    # Create test repositories
    create_repository "rfh-test-registry-public" "Public test registry with valid structure" false
    create_repository "rfh-test-registry-private" "Private test registry with valid structure" true  
    create_repository "rfh-test-invalid-registry" "Invalid registry structure for error testing" false
    
    # Populate repositories with test data
    populate_repository "rfh-test-registry-public" "valid"
    populate_repository "rfh-test-registry-private" "valid"
    populate_repository "rfh-test-invalid-registry" "invalid"
    
    log_info "Gitea test setup complete!"
    log_info "Access Gitea at: ${GITEA_URL}"
    log_info "Admin user: ${ADMIN_USER} / ${ADMIN_PASSWORD}"
    log_info "Admin token: ${ADMIN_TOKEN}"
    log_info ""
    log_info "Test repositories created:"
    log_info "- ${GITEA_URL}/${ADMIN_USER}/rfh-test-registry-public (public)"
    log_info "- ${GITEA_URL}/${ADMIN_USER}/rfh-test-registry-private (private)"
    log_info "- ${GITEA_URL}/${ADMIN_USER}/rfh-test-invalid-registry (public, invalid structure)"
}

# Run main function
main "$@"