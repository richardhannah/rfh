Feature: Generate simple package names without scopes
  As a developer
  I want RFH to generate simple package names without scope characters
  So that the scope removal initiative is properly implemented

  Background:
    Given I am in an empty directory
    And RFH is installed and accessible

  Scenario: Default initialization uses simple name
    When I run "rfh init"
    Then the default package name should be "example-rules"
    And the package name should not be "@acme/example-rules"
    And no scope characters "@" or "/" should appear in the manifest

  Scenario: Verify no legacy scoped names in output
    When I run "rfh init"
    Then I should not see "@acme" anywhere in the output
    And I should not see any scoped package names in the output
    And all example commands should use simple names

  Scenario: Manifest validation against scoped names
    When I run "rfh init"
    And I read the generated "rulestack.json"
    Then the JSON should not contain the string "@acme"
    And the JSON should not contain any "@" characters in name fields
    And the JSON should not contain any "/" characters in name fields

  Scenario: Core dependencies use simple names
    When I run "rfh init"
    Then the ".rulestack/core.v1.0.0/" directory should exist
    And the core package should not have scoped names
    And any dependency references should use simple names

  Scenario: Instructions and next steps use simple names
    When I run "rfh init"
    Then the "Next steps" output should not contain scoped examples
    And any example commands should use format "rfh add <simple-name>"
    And no examples should show "@org/package" format

  Scenario: CLAUDE.md integration file uses simple names
    When I run "rfh init"
    And I read the generated "CLAUDE.md" file
    Then it should not contain any scoped package examples
    And any package references should use simple naming format
    And example commands should demonstrate simple package names

  Scenario: Consistent naming across all generated files
    When I run "rfh init"
    Then all generated files should use consistent simple naming
    And no file should reference scoped packages
    And the naming should be consistent between:
      | file              | location                    |
      | rulestack.json    | manifest name field         |
      | CLAUDE.md         | example package references  |
      | core rules        | dependency examples         |

  Scenario: Validation prevents creation of scoped names
    When I attempt to manually create a manifest with scoped name "@test/rules"
    And I run any RFH command that validates the manifest
    Then I should see a warning about scoped names being deprecated
    And I should see "Consider updating to simple name: test-rules"

  Scenario: Migration hints for existing scoped projects
    Given a manifest exists with scoped name "@myorg/my-rules"
    When I run "rfh init --migrate"
    Then I should see "Migrating from scoped name"
    And I should see "Suggested new name: myorg-my-rules"
    And the manifest should be updated to use the simple name