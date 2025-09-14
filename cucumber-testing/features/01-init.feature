Feature: Initialize RFH project
  As a developer
  I want to initialize RuleStack projects
  So that I can manage dependencies and create packages

  Background:
    Given RFH is installed and accessible
    And I am in an empty directory

  Scenario: Basic initialization in empty directory
    When I run "rfh init"
    Then I should see "Initialized RuleStack project"
    And a file "rulestack.json" should be created
    And a directory ".rulestack" should be created
    And a file "CLAUDE.md" should be created
    And core rules should be downloaded to ".rulestack/core.v1.0.0"
    And the project manifest should contain:
      | field        | value |
      | version      | 1.0.0 |
      | dependencies | {}    |

  Scenario: Cannot initialize when project already exists
    Given a file "rulestack.json" already exists
    When I run "rfh init"
    Then I should see "RuleStack project already initialized"
    And I should see "Use --force to reinitialize"
    And the existing "rulestack.json" should not be overwritten

  Scenario: Force reinitialize overwrites existing project
    Given a file "rulestack.json" already exists
    When I run "rfh init --force"
    Then I should see "Initialized RuleStack project"
    And the project manifest should be created correctly
    And a file "CLAUDE.md" should be created

  Scenario: Initialize preserves existing CLAUDE.md
    Given a file "CLAUDE.md" already exists
    When I run "rfh init"
    Then I should see "Initialized RuleStack project"
    And a file "rulestack.json" should be created
    And a directory ".rulestack" should be created

  Scenario: Verify complete project structure after init
    When I run "rfh init"
    Then the following files and directories should exist:
      | path                                  | type      |
      | rulestack.json                       | file      |
      | CLAUDE.md                            | file      |
      | .rulestack/                          | directory |
      | .rulestack/core.v1.0.0/              | directory |
      | .rulestack/core.v1.0.0/core_rules.md | file      |