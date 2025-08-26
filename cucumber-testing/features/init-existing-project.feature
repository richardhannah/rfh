Feature: Initialize in directory with existing project
  As a developer
  I want to be warned when trying to initialize in a directory with an existing project
  So that I don't accidentally overwrite my work

  Background:
    Given I am in a directory
    And RFH is installed and accessible

  Scenario: Initialize in directory with existing rulestack.json
    Given a file "rulestack.json" already exists with content:
      ```json
      {
        "name": "my-existing-rules",
        "version": "1.0.0",
        "description": "My existing ruleset"
      }
      ```
    When I run "rfh init"
    Then I should see a warning about existing project
    And I should see "RuleStack project already exists"
    And the existing "rulestack.json" should not be overwritten
    And the original content should remain unchanged

  Scenario: Initialize with force flag overwrites existing project
    Given a file "rulestack.json" already exists
    When I run "rfh init --force"
    Then I should see "Overwriting existing project"
    And the "rulestack.json" should be replaced with default content
    And the manifest should have name "example-rules"

  Scenario: Initialize in directory with partial project files
    Given a file "CLAUDE.md" already exists
    And no "rulestack.json" exists
    When I run "rfh init"
    Then I should see "Some project files already exist"
    And a "rulestack.json" should be created
    And the existing "CLAUDE.md" should not be overwritten
    And the ".rulestack/" directory should be created

  Scenario: Initialize with existing .rulestack directory
    Given a directory ".rulestack" already exists with some files
    When I run "rfh init"
    Then the existing ".rulestack" directory should not be deleted
    And core rules should still be downloaded to ".rulestack/core.v1.0.0/"
    And existing files in ".rulestack" should be preserved

  Scenario: Prompt for confirmation on existing project
    Given a complete RuleStack project already exists
    When I run "rfh init" interactively
    Then I should be prompted "Project already exists. Overwrite? (y/N)"
    When I respond "n"
    Then the command should exit without changes
    And I should see "Initialization cancelled"

  Scenario: Confirm overwrite on existing project
    Given a complete RuleStack project already exists
    When I run "rfh init" interactively
    And I am prompted "Project already exists. Overwrite? (y/N)"
    And I respond "y"
    Then the existing project should be overwritten
    And I should see "Initialized RuleStack project"