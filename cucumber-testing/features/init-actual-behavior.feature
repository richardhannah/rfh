Feature: RFH Init - Actual Current Behavior
  As a developer
  I want to test the actual current behavior of rfh init
  So that I can validate what currently works

  Background:
    Given I am in an empty directory
    And RFH is installed and accessible

  Scenario: Basic initialization works correctly
    When I run "rfh init"
    Then I should see "Initialized RuleStack project"
    And a file "rulestack.json" should be created
    And the default package name should be "example-rules"
    And the package name should not be "@acme/example-rules"
    And no scope characters "@" or "/" should appear in the manifest
    And a directory ".rulestack" should be created
    And a file "CLAUDE.md" should be created
    And core rules should be downloaded to ".rulestack/core.v1.0.0"

  Scenario: Force flag works for overwriting existing project
    Given a file "rulestack.json" already exists
    When I run "rfh init --force"
    Then I should see "Initialized RuleStack project"
    And the manifest should have name "example-rules"

  Scenario: Scope removal is properly implemented
    When I run "rfh init"
    Then I should not see "@acme" anywhere in the output
    And the manifest should not contain scope characters "@" or "/"
    And the default package name should be "example-rules"

  Scenario: Verify available flags match help output
    When I run "rfh init --help"
    Then I should see "force"
    And I should see "help"
    And I should not see "name"
    And I should not see "migrate"