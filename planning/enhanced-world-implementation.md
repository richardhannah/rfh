# Enhanced World Implementation Plan for RuleStack Cucumber Testing

## Overview

This plan outlines the implementation of an enhanced Cucumber World class that provides centralized authentication, user role management, and test data setup for RuleStack testing. This approach follows Cucumber best practices and eliminates code duplication across test scenarios.

## Research Summary

### Is This Acceptable Practice? ✅ **YES**

Based on official Cucumber documentation and industry best practices:

1. **Official Cucumber Guidance**: Using the World object for setup, authentication, and state management is explicitly encouraged
2. **Scenario Isolation**: Each scenario gets a fresh World instance, preventing state leakage
3. **Reusable Utilities**: Helper methods in World are a best practice for reducing code duplication
4. **Clean Setup/Teardown**: Using hooks and World methods for database setup/cleanup is standard

### Key Principles from Research

- "Scenarios must be independent of one another"
- Use dependency injection approaches for state management
- World object provides isolated context for each scenario
- Helper methods in World reduce redundancy and improve maintainability

## Implementation Plan

### Phase 1: Core Authentication & Registry Setup

**Objective**: Enhance the World class with automatic registry configuration and root user authentication.

```javascript
class CustomWorld extends World {
  constructor(options) {
    super(options);
    // Existing properties...
    this.rootJwtToken = null;
    this.currentUser = null;
    this.registryConfigured = false;
    this.testUsers = new Map(); // Track created test users
  }

  // Auto-setup registry and root authentication
  async ensureRegistrySetup() {
    if (!this.registryConfigured) {
      await this.runCommand('rfh registry add test-registry http://localhost:8081');
      await this.runCommand('rfh registry use test-registry');
      await this.loginAsRoot();
      this.registryConfigured = true;
    }
  }

  async loginAsRoot() {
    await this.runCommand('rfh auth login --username root --password root1234');
    this.rootJwtToken = this.extractJwtTokenFromConfig();
    this.currentUser = 'root';
  }

  extractJwtTokenFromConfig() {
    // Read JWT token from ~/.rfh/config.toml
    const configPath = path.join(os.homedir(), '.rfh', 'config.toml');
    if (!fs.existsSync(configPath)) {
      return null;
    }
    
    try {
      const config = fs.readFileSync(configPath, 'utf8');
      const parsed = toml.parse(config);
      
      // Try per-registry JWT token first
      if (parsed.current && parsed.registries && parsed.registries[parsed.current]) {
        const registry = parsed.registries[parsed.current];
        if (registry.jwt_token) {
          return registry.jwt_token;
        }
      }
      
      // Fallback to global user JWT token
      if (parsed.user && parsed.user.token) {
        return parsed.user.token;
      }
      
      return null;
    } catch (error) {
      console.error('Failed to parse config.toml for JWT token:', error);
      return null;
    }
  }
}
```

### Phase 2: User Role Management

**Objective**: Implement methods to create and authenticate as different user roles.

```javascript
// User creation methods using rfh auth register
async createTestUser(username, role = 'user') {
  await this.ensureRegistrySetup();
  
  // Check if user already exists in this scenario
  if (this.testUsers.has(username)) {
    return this.testUsers.get(username);
  }
  
  // For this initial implementation, we only support 'user' and 'root' roles
  // All created test users will have 'user' role - use root for write access
  if (role !== 'user' && role !== 'root') {
    throw new Error(`Unsupported role: ${role}. Use 'user' for read-only or 'root' for write access.`);
  }
  
  // Root user already exists, no need to create
  if (role === 'root' && username === 'root') {
    const userData = { username: 'root', role: 'root', email: 'root@rulestack.init' };
    this.testUsers.set(username, userData);
    return userData;
  }
  
  // Use rfh auth register to create test user (always 'user' role)
  try {
    await this.runCommand(
      `rfh auth register --username ${username} --email ${username}@test.com --password testpass123`
    );
    
    const userData = { username, role: 'user', email: `${username}@test.com` };
    this.testUsers.set(username, userData);
    return userData;
    
  } catch (error) {
    // If registration fails (user might already exist), that's usually okay
    // We'll assume the user exists and proceed
    const userData = { username, role: 'user', email: `${username}@test.com` };
    this.testUsers.set(username, userData);
    return userData;
  }
}

// Helper method to ensure we have root authentication for role changes
async ensureRootAuthentication() {
  if (this.currentUser !== 'root') {
    await this.loginAsRoot();
  }
}

// Role-specific login methods (simplified to user/root only)
async loginAsUser(username = 'test-user') {
  await this.createTestUser(username, 'user');
  await this.runCommand(`rfh auth login --username ${username} --password testpass123`);
  this.currentUser = username;
}

// Helper method for API calls using JWT tokens
async apiCall(method, endpoint, data = null) {
  const baseUrl = 'http://localhost:8081';
  const options = {
    method,
    headers: {
      'Content-Type': 'application/json',
    }
  };
  
  // Extract current JWT token from config
  const jwtToken = this.extractJwtTokenFromConfig();
  if (jwtToken) {
    options.headers['Authorization'] = `Bearer ${jwtToken}`;
  }
  
  if (data) {
    options.body = JSON.stringify(data);
  }
  
  const response = await fetch(`${baseUrl}${endpoint}`, options);
  return await response.json();
}
```

