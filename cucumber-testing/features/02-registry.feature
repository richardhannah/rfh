Feature: Registry Management
  As a developer
  I want to manage registry configurations
  So that I can work with different package repositories

  Background:
    Given RFH is installed and accessible
    And I have a clean config file

  # Basic registry operations
  
  Scenario: Add first registry
    When I run "rfh registry add production https://registry.example.com"
    Then I should see "Added registry 'production'"
    And I should see "Set as active registry"
    And I should see "Use 'rfh auth login' to authenticate"
    And the config should contain registry "production" with URL "https://registry.example.com"
    And "production" should be the current active registry

  Scenario: Add additional registry
    Given I have a registry "production" configured at "https://registry.example.com"
    When I run "rfh registry add staging https://staging.example.com"
    Then I should see "Added registry 'staging'"
    And the config should contain both registries
    And "production" should remain the active registry

  Scenario: List registries when none configured
    When I run "rfh registry list"
    Then I should see "No registries configured"

  Scenario: List single registry
    Given I have a registry "production" configured at "https://registry.example.com"
    When I run "rfh registry list"
    Then I should see "production" in the registry list
    And I should see "https://registry.example.com" in the registry list
    And I should see "production" marked as active

  Scenario: List multiple registries
    Given I have a registry "production" configured at "https://registry.example.com"
    And I have a registry "staging" configured at "https://staging.example.com"
    And "staging" is the active registry
    When I run "rfh registry list"
    Then I should see both registries in the list
    And I should see "staging" marked as active
    And I should not see "production" marked as active

  Scenario: Switch active registry
    Given I have a registry "production" configured at "https://registry.example.com"
    And I have a registry "staging" configured at "https://staging.example.com"
    And "production" is the active registry
    When I run "rfh registry use staging"
    Then I should see "Set 'staging' as active registry"
    And "staging" should be the current active registry

  Scenario: Remove non-active registry
    Given I have a registry "production" configured at "https://registry.example.com"
    And I have a registry "staging" configured at "https://staging.example.com"
    And "staging" is the active registry
    When I run "rfh registry remove production"
    Then I should see "Removed registry 'production'"
    And "staging" should remain the active registry
    And the config should not contain registry "production"

  Scenario: Remove active registry
    Given I have a registry "production" configured at "https://registry.example.com"
    And I have a registry "staging" configured at "https://staging.example.com"
    And "staging" is the active registry
    When I run "rfh registry remove staging"
    Then I should see "Removed active registry"
    And I should see a warning about setting a new active registry
    And no registry should be active

  Scenario: Handle invalid registry operations
    When I run "rfh registry use nonexistent"
    Then I should see an error about registry not found
    And the command should exit with non-zero status

  Scenario: Handle duplicate registry name
    Given I have a registry "production" configured at "https://registry.example.com"
    When I run "rfh registry add production https://different.example.com"
    Then I should see "Added registry 'production'"
    And the config should contain registry "production" with URL "https://different.example.com"

  # Registry types

  Scenario: Add HTTP registry with explicit type
    When I run "rfh registry add http-typed https://registry.example.com --type remote-http"
    Then I should see "Added registry 'http-typed'"
    And I should see "Type: remote-http"
    And I should see "Use 'rfh auth login' to authenticate"
    And the config should contain registry "http-typed" with type "remote-http"
    
  Scenario: Add Git registry
    When I run "rfh registry add git-registry https://github.com/org/registry --type git"
    Then I should see "Added registry 'git-registry'"
    And I should see "Type: git"
    And I should see "Set git_token in config or use GITHUB_TOKEN environment variable"
    And the config should contain registry "git-registry" with type "git"
    
  Scenario: Add registry without type defaults to HTTP
    When I run "rfh registry add default-registry https://registry.example.com"
    Then I should see "Added registry 'default-registry'"
    And I should see "Type: remote-http"
    And the config should contain registry "default-registry" with type "remote-http"
    
  Scenario: Reject invalid registry type
    When I run "rfh registry add invalid https://example.com --type invalid-type"
    Then the command should exit with non-zero status
    And I should see an error containing "unsupported registry type"
    
  Scenario: List shows registry types
    Given I have a registry "typed-http" configured at "https://http.example.com" with type "remote-http"
    And I have a registry "typed-git" configured at "https://github.com/org/repo" with type "git"
    When I run "rfh registry list"
    Then I should see "typed-http (remote-http)" in the registry list
    And I should see "typed-git (git)" in the registry list

  Scenario: Git registry URL validation warning
    When I run "rfh registry add git-invalid https://example.com/not-git --type git"
    Then I should see "Added registry 'git-invalid'"
    And I should see "Warning: Git registry URL may not be valid"
    And I should see "Type: git"

  # Backward compatibility

  Scenario: Load config with registries missing type field
    Given a config file with content:
      """
      current = "legacy"
      
      [registries.legacy]
      url = "https://legacy.example.com"
      jwt_token = "token123"
      """
    When I run "rfh registry list"
    Then I should see "legacy (remote-http)" in the registry list
    And the command should exit with zero status
    
  Scenario: Existing registry operations work without type
    Given a config file with content:
      """
      [registries.old]
      url = "https://old.example.com"
      """
    When I run "rfh registry use old"
    Then I should see "Set 'old' as active registry"
    And "old" should be the current active registry
    And the command should exit with zero status

  Scenario: Mixed old and new registries work together
    Given a config file with content:
      """
      current = "old"
      
      [registries.old]
      url = "https://old.example.com"
      
      [registries.new]
      url = "https://github.com/org/registry"
      type = "git"
      """
    When I run "rfh registry list"
    Then I should see "old (remote-http)" in the registry list
    And I should see "new (git)" in the registry list
    And I should see "old" marked as active

  Scenario: Adding new registry to existing config preserves old ones
    Given a config file with content:
      """
      [registries.legacy]
      url = "https://legacy.example.com"
      jwt_token = "oldtoken"
      """
    When I run "rfh registry add modern https://github.com/org/registry --type git"
    Then I should see "Added registry 'modern'"
    And the config should contain registry "legacy" with type "remote-http"
    And the config should contain registry "modern" with type "git"
    # Ensure JWT token is preserved
    And the config should contain "jwt_token = 'oldtoken'"

  Scenario: Legacy client code continues to work with new HTTP client
    Given I have a registry "legacy-http" configured at "http://localhost:8080"
    # This registry has no type field, testing backward compatibility
    And "legacy-http" is the active registry
    When I run "rfh search package"
    Then the command should succeed
    And I should see search results or "No packages found"