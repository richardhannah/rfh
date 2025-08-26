# RFH Usage Scenarios and Expected Behavior

This document defines the expected behavior for each RFH command using Cucumber/Gherkin format to ensure consistent functionality and user experience.

## ✅ rfh init

**Feature**: Initialize RuleStack Project

```gherkin
Feature: Initialize RuleStack Project
  As a developer
  I want to initialize a new RuleStack project
  So that I can create and manage rule packages

  Background:
    Given I am in an empty directory

  Scenario: Initialize new project successfully
    When I run "rfh init"
    Then I should see "Initialized RuleStack project"
    And a file "rulestack.json" should be created
    And the manifest should have name "example-rules"
    And the manifest should not contain scope characters "@" or "/"
    And a directory ".rulestack" should be created
    And a file "CLAUDE.md" should be created
    And core rules should be downloaded to ".rulestack/core.v1.0.0/"

  Scenario: Initialize in directory with existing project
    Given a RuleStack project already exists
    When I run "rfh init"
    Then I should see a warning about existing project
    And existing files should not be overwritten without confirmation

  Scenario: Initialize with custom project name
    When I run "rfh init --name my-custom-rules"
    Then the manifest should have name "my-custom-rules"
    And the name should not contain scope characters

  Scenario: Generate simple package names without scopes
    When I run "rfh init"
    Then the default package name should be "example-rules"
    And the package name should not be "@acme/example-rules"
    And no scope characters should appear in the manifest
```

**Status**: ✅ All scenarios passing

---

## ✅ rfh registry

**Feature**: Manage Registry Configurations

```gherkin
Feature: Manage Registry Configurations
  As a developer
  I want to manage multiple registry configurations
  So that I can work with different package repositories

  Background:
    Given I have RFH installed
    And my config file is empty or doesn't exist

  Scenario: Add first registry
    When I run "rfh registry add production https://registry.example.com"
    Then I should see "Added registry 'production'"
    And I should see "Set as active registry"
    And the config should contain registry "production" with URL "https://registry.example.com"
    And "production" should be set as the current active registry

  Scenario: Add additional registry
    Given I have a registry "production" configured
    When I run "rfh registry add staging https://staging.example.com"
    Then I should see "Added registry 'staging'"
    And the config should contain both registries
    And "production" should remain the active registry

  Scenario: Switch active registry
    Given I have registries "production" and "staging" configured
    And "production" is the active registry
    When I run "rfh registry use staging"
    Then I should see "Set 'staging' as active registry"
    And "staging" should be the current active registry

  Scenario: List registries
    Given I have registries "production" and "staging" configured
    And "staging" is the active registry
    When I run "rfh registry list"
    Then I should see both registries listed
    And "staging" should be marked as active

  Scenario: Remove non-active registry
    Given I have registries "production" and "staging" configured
    And "staging" is the active registry
    When I run "rfh registry remove production"
    Then I should see "Removed registry 'production'"
    And "staging" should remain the active registry

  Scenario: Remove active registry
    Given I have registries "production" and "staging" configured
    And "staging" is the active registry
    When I run "rfh registry remove staging"
    Then I should see "Removed active registry"
    And I should see warning about setting a new active registry
    And no registry should be active

  Scenario: Handle invalid registry operations
    Given I have no registries configured
    When I run "rfh registry use nonexistent"
    Then I should see an error "Registry 'nonexistent' not found"

  Scenario: Validate URL format
    When I run "rfh registry add test invalid-url"
    Then I should see an error about invalid URL format
```

**Status**: ✅ All scenarios passing

---

## ✅ rfh auth

**Feature**: Handle User Authentication

```gherkin
Feature: Handle User Authentication
  As a developer
  I want to authenticate with registries
  So that I can publish packages and access private repositories

  Background:
    Given I have a registry "test" configured at "http://localhost:8080"
    And "test" is the active registry

  Scenario: Login successfully
    When I run "rfh auth login"
    And I enter username "testuser"
    And I enter password "password123"
    Then I should see "Login successful"
    And a JWT token should be stored in the registry configuration
    And the token should be valid

  Scenario: Login with invalid credentials
    When I run "rfh auth login"
    And I enter username "testuser"
    And I enter password "wrongpassword"
    Then I should see "Invalid credentials"
    And no token should be stored

  Scenario: Check authentication status when logged in
    Given I am authenticated as "testuser" with role "publisher"
    When I run "rfh auth status"
    Then I should see "Authenticated as: testuser"
    And I should see "Role: publisher"
    And I should see token expiration information

  Scenario: Check authentication status when not logged in
    Given I am not authenticated
    When I run "rfh auth status"
    Then I should see "Not authenticated"

  Scenario: Logout successfully
    Given I am authenticated as "testuser"
    When I run "rfh auth logout"
    Then I should see "Logged out successfully"
    And the JWT token should be removed from configuration

  Scenario: Logout when not logged in
    Given I am not authenticated
    When I run "rfh auth logout"
    Then I should see "Already logged out"
    And no errors should occur

  Scenario: Login without active registry
    Given I have no active registry configured
    When I run "rfh auth login"
    Then I should see an error "No active registry configured"

  Scenario: Handle token expiration
    Given I have an expired JWT token
    When I run a command requiring authentication
    Then I should see "Token expired, please login again"
    And the expired token should be cleared

  Scenario: Role-based access control
    Given I am authenticated with role "user"
    When I attempt to publish a package
    Then I should see "Insufficient permissions"
    And the operation should fail
```

**Status**: ✅ All scenarios passing

---

## 🚧 rfh pack (Under Review)

**Feature**: Create Distributable Archives

*Scenarios to be defined based on testing results*

---

## 🚧 rfh publish (Under Review)

**Feature**: Upload Packed Archives

*Scenarios to be defined based on testing results*

---

## Testing Guidelines

**Common Expectations Across All Features:**

```gherkin
Feature: Common Command Behaviors
  As a developer using RFH
  I want consistent behavior across all commands
  So that I have a predictable user experience

  Scenario: Commands provide helpful error messages
    When any command encounters an error
    Then I should see a clear, actionable error message
    And the message should include suggested next steps where applicable

  Scenario: Configuration management
    Given any command that uses configuration
    When the command runs
    Then configuration should be stored in "~/.rfh/config.toml"
    And the config file should use consistent formatting

  Scenario: Prerequisite validation
    Given a command requires authentication
    When I run the command without being authenticated
    Then I should see "Authentication required. Run 'rfh auth login' first"

  Scenario: Network error handling
    Given any command that makes network requests
    When the network is unavailable
    Then I should see a clear network error message
    And the command should timeout gracefully within reasonable limits

  Scenario: File operation safety
    Given any command that modifies files
    When the operation could cause data loss
    Then I should be prompted for confirmation
    And the operation should be atomic where possible
```