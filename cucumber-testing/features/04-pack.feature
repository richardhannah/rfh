Feature: Package Creation
  As a developer
  I want to pack my ruleset files into a distributable archive
  So that I can publish them to a registry

  Background:
    Given I am in an empty directory
    And RFH is installed and accessible
    And RFH is initialized in the directory for package creation

  # Basic functionality tests
  
  Scenario: Pack command with no --file flag should error
    When I run "rfh pack" in the project directory
    Then I should see "--file flag is required"
    And the command should exit with non-zero status

  Scenario: Pack command with invalid file extension should error
    Given I have a rule file "test-rule.txt" with content "# Test Rule"
    When I run "rfh pack --file=test-rule.txt" in the project directory
    Then I should see "file must be a valid .mdc file"
    And the command should exit with non-zero status

  Scenario: Pack command help text
    When I run "rfh pack --help"
    Then I should see "Creates a tar.gz archive containing ruleset files"
    And I should see "-f, --file string      .mdc file to pack (required)"
    And I should see "-o, --output string    output archive path"
    And I should see "-p, --package string   package name (enables non-interactive mode)"
    And I should see "--version string   package version (auto-increments for existing packages, defaults to 1.0.0 for new packages)"

  # Creating new packages
  
  Scenario: Pack with existing rulestack.json manifest
    Given I have a rulestack.json manifest with name "test-rules" and version "1.0.0"
    And I have a rule file "security-rules.mdc" with content "# Security Rules"
    When I run "rfh pack --file=security-rules.mdc --package=test-rules" in the project directory
    Then I should see "✅ Created new package: test-rules v1.0.0"
    And I should see "📦 Archive: .rulestack/staged/test-rules-1.0.0.tgz"
    And the archive file ".rulestack/staged/test-rules-1.0.0.tgz" should exist
    And the command should exit with zero status

  Scenario: Pack single file with auto-manifest creation
    Given RFH is initialized in the directory
    And I have a rule file "my-security-rules.mdc" with content "# My Security Rules"
    When I run "rfh pack --file=my-security-rules.mdc --package=my-security-rules" in the project directory
    Then I should see "✅ Created new package: my-security-rules v1.0.0"
    And the archive file ".rulestack/staged/my-security-rules-1.0.0.tgz" should exist

  Scenario: Pack command non-interactive mode - create new package with directory structure
    Given RFH is initialized in the directory
    And I have a rule file "security-rule.mdc" with content "# Security Rule\nNever hardcode passwords."
    When I run "rfh pack --file=security-rule.mdc --package=security-rules" in the project directory
    Then I should see "✅ Created new package: security-rules v1.0.0"
    And I should see "📁 Package directory: .rulestack/security-rules.1.0.0"
    And I should see "📦 Archive: .rulestack/staged/security-rules-1.0.0.tgz"
    And the archive file ".rulestack/staged/security-rules-1.0.0.tgz" should exist
    And the directory ".rulestack/security-rules.1.0.0" should exist
    And the command should exit with zero status

  # Version management

  Scenario: Pack with custom version number
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have a rule file "version-test.mdc" with content "# Version Test Rule"
    When I run "rfh pack --file=version-test.mdc --package=version-test --version=2.1.5" in the project directory
    Then I should see "✅ Created new package: version-test v2.1.5"
    And I should see "📦 Archive: .rulestack/staged/version-test-2.1.5.tgz"
    And the archive file ".rulestack/staged/version-test-2.1.5.tgz" should exist
    And the directory ".rulestack/version-test.2.1.5" should exist
    And the command should exit with zero status

  Scenario: Pack with default version when --version omitted
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have a rule file "default-version.mdc" with content "# Default Version Test"
    When I run "rfh pack --file=default-version.mdc --package=default-test" in the project directory
    Then I should see "✅ Created new package: default-test v1.0.0"
    And I should see "📦 Archive: .rulestack/staged/default-test-1.0.0.tgz"
    And the archive file ".rulestack/staged/default-test-1.0.0.tgz" should exist
    And the directory ".rulestack/default-test.1.0.0" should exist
    And the command should exit with zero status

  # Output options
  
  Scenario: Pack with custom output path
    Given I have a rulestack.json manifest with name "output-test" and version "1.0.0"
    And I have a rule file "rules.mdc" with content "# Output Test Rules"
    When I run "rfh pack --file=rules.mdc --package=output-test --output custom-output.tgz" in the project directory
    Then I should see "✅ Created new package: output-test v1.0.0"
    And I should see "📦 Archive: .rulestack/staged/output-test-1.0.0.tgz"
    And the archive file ".rulestack/staged/output-test-1.0.0.tgz" should exist

  # Error cases
  
  Scenario: Pack without package flag requires interactive input
    Given I have a rule file "orphan-rules.mdc" with content "# Orphan Rules"
    When I run "rfh pack --file=orphan-rules.mdc" in the project directory
    Then I should see "failed to read input"
    And the command should exit with non-zero status

  Scenario: Pack with missing file in manifest
    Given I have a rulestack.json manifest with name "broken-rules" and version "1.0.0"
    And the manifest includes file "missing-file.mdc"
    And I have a rule file "exists.mdc" with content "# Exists"
    When I run "rfh pack --file=exists.mdc --package=broken-rules" in the project directory
    Then I should see "✅ Created new package: broken-rules v1.0.0"
    And the archive file ".rulestack/staged/broken-rules-1.0.0.tgz" should exist
    And the command should exit with zero status

  # Multi-package management
  
  Scenario: Pack command creates staging directory structure for multiple packages
    Given RFH is initialized in the directory
    And I have a rule file "rule1.mdc" with content "# Rule 1"
    And I have a rule file "rule2.mdc" with content "# Rule 2"
    When I run "rfh pack --file=rule1.mdc --package=package1" in the project directory
    And I run "rfh pack --file=rule2.mdc --package=package2" in the project directory
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
    When I run "rfh pack --file=remote.mdc --package=remote-rules" in the "remote-project" directory
    Then I should see "✅ Created new package: remote-rules v1.0.0"

  # Status command tests
  
  Scenario: Status with no staged packages
    Given I am in an empty directory
    And RFH is initialized in the directory
    When I run "rfh status" in the project directory
    Then I should see "No staged packages found"
    And the command should exit with zero status

  Scenario: Status with single staged package
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have a rule file "test-rule.mdc" with content "# Test Rule"
    When I run "rfh pack --file=test-rule.mdc --package=test-package" in the project directory
    And I run "rfh status" in the project directory
    Then I should see "test-package-1.0.0.tgz"
    And the command should exit with zero status

  Scenario: Status with multiple staged packages
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have a rule file "rule1.mdc" with content "# Rule 1"
    And I have a rule file "rule2.mdc" with content "# Rule 2"
    When I run "rfh pack --file=rule1.mdc --package=package1" in the project directory
    And I run "rfh pack --file=rule2.mdc --package=package2" in the project directory
    And I run "rfh status" in the project directory
    Then I should see "package1-1.0.0.tgz"
    And I should see "package2-1.0.0.tgz"
    And the command should exit with zero status

  Scenario: Status shows correct filenames with custom versions
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have a rule file "custom-ver.mdc" with content "# Custom Version Rule"
    When I run "rfh pack --file=custom-ver.mdc --package=custom-version --version=3.2.1" in the project directory
    And I run "rfh status" in the project directory
    Then I should see "custom-version-3.2.1.tgz"
    And the command should exit with zero status

  # Enhanced Pack Functionality - Existing Package Updates

  Scenario: Pack adds file to existing package with explicit version
    Given I am in an empty directory
    And I have installed package "test-rules@1.0.0" containing file "existing.mdc"
    And I have a rule file "new-rule.mdc" with content "# New Rule"
    And I debug the current test environment state
    When I run "rfh pack --file=new-rule.mdc --package=test-rules --version=1.1.0" in the project directory
    Then I should see "📦 Found existing package test-rules@1.0.0 with 1 files"
    And I should see "✅ Updated existing package: test-rules v1.0.0 -> v1.1.0"
    And I should see "📋 Files included: existing.mdc, new-rule.mdc"
    And the staged archive should contain "existing.mdc"
    And the staged archive should contain "new-rule.mdc"
    And the staged archive should contain "rulestack.json"
    And the staged archive should be named "test-rules-1.1.0.tgz"
    And the command should exit with zero status

  Scenario: Pack auto-increments version for existing package
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have installed package "test-rules@1.2.3" containing file "existing.mdc"
    And I have a rule file "auto-increment.mdc" with content "# Auto Increment Test"
    When I run "rfh pack --file=auto-increment.mdc --package=test-rules" in the project directory
    Then I should see "📦 Found existing package test-rules@1.2.3 with 1 files"
    And I should see "🔄 Auto-incrementing version to 1.2.4"
    And I should see "✅ Updated existing package: test-rules v1.2.3 -> v1.2.4"
    And I should see "📋 Files included: existing.mdc, auto-increment.mdc"
    And the staged archive should be named "test-rules-1.2.4.tgz"
    And the command should exit with zero status

  Scenario: Pack fails on version decrease
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have installed package "test-rules@2.0.0" containing file "existing.mdc"
    And I have a rule file "downgrade.mdc" with content "# Downgrade Test"
    When I run "rfh pack --file=downgrade.mdc --package=test-rules --version=1.5.0" in the project directory
    Then I should see "📦 Found existing package test-rules@2.0.0 with 1 files"
    And I should see "Error: version validation failed: new version 1.5.0 must be greater than current version 2.0.0"
    And the command should exit with non-zero status

  Scenario: Pack fails on file name conflict
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have installed package "test-rules@1.0.0" containing file "conflict.mdc"
    And I have a rule file "conflict.mdc" with content "# Different content but same name"
    When I run "rfh pack --file=conflict.mdc --package=test-rules --version=1.1.0" in the project directory
    Then I should see "📦 Found existing package test-rules@1.0.0 with 1 files"
    And I should see "Error: file conflict.mdc already exists in package test-rules@1.0.0, use a different filename or increment version to replace"
    And the command should exit with non-zero status

  Scenario: Pack handles existing package with multiple files
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have installed package "multi-rules@1.0.0" containing files:
      | filename    | content           |
      | auth.mdc    | # Auth Rules      |
      | security.mdc| # Security Rules  |
      | logging.mdc | # Logging Rules   |
    And I have a rule file "validation.mdc" with content "# Validation Rules"
    When I run "rfh pack --file=validation.mdc --package=multi-rules --version=1.1.0" in the project directory
    Then I should see "📦 Found existing package multi-rules@1.0.0 with 3 files"
    And I should see "✅ Updated existing package: multi-rules v1.0.0 -> v1.1.0"
    And I should see "📋 Files included: auth.mdc, logging.mdc, security.mdc, validation.mdc"
    And the staged archive should contain "auth.mdc"
    And the staged archive should contain "security.mdc"
    And the staged archive should contain "logging.mdc"
    And the staged archive should contain "validation.mdc"
    And the staged archive should contain "rulestack.json"
    And the command should exit with zero status

  Scenario: Pack validates file format for existing packages
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have installed package "test-rules@1.0.0" containing file "existing.mdc"
    And I have a rule file "invalid.txt" with content "# Not an MDC file"
    When I run "rfh pack --file=invalid.txt --package=test-rules --version=1.1.0" in the project directory
    Then I should see "Error: file must be a valid .mdc file: invalid.txt"
    And the command should exit with non-zero status

  Scenario: Pack works with semantic versioning including pre-release
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have installed package "test-rules@1.0.0-alpha" containing file "existing.mdc"
    And I have a rule file "beta-feature.mdc" with content "# Beta Feature"
    When I run "rfh pack --file=beta-feature.mdc --package=test-rules --version=1.0.0" in the project directory
    Then I should see "📦 Found existing package test-rules@1.0.0-alpha with 1 files"
    And I should see "✅ Updated existing package: test-rules v1.0.0-alpha -> v1.0.0"
    And I should see "📋 Files included: existing.mdc, beta-feature.mdc"
    And the command should exit with zero status