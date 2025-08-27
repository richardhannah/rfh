Feature: Initialize RFH project in project mode (default)
  As a developer
  I want to initialize a new RuleStack project for dependency management
  So that I can add packages created by others

  Background:
    Given I am in an empty directory
    And RFH is installed and accessible

  Scenario: Basic project initialization works correctly
    When I run "rfh init"
    Then I should see "Initialized RuleStack project"
    And I should see "Creating project manifest for dependency management"
    And a file "rulestack.json" should be created
    And the project manifest should have version "1.0.0"
    And the project manifest should have projectRoot
    And the project manifest should have empty dependencies
    And a directory ".rulestack" should be created
    And a file "CLAUDE.md" should be created

  Scenario: Project manifest has correct structure
    When I run "rfh init"
    Then the "rulestack.json" file should be valid JSON
    And the project manifest should contain:
      | field         | value   |
      | version       | 1.0.0   |
      | dependencies  | {}      |
    And the manifest should be a valid project manifest for add command