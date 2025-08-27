Feature: RFH Publish Command
  As a developer
  I want to publish my rule packages to a registry
  So that others can discover and use my rulesets

  Background:
    Given RFH is installed and accessible

  Scenario: Publish command help text
    When I run "rfh publish --help"
    Then I should see "Publish all staged ruleset packages to the configured registry"
    And I should see "Requires authentication token to be configured in the registry"

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

  Scenario: Publish with missing manifest
    Given I have a temporary project directory
    When I run "rfh publish" in the project directory
    Then I should see "no staged archives found"
    And I should see "Use 'rfh pack' to create archives first"
    And the command should exit with non-zero status

  Scenario: Publish with missing archive file
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    When I run "rfh publish" in the project directory
    Then I should see "no staged archives found"
    And I should see "Use 'rfh pack' to create archives first"
    And the command should exit with non-zero status

  Scenario: Publish with no staged archives shows guidance
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    When I run "rfh publish" in the project directory
    Then I should see "no staged archives found"
    And I should see "Use 'rfh pack' to create archives first"
    And the command should exit with non-zero status

  Scenario: Publish verbose output shows configuration
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    When I run "rfh publish --verbose" in the project directory
    Then I should see "RFH version: 1.0.0"
    And I should see "Config file:"
    And I should see "no staged archives found"
    And the command should exit with non-zero status

  Scenario: Publish with token override flag
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "rules.mdc" with content "# Token Test Rules"
    When I run "rfh pack rules.mdc --package=token-test" in the project directory
    And I run "rfh publish --token dummy-token" in the project directory
    Then I should see either authentication error or connection error
    And the command should exit with non-zero status

  Scenario: Publish with registry URL override
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "rules.mdc" with content "# Override Test Rules"
    When I run "rfh pack rules.mdc --package=override-test" in the project directory
    And I run "rfh publish --registry http://localhost:9999 --token dummy-token" in the project directory
    Then I should see either authentication error or connection error
    And the command should exit with non-zero status

  Scenario: Publish shows proper workflow guidance
    Given I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    When I run "rfh publish" in the project directory
    Then I should see "no staged archives found"
    And I should see "Use 'rfh pack' to create archives first"
    And the command should exit with non-zero status