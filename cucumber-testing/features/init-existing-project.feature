Feature: Initialize RFH project with existing files
  As a developer
  I want to understand how RFH init behaves with existing files
  So that I can work with existing projects correctly

  Background:
    Given RFH is installed and accessible
    And I am in an empty directory

  Scenario: Initialize in directory with existing rulestack.json
    Given a file "rulestack.json" already exists
    When I run "rfh init"
    Then I should see "RuleStack project already initialized"
    And I should see "Use --force to reinitialize"
    And the existing "rulestack.json" should not be overwritten

  Scenario: Force flag overwrites existing project
    Given a file "rulestack.json" already exists
    When I run "rfh init --force --package"
    Then I should see "Initialized RuleStack project"
    And the manifest should have name "example-rules"
    And a file "CLAUDE.md" should be created
    And a directory ".rulestack" should be created

  Scenario: Initialize with existing CLAUDE.md file
    Given a file "CLAUDE.md" already exists
    When I run "rfh init"
    Then I should see "Initialized RuleStack project"
    And a file "rulestack.json" should be created
    And a directory ".rulestack" should be created