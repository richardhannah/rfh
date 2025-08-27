Feature: New Pack Command Functionality
  As a developer
  I want to use the enhanced pack command with .mdc files
  So that I can manage packages interactively and non-interactively

  Background:
    Given RFH is installed and accessible

  Scenario: Pack command with no arguments should error
    Given I have a temporary project directory
    When I run "rfh pack" in the project directory
    Then I should see "accepts 1 arg(s), received 0"
    And the command should exit with non-zero status

  Scenario: Pack command with invalid file extension should error
    Given I have a temporary project directory
    And I have a rule file "test-rule.txt" with content "# Test Rule"
    When I run "rfh pack test-rule.txt" in the project directory
    Then I should see "file must be a valid .mdc file"
    And the command should exit with non-zero status

  Scenario: Pack command non-interactive mode - create new package
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "security-rule.mdc" with content "# Security Rule\nNever hardcode passwords."
    When I run "rfh pack security-rule.mdc --package=security-rules" in the project directory
    Then I should see "✅ Created new package: security-rules v1.0.0"
    And I should see "📁 Package directory: .rulestack/security-rules.1.0.0"
    And I should see "📦 Archive: .rulestack/staged/security-rules-1.0.0.tgz"
    And the archive file ".rulestack/staged/security-rules-1.0.0.tgz" should exist
    And the directory ".rulestack/security-rules.1.0.0" should exist
    And the rulestack.json should contain package "security-rules" with version "1.0.0"
    And the command should exit with zero status

  Scenario: Pack command non-interactive mode - add to existing package
    Given I have a temporary project directory  
    And RFH is initialized in the directory
    And I have a rule file "security-rule.mdc" with content "# Security Rule\nNever hardcode passwords."
    And I run "rfh pack security-rule.mdc --package=security-rules" in the project directory
    And I have a rule file "auth-rule.mdc" with content "# Auth Rule\nUse strong authentication."
    When I run "rfh pack auth-rule.mdc --package=security-rules --version=1.0.1 --add-to-existing" in the project directory
    Then I should see "Adding auth-rule.mdc to package: security-rules (v1.0.0 -> v1.0.1)"
    And I should see "✅ Updated package: security-rules v1.0.0 -> v1.0.1"
    And I should see "📁 Package directory: .rulestack/security-rules.1.0.1"
    And I should see "📦 Archive: .rulestack/staged/security-rules-1.0.1.tgz"
    And I should see "🗑️  Removed old archive: security-rules-1.0.0.tgz"
    And the archive file ".rulestack/staged/security-rules-1.0.1.tgz" should exist
    And the archive file ".rulestack/staged/security-rules-1.0.0.tgz" should not exist
    And the directory ".rulestack/security-rules.1.0.1" should exist
    And the directory ".rulestack/security-rules.1.0.0" should not exist
    And the rulestack.json should contain package "security-rules" with version "1.0.1"
    And the command should exit with zero status

  Scenario: Pack command version validation - reject downgrade
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "security-rule.mdc" with content "# Security Rule"
    And I run "rfh pack security-rule.mdc --package=security-rules" in the project directory
    And I have a rule file "auth-rule.mdc" with content "# Auth Rule"
    And I run "rfh pack auth-rule.mdc --package=security-rules --version=1.0.1 --add-to-existing" in the project directory
    And I have a rule file "network-rule.mdc" with content "# Network Rule"
    When I run "rfh pack network-rule.mdc --package=security-rules --version=1.0.0 --add-to-existing" in the project directory
    Then I should see "version validation failed"
    And I should see "new version 1.0.0 must be greater than current version 1.0.1"
    And the command should exit with non-zero status

  Scenario: Pack command non-interactive mode - missing version for add-to-existing
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "security-rule.mdc" with content "# Security Rule"
    And I run "rfh pack security-rule.mdc --package=security-rules" in the project directory
    And I have a rule file "auth-rule.mdc" with content "# Auth Rule"
    When I run "rfh pack auth-rule.mdc --package=security-rules --add-to-existing" in the project directory
    Then I should see "--version is required when using --add-to-existing"
    And the command should exit with non-zero status

  Scenario: Pack command non-interactive mode - package not found
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "security-rule.mdc" with content "# Security Rule"
    When I run "rfh pack security-rule.mdc --package=nonexistent-package --version=1.0.1 --add-to-existing" in the project directory
    Then I should see "package 'nonexistent-package' not found in manifest"
    And the command should exit with non-zero status

  Scenario: Pack command creates staging directory structure
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "rule1.mdc" with content "# Rule 1"
    And I have a rule file "rule2.mdc" with content "# Rule 2"
    When I run "rfh pack rule1.mdc --package=package1" in the project directory
    And I run "rfh pack rule2.mdc --package=package2" in the project directory
    Then the directory ".rulestack/staged" should exist
    And the archive file ".rulestack/staged/package1-1.0.0.tgz" should exist
    And the archive file ".rulestack/staged/package2-1.0.0.tgz" should exist
    And the directory ".rulestack/package1.1.0.0" should exist
    And the directory ".rulestack/package2.1.0.0" should exist