### Phase 3: Test Data Management

**Objective**: Provide methods to create and manage test packages with specific versions.

```javascript
// Package management methods
async publishPackage(name, version, content) {
  // Ensure we have root privileges for publishing
  if (this.currentUser !== 'root') {
    await this.loginAsRoot();
  }
  
  const tempDir = await this.createPackageDir(name, version, content);
  
  try {
    await this.runCommand(`rfh pack rules.mdc --package ${name}`, { cwd: tempDir });
    await this.runCommand('rfh publish', { cwd: tempDir });
  } finally {
    // Clean up temp package directory
    await fs.remove(tempDir);
  }
}

async createPackageDir(name, version, content) {
  const tempDir = await fs.mkdtemp(path.join(os.tmpdir(), `rfh-package-${name}-`));
  
  // Create package structure
  await this.runCommand('rfh init', { cwd: tempDir });
  
  // Write rule content
  await fs.writeFile(path.join(tempDir, 'rules.mdc'), content);
  
  // Configure registry in temp directory
  await this.runCommand('rfh registry add test-registry http://localhost:8081', { cwd: tempDir });
  await this.runCommand('rfh registry use test-registry', { cwd: tempDir });
  
  // Login as root user in temp directory (needed for publishing)
  await this.runCommand('rfh auth login --username root --password root1234', { cwd: tempDir });
  
  return tempDir;
}

// Standard test data setup
async setupTestData() {
  await this.ensureRegistrySetup();
  await this.publishPackage('security-rules', '1.0.1', '# Security Rules\n\nTest security rules package');
  await this.publishPackage('example-rules', '0.1.0', '# Example Rules\n\nTest example rules package');
}

// Verify package exists
async verifyPackageExists(name, version) {
  const response = await this.apiCall('GET', `/v1/packages/${name}/versions/${version}`);
  return response && !response.error;
}
```

### Phase 4: Integration with Existing Tests

**Objective**: Update existing step definitions to use the enhanced World methods.

#### New Step Definitions

```javascript
// Authentication steps (simplified to user/root only)
Given('I am logged in as a user', async function () {
  await this.loginAsUser();
});

Given('I am logged in as root', async function () {
  await this.loginAsRoot();
});

// Test data steps
Given('test packages are available', async function () {
  await this.setupTestData();
});

Given('package {string} version {string} is published', async function (name, version) {
  await this.publishPackage(name, version, `# ${name}\n\nTest package for ${name} v${version}`);
});

// Registry setup (automatic in most cases)
Given('the test registry is configured', async function () {
  await this.ensureRegistrySetup();
});
```

#### Updated Existing Step Definitions

```javascript
// Update registry_steps.js
Given('I have a registry {string} configured at {string}', async function (name, url) {
  // Remove manual setup - now handled by ensureRegistrySetup
  if (url === 'http://localhost:8081') {
    await this.ensureRegistrySetup();
  } else {
    // Keep existing logic for non-test registries
    await this.runCommand(`rfh registry add ${name} ${url}`);
  }
});
```

### Phase 5: Enhanced Cleanup & Hooks

**Objective**: Ensure proper cleanup and setup for each scenario.

```javascript
// Update hooks.js
const { Before, After } = require('@cucumber/cucumber');

Before(async function () {
  // Registry setup is now on-demand, no need for automatic setup
  // Keep existing temp directory setup
});

After(async function () {
  // Enhanced cleanup
  await this.cleanup();
  
  // Clear test users map
  if (this.testUsers) {
    this.testUsers.clear();
  }
  
  // Reset authentication state
  this.currentUser = null;
  this.rootJwtToken = null;
  this.registryConfigured = false;
});

