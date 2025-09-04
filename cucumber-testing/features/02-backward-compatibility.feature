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