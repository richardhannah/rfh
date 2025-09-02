Feature: Install Command
  As a RuleStack user
  I want to install all packages from my project manifest
  So that I can ensure my project dependencies are up-to-date

  Background:
    Given I am in an empty directory
    And RFH is installed and accessible

  Scenario: Install missing packages from empty project
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name    | version |
      | security-rules  | 1.0.0   |
      | logging-rules   | 2.1.0   |
    And the test registry is configured
    And test packages are available
    And I am logged in as a user
    When I run "rfh install ."
    Then I should see "Installation Summary:"
    And I should see "security-rules@1.0.0 → installed successfully"
    And I should see "logging-rules@2.1.0 → installed successfully"
    And I should see "Summary: 2 installed, 0 updated, 0 skipped, 0 failed"
    And the command should exit with zero status

  Scenario: Skip packages that are already up-to-date
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name    | version |
      | security-rules  | 1.0.0   |
    And the test registry is configured
    And test packages are available
    And I am logged in as a user
    And I have already installed "security-rules" version "1.0.0"
    When I run "rfh install ."
    Then I should see "Installation Summary:"
    And I should see "security-rules@1.0.0 → Already up-to-date"
    And I should see "Summary: 0 installed, 0 updated, 1 skipped, 0 failed"
    And the command should exit with zero status

  Scenario: Update packages to higher versions
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name    | version |
      | security-rules  | 1.2.0   |
    And the test registry is configured  
    And test packages are available
    And I am logged in as a user
    And I have already installed "security-rules" version "1.0.0"
    When I run "rfh install ."
    Then I should see "Installation Summary:"
    And I should see "security-rules@1.2.0 → Updated from 1.0.0"
    And I should see "Summary: 0 installed, 1 updated, 0 skipped, 0 failed"
    And the command should exit with zero status

  Scenario: Skip packages when installed version is newer
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name    | version |
      | security-rules  | 1.0.0   |
    And the test registry is configured
    And test packages are available  
    And I am logged in as a user
    And I have already installed "security-rules" version "1.2.0"
    When I run "rfh install ."
    Then I should see "Installation Summary:"
    And I should see "security-rules@1.0.0 → Installed version 1.2.0 is newer than required 1.0.0"
    And I should see "Summary: 0 installed, 0 updated, 1 skipped, 0 failed"
    And the command should exit with zero status

  Scenario: Handle mixed package states
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name      | version |
      | security-rules    | 1.2.0   |
      | logging-rules     | 2.0.0   |
      | best-practices    | 1.0.1   |
      | new-package       | 1.0.0   |
    And the test registry is configured
    And test packages are available
    And I am logged in as a user
    And I have already installed "security-rules" version "1.0.0"
    And I have already installed "logging-rules" version "2.0.0"
    And I have already installed "best-practices" version "1.5.0"
    When I run "rfh install ."
    Then I should see "Installation Summary:"
    And I should see "security-rules@1.2.0 → Updated from 1.0.0"
    And I should see "logging-rules@2.0.0 → Already up-to-date"
    And I should see "best-practices@1.0.1 → Installed version 1.5.0 is newer than required 1.0.1"
    And I should see "new-package@1.0.0 → installed successfully"
    And I should see "Summary: 1 installed, 1 updated, 2 skipped, 0 failed"
    And the command should exit with zero status

  Scenario: Continue processing after package failure
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name      | version |
      | security-rules    | 1.0.0   |
      | nonexistent-pkg   | 1.0.0   |
      | logging-rules     | 2.0.0   |
    And the test registry is configured
    And test packages are available
    And I am logged in as a user
    When I run "rfh install ."
    Then I should see "Installation Summary:"
    And I should see "security-rules@1.0.0 → installed successfully"
    And I should see "nonexistent-pkg@1.0.0 → failed"
    And I should see "logging-rules@2.0.0 → installed successfully"
    And I should see "Summary: 2 installed, 0 updated, 0 skipped, 1 failed"
    And I should see "Some packages failed to install"
    And the command should exit with zero status

  Scenario: Install command with no dependencies
    Given I have initialized an RFH project
    And I have a project manifest with no dependencies
    When I run "rfh install ."
    Then I should see "No dependencies found in rulestack.json"
    And the command should exit with zero status

  Scenario: Install command without project manifest
    Given I am in an empty directory
    When I run "rfh install ."
    Then I should see "failed to find project root"
    And the command should exit with non-zero status

  Scenario: Install command with invalid argument
    Given I have initialized an RFH project
    When I run "rfh install something-else"
    Then I should see "only '.' is supported"
    And the command should exit with non-zero status

  Scenario: Install command without authentication
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name    | version |
      | security-rules  | 1.0.0   |
    When I run "rfh install ."
    Then I should see "no registry configured"
    And the command should exit with non-zero status

  Scenario: Verbose output shows detailed progress
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name    | version |
      | security-rules  | 1.0.0   |
    And the test registry is configured
    And test packages are available
    And I am logged in as a user
    When I run "rfh install . --verbose"
    Then I should see "Installing packages from project manifest"
    And I should see "Project root:"
    And I should see "Installing security-rules@1.0.0"
    And I should see "Installation Summary:"
    And the command should exit with zero status

  Scenario: Install updates project manifests and CLAUDE.md
    Given I have initialized an RFH project
    And I have a project manifest with the following dependencies:
      | package-name    | version |
      | security-rules  | 1.0.0   |
    And the test registry is configured
    And test packages are available
    And I am logged in as a user
    When I run "rfh install ."
    Then I should see "security-rules@1.0.0 → installed successfully"
    And "rulestack.json" should contain dependency "security-rules": "1.0.0"
    And "rulestack.lock.json" should contain package "security-rules" with version "1.0.0"
    And the package should be downloaded to ".rulestack/security-rules.1.0.0/"