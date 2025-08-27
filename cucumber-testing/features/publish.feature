Feature: RFH Publish Command
  As a developer
  I want to publish my rule packages to a registry
  So that others can discover and use my rulesets

  Background:
    Given RFH is installed and accessible

  # Basic functionality
  
  Scenario: Publish command help text
    When I run "rfh publish --help"
    Then I should see "Publish all staged ruleset packages to the configured registry"
    And I should see "Requires authentication token to be configured in the registry"

  # No staged archives scenarios
  
  Scenario: Publish command with no staged archives
    Given I have a temporary project directory
    And RFH is initialized in the directory  
    When I run "rfh publish" in the project directory
    Then I should see "no staged archives found"
    And I should see "Use 'rfh pack' to create archives first"
    And the command should exit with non-zero status

  Scenario: Publish with missing manifest
    Given I have a temporary project directory
    When I run "rfh publish" in the project directory
    Then I should see "no staged archives found"
    And I should see "Use 'rfh pack' to create archives first"
    And the command should exit with non-zero status

  # Registry configuration scenarios
  
  Scenario: Publish with no active registry
    Given I have a clean config file with no registries
    And I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "test.mdc" with content "# Test Rules"
    When I run "rfh pack test.mdc --package=no-registry" in the project directory
    And I run "rfh publish" in the project directory
    Then I should see "no active registry configured"
    And I should see "Use 'rfh registry add' to add one"
    And the command should exit with non-zero status

  # Authentication scenarios
  
  Scenario: Publish without authentication token
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "rules.mdc" with content "# Unauth Test Rules"
    When I run "rfh pack rules.mdc --package=unauth-test" in the project directory
    And I run "rfh publish" in the project directory
    Then I should see "no authentication token available"
    And I should see "Use 'rfh auth login' to authenticate"
    And the command should exit with non-zero status

  Scenario: Publish with no authentication configured
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "rules.mdc" with content "# Token Test Rules"
    When I run "rfh pack rules.mdc --package=token-test" in the project directory
    And I run "rfh publish" in the project directory
    Then I should see either authentication error or connection error
    And the command should exit with non-zero status

  # Multiple staged archives
  
  Scenario: Publish multiple staged archives
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a rule file "security-rule.mdc" with content "# Security Rule"
    And I run "rfh pack security-rule.mdc --package=security-rules" in the project directory
    And I have a rule file "network-rule.mdc" with content "# Network Rule"  
    And I run "rfh pack network-rule.mdc --package=network-rules" in the project directory
    When I run "rfh publish" in the project directory
    Then I should see "Found 2 staged archive(s) to publish"
    And I should see "- security-rules-1.0.0.tgz"
    And I should see "- network-rules-1.0.0.tgz"
    And I should see either authentication error or connection error

  # Registry override scenarios
  
  Scenario: Publish with invalid registry configuration  
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:9999"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "rules.mdc" with content "# Override Test Rules"
    When I run "rfh pack rules.mdc --package=override-test" in the project directory
    And I run "rfh publish" in the project directory
    Then I should see either authentication error or connection error
    And the command should exit with non-zero status

  # Verbose output
  
  Scenario: Publish verbose output shows configuration
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    When I run "rfh publish --verbose" in the project directory
    Then I should see "RFH version: 1.0.0"
    And I should see "no staged archives found"
    And the command should exit with non-zero status

  # Staging cleanup verification
  
  Scenario: Verify staged archives exist after pack
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "test-rule.mdc" with content "# Test Rule"
    When I run "rfh pack test-rule.mdc --package=test-package" in the project directory
    Then the archive file ".rulestack/staged/test-package-1.0.0.tgz" should exist
    # Note: Actual cleanup after successful publish would require a mock registry