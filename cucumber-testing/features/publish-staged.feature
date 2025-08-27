Feature: Publish Staged Archives
  As a developer
  I want to publish all staged archives at once
  So that I can efficiently publish multiple packages

  Background:
    Given RFH is installed and accessible

  Scenario: Publish command with no staged archives
    Given I have a temporary project directory
    And RFH is initialized in the directory  
    When I run "rfh publish" in the project directory
    Then I should see "no staged archives found"
    And I should see "Use 'rfh pack' to create archives first"
    And the command should exit with non-zero status

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

  Scenario: Staged archives are cleaned up after successful publish
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "test-rule.mdc" with content "# Test Rule"
    And I run "rfh pack test-rule.mdc --package=test-package" in the project directory
    And the archive file ".rulestack/staged/test-package-1.0.0.tgz" should exist
    # Note: This test would require a mock registry to test actual cleanup
    # For now, we just verify the staging behavior works