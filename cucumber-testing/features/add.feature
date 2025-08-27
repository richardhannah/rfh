Feature: Add Package Command
  As a developer using RFH
  I want to add (download and install) packages from a registry
  So that I can use rule packages created by others

  Background:
    Given RFH is installed and accessible
    And I have a clean config file
    And I have a registry "test-registry" configured at "http://localhost:8080"
    And "test-registry" is the active registry
    And I have a temporary project directory
    And RFH is initialized in the directory for add tests

  # Basic functionality tests

  Scenario: Add command is available
    When I run "rfh add --help" in the project directory
    Then I should see "Download and add a ruleset package"
    And I should see "Usage:"
    And I should see "rfh add <package@version>"
    And I should see "Global Flags:"
    And I should see "-v, --verbose"
    And the command should exit with zero status

  Scenario: Add requires package@version argument
    When I run "rfh add" in the project directory
    Then I should see "Error: accepts 1 arg(s), received 0"
    And the command should exit with non-zero status

  Scenario: Add existing package successfully
    When I run "rfh add security-rules@1.0.1" in the project directory
    Then I should see "✅ Successfully added security-rules@1.0.1"
    And the package should be downloaded to ".rulestack/security-rules.1.0.1/"
    And "rulestack.json" should contain dependency "security-rules": "1.0.1"
    And "rulestack.lock.json" should contain package "security-rules" with version "1.0.1"
    And the command should exit with zero status

  Scenario: Add with verbose output
    When I run "rfh add example-rules@0.1.0 --verbose" in the project directory
    Then I should see "RFH version: 1.0.0"
    And I should see "📦 Adding package: example-rules@0.1.0"
    And I should see "📁 Project root:"
    And I should see "✅ Successfully added example-rules@0.1.0"
    And the command should exit with zero status

  # Error scenarios

  Scenario: Add package with invalid format
    When I run "rfh add invalid-package" in the project directory
    Then I should see "version must be specified: use package@version format"
    And the command should exit with non-zero status

  Scenario: Add package with empty name
    When I run "rfh add @1.0.0" in the project directory
    Then I should see "scoped packages are not supported"
    And the command should exit with non-zero status

  Scenario: Add package with empty version
    When I run "rfh add package@" in the project directory
    Then I should see "package version cannot be empty"
    And the command should exit with non-zero status

  Scenario: Add non-existent package
    When I run "rfh add nonexistent-package@1.0.0" in the project directory
    Then I should see "failed to get package version"
    And the command should exit with non-zero status

  Scenario: Add package with no registry configured
    Given I have a truly clean config with no registries
    When I run "rfh add some-package@1.0.0" in the project directory
    Then I should see "no registry configured"
    And I should see "Use 'rfh registry add' to add a registry"
    And the command should exit with non-zero status

  Scenario: Add package outside RFH project
    Given I have a directory with no rulestack.json
    When I run "rfh add some-package@1.0.0" in that directory
    Then I should see "no RuleStack project found"
    And I should see "Run 'rfh init' first to initialize a project"
    And the command should exit with non-zero status

  Scenario: Add existing package prompts for confirmation (decline)
    Given I have already added package "example-rules@0.1.0"
    When I run "rfh add example-rules@0.1.0" with input "n" in the project directory
    Then I should see "⚠️  Package example-rules already exists"
    And I should see "Do you want to reinstall it? (y/N):"
    And I should see "⏭️  Skipping example-rules"
    And the command should exit with zero status

  Scenario: Add existing package prompts for confirmation (accept)
    Given I have already added package "example-rules@0.1.0"
    When I run "rfh add example-rules@0.1.0" with input "y" in the project directory
    Then I should see "⚠️  Package example-rules already exists"
    And I should see "Do you want to reinstall it? (y/N):"
    And I should see "✅ Successfully added example-rules@0.1.0"
    And the command should exit with zero status