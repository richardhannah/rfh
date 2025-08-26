Feature: RFH Publish Command
  As a developer
  I want to publish my rule packages to a registry
  So that others can discover and use my rulesets

  Background:
    Given RFH is installed and accessible

  Scenario: Publish command help text
    When I run "rfh publish --help"
    Then I should see "Publish a ruleset package to the configured registry"
    And I should see "--archive string   path to archive file"
    And I should see "Requires authentication token to be configured in the registry"

  Scenario: Publish with no active registry
    Given I have a clean config file with no registries
    And I have a temporary project directory
    And I have a rulestack.json manifest with name "no-registry" and version "1.0.0"
    And I have a rule file "test.md" with content "# Test Rules"
    When I run "rfh pack" in the project directory
    And I run "rfh publish" in the project directory
    Then I should see "no active registry configured"
    And I should see "Use 'rfh registry add' to add one"
    And the command should exit with non-zero status

  Scenario: Publish without authentication token
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And I have a rulestack.json manifest with name "unauth-test" and version "1.0.0"
    And I have a rule file "rules.md" with content "# Unauth Test Rules"
    When I run "rfh pack" in the project directory
    And I run "rfh publish" in the project directory
    Then I should see "no authentication token available"
    And I should see "Use 'rfh auth login' to authenticate"
    And the command should exit with non-zero status

  Scenario: Publish with missing manifest
    Given I have a temporary project directory
    When I run "rfh publish" in the project directory
    Then I should see "failed to load manifest"
    And I should see "The system cannot find the file specified"
    And the command should exit with non-zero status

  Scenario: Publish with missing archive file
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And I have a rulestack.json manifest with name "missing-archive" and version "1.0.0"
    When I run "rfh publish" in the project directory
    Then I should see "archive not found: missing-archive-1.0.0.tgz"
    And I should see "Run 'rfh pack' first or specify --archive"
    And the command should exit with non-zero status

  Scenario: Publish with custom missing archive file
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And I have a rulestack.json manifest with name "custom-archive" and version "1.0.0"
    When I run "rfh publish --archive nonexistent.tgz" in the project directory
    Then I should see "archive not found: nonexistent.tgz"
    And I should see "Run 'rfh pack' first or specify --archive"
    And the command should exit with non-zero status

  Scenario: Publish verbose output shows configuration
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And I have a rulestack.json manifest with name "verbose-test" and version "1.0.0"
    When I run "rfh publish --verbose" in the project directory
    Then I should see "RFH version: 1.0.0"
    And I should see "Config file:"
    And I should see "archive not found: verbose-test-1.0.0.tgz"
    And the command should exit with non-zero status

  Scenario: Publish with token override flag
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And I have a rulestack.json manifest with name "token-test" and version "1.0.0"
    And I have a rule file "rules.md" with content "# Token Test Rules"
    When I run "rfh pack" in the project directory
    And I run "rfh publish --token dummy-token" in the project directory
    Then I should see either authentication error or connection error
    And the command should exit with non-zero status

  Scenario: Publish with registry URL override
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And I have a rulestack.json manifest with name "override-test" and version "1.0.0"
    And I have a rule file "rules.md" with content "# Override Test Rules"
    When I run "rfh pack" in the project directory
    And I run "rfh publish --registry http://localhost:9999 --token dummy-token" in the project directory
    Then I should see either authentication error or connection error
    And the command should exit with non-zero status

  Scenario: Publish shows proper workflow guidance
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And I have a rulestack.json manifest with name "workflow-test" and version "1.0.0"
    When I run "rfh publish" in the project directory
    Then I should see "archive not found"
    And I should see "Run 'rfh pack' first"
    And the command should exit with non-zero status