Feature: Package Creation
  As a developer
  I want to pack my ruleset files into a distributable archive
  So that I can publish them to a registry

  Background:
    Given RFH is installed and accessible

  # Basic functionality tests
  
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

  Scenario: Pack command help text
    When I run "rfh pack --help"
    Then I should see "Creates a tar.gz archive containing ruleset files"
    And I should see "--file string       override single file to pack"
    And I should see "--output string     output archive path"
    And I should see "--package string    package name (enables non-interactive mode)"

  # Creating new packages
  
  Scenario: Pack with existing rulestack.json manifest
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "test-rules" and version "1.0.0"
    And I have a rule file "security-rules.mdc" with content "# Security Rules"
    When I run "rfh pack security-rules.mdc --package=test-rules" in the project directory
    Then I should see "✅ Created new package: test-rules v1.0.0"
    And I should see "📦 Archive: .rulestack/staged/test-rules-1.0.0.tgz"
    And the archive file ".rulestack/staged/test-rules-1.0.0.tgz" should exist
    And the command should exit with zero status

  Scenario: Pack single file with auto-manifest creation
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "my-security-rules.mdc" with content "# My Security Rules"
    When I run "rfh pack my-security-rules.mdc --package=my-security-rules --verbose" in the project directory
    Then I should see "RFH version: 1.0.0"
    And I should see "✅ Created new package: my-security-rules v1.0.0"
    And the archive file ".rulestack/staged/my-security-rules-1.0.0.tgz" should exist

  Scenario: Pack command non-interactive mode - create new package with directory structure
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

  # Updating existing packages
  
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

  # Version management
  
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

  # Output options
  
  Scenario: Pack with custom output path
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "output-test" and version "1.0.0"
    And I have a rule file "rules.mdc" with content "# Output Test Rules"
    When I run "rfh pack rules.mdc --package=output-test --output custom-output.tgz" in the project directory
    Then I should see "✅ Created new package: output-test v1.0.0"
    And I should see "📦 Archive: .rulestack/staged/output-test-1.0.0.tgz"
    And the archive file ".rulestack/staged/output-test-1.0.0.tgz" should exist

  Scenario: Pack verbose output
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "verbose-test" and version "1.0.0"
    And I have a rule file "verbose-rules.mdc" with content "# Verbose Rules"
    When I run "rfh pack verbose-rules.mdc --package=verbose-test --verbose" in the project directory
    Then I should see "RFH version: 1.0.0"
    And I should see "✅ Created new package: verbose-test v1.0.0"

  # Error cases
  
  Scenario: Pack with missing manifest
    Given I have a temporary project directory
    And I have a rule file "orphan-rules.mdc" with content "# Orphan Rules"
    When I run "rfh pack orphan-rules.mdc" in the project directory
    Then I should see "failed to load manifest"
    And I should see a file not found error
    And the command should exit with non-zero status

  Scenario: Pack with missing file in manifest
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "broken-rules" and version "1.0.0"
    And the manifest includes file "missing-file.mdc"
    And I have a rule file "exists.mdc" with content "# Exists"
    When I run "rfh pack exists.mdc --package=broken-rules" in the project directory
    Then I should see "✅ Created new package: broken-rules v1.0.0"
    And the archive file ".rulestack/staged/broken-rules-1.0.0.tgz" should exist
    And the command should exit with zero status

  Scenario: Pack command non-interactive mode - package not found
    Given I have a temporary project directory
    And RFH is initialized in the directory
    And I have a rule file "security-rule.mdc" with content "# Security Rule"
    When I run "rfh pack security-rule.mdc --package=nonexistent-package --version=1.0.1 --add-to-existing" in the project directory
    Then I should see "package 'nonexistent-package' not found in manifest"
    And the command should exit with non-zero status

  # Multi-package management
  
  Scenario: Pack command creates staging directory structure for multiple packages
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

  # Cross-directory operations
  
  Scenario: Pack from different project directory
    Given I have a temporary project directory at "remote-project"
    And I have a rulestack.json manifest with name "remote-rules" and version "1.5.0" in "remote-project"
    And I have a rule file "remote.mdc" with content "# Remote Rules" in "remote-project"
    When I run "rfh pack remote.mdc --package=remote-rules --verbose" in the "remote-project" directory
    Then I should see "RFH version: 1.0.0"
    And I should see "✅ Created new package: remote-rules v1.0.0"