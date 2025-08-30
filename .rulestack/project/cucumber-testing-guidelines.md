# Cucumber Testing Guidelines

This rule provides guidance for writing maintainable and reliable Cucumber tests for the RFH project.

## Test Reliability Rules

### Never Test Verbose Output
**NEVER** write Cucumber scenarios that test `--verbose` flag output or debug information.

**❌ Bad Example:**
```gherkin
Scenario: Add with verbose output
  When I run "rfh add example-rules@0.1.0 --verbose" in the project directory
  Then I should see "RFH version: 1.0.0"
  And I should see "📦 Adding package: example-rules@0.1.0"
```

**✅ Good Example:**
```gherkin
Scenario: Add existing package successfully
  When I run "rfh add example-rules@0.1.0" in the project directory
  Then I should see "✅ Successfully added example-rules@0.1.0"
  And the command should exit with zero status
```

**Rationale:** Verbose output is for debugging and can change frequently during development. Testing it creates brittle tests that fail when debug output formatting changes.

### Use Enhanced World Setup Functions

**ALWAYS** use the enhanced World setup functions instead of manual configuration when writing step definitions.

**❌ Bad Example:**
```gherkin
Background:
  Given RFH is installed and accessible
  And I have a clean config file
  And I have a registry "test-registry" configured at "http://localhost:8081"
  And "test-registry" is the active registry
```

**✅ Good Example:**
```gherkin
Background:
  Given RFH is installed and accessible
  And the test registry is configured
  And test packages are available
  And I am logged in as root
```

## Available Enhanced World Functions

### Registry and Authentication Setup
- `Given the test registry is configured` - Sets up test registry at localhost:8081 with root authentication
- `Given test packages are available` - Publishes standard test packages (security-rules@1.0.1, example-rules@0.1.0)
- `Given I am logged in as root` - Authenticates as root user for write operations
- `Given I am logged in as a user` - Creates and authenticates as test user for read-only operations
- `Given I am logged in as a user named {string}` - Creates and authenticates as specific named user

### Package Management
- `Given package {string} version {string} is published` - Publishes a specific test package
- `Then package {string} version {string} should exist` - Verifies package exists via API
- `Then I should be authenticated as {string}` - Verifies current authentication state
- `Then the registry should be configured for testing` - Verifies registry setup

## When to Use Manual Setup vs Enhanced World

### Use Enhanced World When:
- Testing package operations (add, publish)
- Testing authentication flows that need existing packages
- Testing features that require a working registry setup
- Writing new step definitions for package-related features

### Use Manual Setup When:
- Testing registry management commands themselves (add, remove, list registries)
- Testing error conditions for missing registries
- Testing authentication failures
- Testing initialization commands that don't need registry setup

## Step Definition Writing Guidelines

### Registry Configuration
```javascript
// ❌ Bad - Manual registry setup in step definitions
Given('I have a registry configured', async function () {
  await this.runCommand('rfh registry add test http://localhost:8081');
  await this.runCommand('rfh registry use test');
});

// ✅ Good - Use enhanced World method
Given('I have a registry configured', async function () {
  await this.ensureRegistrySetup();
});
```

### User Authentication
```javascript
// ❌ Bad - Manual authentication
Given('I am authenticated as root', async function () {
  await this.runCommand('rfh auth login --username root --password root1234');
});

// ✅ Good - Use enhanced World method
Given('I am authenticated as root', async function () {
  await this.loginAsRoot();
});
```

### Package Setup
```javascript
// ❌ Bad - Manual package publishing
Given('packages exist for testing', async function () {
  // Complex manual package creation and publishing logic
});

// ✅ Good - Use enhanced World method
Given('packages exist for testing', async function () {
  await this.setupTestData();
});
```

## Test Structure Guidelines

### Feature Background
Keep backgrounds simple and focused:
```gherkin
Background:
  Given RFH is installed and accessible
  And the test registry is configured
  And test packages are available
  And I am logged in as root
```

### Scenario Independence
Each scenario should be independent and not rely on side effects from other scenarios. The enhanced World provides fresh state for each scenario.

### Error Testing
When testing error conditions, be specific about what you're testing:
```gherkin
# ✅ Good - Tests specific error condition
Scenario: Add package with no registry configured
  Given I have a clean config file with no registries
  When I run "rfh add some-package@1.0.0"
  Then I should see "no registry configured"
```

## Help Text Testing

It's acceptable to verify that help text mentions verbose flags, as this tests the CLI interface:
```gherkin
# ✅ OK - Testing help text content
Then I should see "-v, --verbose"
```

But never test the actual verbose output behavior.

---

Following these guidelines will result in more maintainable, reliable cucumber tests that focus on testing actual functionality rather than debug output formatting.