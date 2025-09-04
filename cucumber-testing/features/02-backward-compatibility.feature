Feature: Registry Type Backward Compatibility
  As a developer upgrading RFH
  I want existing registries without type field to continue working
  So that I don't need to reconfigure my existing setup

  Background:
    Given RFH is installed and accessible
    And I have a clean config file

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
    And the output should match the expected legacy format
    
  Scenario: Map-based responses are converted correctly with legacy wrapper
    Given I have a registry "legacy-maps" configured at "http://localhost:8080" with type "remote-http"
    And "legacy-maps" is the active registry
    And the registry returns structured package data
    When I run "rfh search test --format json" (if format option exists)
    Then the JSON output should be valid
    And the JSON should contain package objects with legacy map structure
    And all expected fields should be present (name, version, description, etc.)

  Scenario: Legacy authentication flow works with refactored HTTP client
    Given I have a registry "legacy-auth" configured at "http://localhost:8080"
    And the registry has legacy JWT token configuration
    When authentication operations are performed
    Then they should work with the new HTTPClient implementation
    And maintain full compatibility with existing token handling