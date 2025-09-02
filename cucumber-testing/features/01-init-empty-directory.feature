Feature: Initialize RFH project in empty directory
  As a developer
  I want to initialize a new RuleStack project in an empty directory
  So that I can start creating and managing rule packages

  Background:
    Given RFH is installed and accessible
    And I am in an empty directory

  Scenario: Basic initialization works correctly
    When I run "rfh init"
    Then I should see "Initialized RuleStack project"
    And a file "rulestack.json" should be created
    And the project manifest should be created correctly
    And a directory ".rulestack" should be created
    And a file "CLAUDE.md" should be created
    And core rules should be downloaded to ".rulestack/core.v1.0.0"

  Scenario: Verify project manifest structure
    When I run "rfh init"
    Then the "rulestack.json" file should be valid JSON
    And the project manifest should contain:
      | field        | value |
      | version      | 1.0.0 |
      | dependencies | {}    |
    And the manifest should have the project structure:
      ```json
      {
        "version": "1.0.0",
        "dependencies": {}
      }
      ```

  Scenario: Verify directory structure after init
    When I run "rfh init"
    Then the following files and directories should exist:
      | path                                      | type      |
      | rulestack.json                           | file      |
      | CLAUDE.md                                | file      |
      | .rulestack/                              | directory |
      | .rulestack/core.v1.0.0/                  | directory |
      | .rulestack/core.v1.0.0/core_rules.md    | file      |

  Scenario: Verify command help output
    When I run "rfh init --help"
    Then I should see "force"
    And I should see "help"
    And I should not see "name"
    And I should not see "migrate"

  Scenario: Verify success message format
    When I run "rfh init"
    Then I should see output containing:
      | message                                    |
      | ✅ Initialized RuleStack project          |
      | 📁 Created:                               |
      | rulestack.json (project manifest)         |
      | CLAUDE.md (Claude Code integration)       |
      | .rulestack/ (dependency directory)        |
      | 🚀 Next steps:                            |