// Optional: BeforeAll hook for one-time setup
BeforeAll(async function () {
  // Could be used for global test environment verification
  // e.g., ensure test API is running
});
```

### Phase 6: Error Handling & Logging

**Objective**: Implement robust error handling and debugging support.

```javascript
// Enhanced error handling
async safeApiCall(method, endpoint, data = null) {
  try {
    return await this.apiCall(method, endpoint, data);
  } catch (error) {
    console.error(`API call failed: ${method} ${endpoint}`, error.message);
    throw new Error(`API call failed: ${error.message}`);
  }
}

async safeRunCommand(command, options = {}) {
  try {
    return await this.runCommand(command, options);
  } catch (error) {
    console.error(`Command failed: ${command}`, error.message);
    throw error;
  }
}

// Debug logging
log(message, level = 'info') {
  if (process.env.CUCUMBER_DEBUG) {
    console.log(`[${level.toUpperCase()}] ${message}`);
  }
}
```

## Benefits of This Implementation

### 1. **DRY Principle**
- Eliminates duplicate registry/auth setup code across step definitions
- Centralizes user creation and authentication logic
- Reduces test maintenance overhead

### 2. **Reliable State Management**
- Fresh database state for each scenario (via test infrastructure)
- Simple two-role system (user/root) is predictable and easy to test
- Consistent package availability for tests

### 3. **Flexible Testing**
- Easy to test read-only (user) vs write (root) scenarios
- Simple package creation for specific test scenarios
- Reusable authentication patterns

### 4. **BDD Compliant**
- Follows official Cucumber best practices
- Maintains scenario independence
- Provides clear, readable step definitions

### 5. **Maintainable**
- Centralized logic is easier to update
- Simplified role model reduces complexity
- Better error handling and debugging

## Considerations & Risks

### 1. **Performance Impact**
- **Risk**: Setup might add time to each scenario
- **Mitigation**: Use lazy loading (on-demand setup) and caching
- **Monitoring**: Track test execution times

### 2. **Role Simplicity**
- **Risk**: Two-role system might be too simple for some test scenarios
- **Mitigation**: Focus on core functionality first, extend roles later if needed
- **Future**: Can add publisher/admin roles in subsequent iterations

### 3. **Error Handling Complexity**
- **Risk**: Setup failures could cause cryptic test failures
- **Mitigation**: Comprehensive error handling and logging
- **Debug Support**: Environment variable for verbose logging

### 4. **Test Data Consistency**
- **Risk**: Package versions might not match test expectations
- **Mitigation**: Explicit version control in package creation
- **Validation**: Verify packages exist before running tests

## Implementation Timeline

### Week 1: Foundation
- [ ] Implement enhanced World class with basic authentication
- [ ] Add registry setup automation
- [ ] Create root user login functionality

### Week 2: User Management
- [ ] Add simplified role-based login methods (user/root only)
- [ ] Create user creation functionality using rfh auth register
- [ ] Test authentication flow with both roles

### Week 3: Package Management
- [ ] Implement package creation and publishing methods
- [ ] Add test data setup functionality
- [ ] Create package verification methods

### Week 4: Integration & Testing
- [ ] Update existing step definitions
- [ ] Add new BDD step definitions
- [ ] Implement error handling and logging
- [ ] Test and debug the complete system

### Week 5: Documentation & Refinement
- [ ] Document new step definitions
- [ ] Create usage examples
- [ ] Performance optimization
- [ ] Final testing and validation

## Success Criteria

1. **Functionality**
   - [ ] All existing tests pass with new World implementation
   - [ ] New role-based testing capabilities work correctly
   - [ ] Package creation and publishing is reliable

2. **Maintainability**
   - [ ] Reduced code duplication in step definitions
   - [ ] Clear separation of test setup concerns
   - [ ] Simple role model is easy to understand and maintain

3. **Performance**
   - [ ] Test execution time doesn't increase significantly
   - [ ] Setup failures are rare and well-handled
   - [ ] Debug information is helpful when issues occur

4. **Reliability**
   - [ ] Tests are more consistent and predictable
   - [ ] Fresh state for each scenario
   - [ ] Proper cleanup prevents state leakage

## Future Enhancements

1. **Configuration Management**
   - Environment-specific test data
   - Configurable user credentials
   - Dynamic API endpoint configuration

2. **Advanced Package Management**
   - Package dependency testing
   - Version conflict scenarios
   - Package update/downgrade testing

3. **Performance Optimization**
   - Parallel user creation
   - Package caching strategies
   - Selective cleanup (only what changed)

4. **Extended Role Testing**
   - Add publisher/admin roles when needed
   - Role transition testing for future role expansions
   - Permission boundary testing between user/root access

---

This implementation plan provides a comprehensive approach to enhancing the Cucumber World class for RuleStack testing, following established best practices while addressing the specific needs of the application's authentication and package management features.