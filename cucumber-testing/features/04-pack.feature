Feature: Package Creation
  As a developer
  I want to pack my ruleset files into a distributable archive
  So that I can publish them to a registry

  Background:
    Given I am in an empty directory
    And RFH is installed and accessible
    And RFH is initialized in the directory for package creation

  # Basic functionality tests
  
  Scenario: Pack command with no arguments should error
    When I run "rfh pack" in the project directory
    Then I should see "accepts 1 arg(s), received 0"
    And the command should exit with non-zero status

  Scenario: Pack command with invalid file extension should error
    Given I have a rule file "test-rule.txt" with content "# Test Rule"
    When I run "rfh pack test-rule.txt" in the project directory
    Then I should see "file must be a valid .mdc file"
    And the command should exit with non-zero status

  Scenario: Pack command help text
    When I run "rfh pack --help"
    Then I should see "Creates a tar.gz archive containing ruleset files"
    And I should see "-f, --file string      override single file to pack"
    And I should see "-o, --output string    output archive path"
    And I should see "-p, --package string   package name (enables non-interactive mode)"

  # Creating new packages
  
  Scenario: Pack with existing rulestack.json manifest
    Given I have a rulestack.json manifest with name "test-rules" and version "1.0.0"
    And I have a rule file "security-rules.mdc" with content "# Security Rules"
    When I run "rfh pack security-rules.mdc --package=test-rules" in the project directory
    Then I should see "✅ Created new package: test-rules v1.0.0"
    And I should see "📦 Archive: .rulestack/staged/test-rules-1.0.0.tgz"
    And the archive file ".rulestack/staged/test-rules-1.0.0.tgz" should exist
    And the command should exit with zero status

  Scenario: Pack single file with auto-manifest creation
    Given RFH is initialized in the directory
    And I have a rule file "my-security-rules.mdc" with content "# My Security Rules"
    When I run "rfh pack my-security-rules.mdc --package=my-security-rules" in the project directory
    Then I should see "✅ Created new package: my-security-rules v1.0.0"
    And the archive file ".rulestack/staged/my-security-rules-1.0.0.tgz" should exist

  Scenario: Pack command non-interactive mode - create new package with directory structure
    Given RFH is initialized in the directory
    And I have a rule file "security-rule.mdc" with content "# Security Rule\nNever hardcode passwords."
    When I run "rfh pack security-rule.mdc --package=security-rules" in the project directory
    Then I should see "✅ Created new package: security-rules v1.0.0"
    And I should see "📁 Package directory: .rulestack/security-rules.1.0.0"
    And I should see "📦 Archive: .rulestack/staged/security-rules-1.0.0.tgz"
    And the archive file ".rulestack/staged/security-rules-1.0.0.tgz" should exist
    And the directory ".rulestack/security-rules.1.0.0" should exist
    And the command should exit with zero status

  # Version management

  # Output options
  
  Scenario: Pack with custom output path
    Given I have a rulestack.json manifest with name "output-test" and version "1.0.0"
    And I have a rule file "rules.mdc" with content "# Output Test Rules"
    When I run "rfh pack rules.mdc --package=output-test --output custom-output.tgz" in the project directory
    Then I should see "✅ Created new package: output-test v1.0.0"
    And I should see "📦 Archive: .rulestack/staged/output-test-1.0.0.tgz"
    And the archive file ".rulestack/staged/output-test-1.0.0.tgz" should exist

  # Error cases
  
  Scenario: Pack without package flag requires interactive input
    Given I have a rule file "orphan-rules.mdc" with content "# Orphan Rules"
    When I run "rfh pack orphan-rules.mdc" in the project directory
    Then I should see "failed to read input"
    And the command should exit with non-zero status

  Scenario: Pack with missing file in manifest
    Given I have a rulestack.json manifest with name "broken-rules" and version "1.0.0"
    And the manifest includes file "missing-file.mdc"
    And I have a rule file "exists.mdc" with content "# Exists"
    When I run "rfh pack exists.mdc --package=broken-rules" in the project directory
    Then I should see "✅ Created new package: broken-rules v1.0.0"
    And the archive file ".rulestack/staged/broken-rules-1.0.0.tgz" should exist
    And the command should exit with zero status

  # Multi-package management
  
  Scenario: Pack command creates staging directory structure for multiple packages
    Given RFH is initialized in the directory
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
    When I run "rfh pack remote.mdc --package=remote-rules" in the "remote-project" directory
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
    When I run "rfh pack test-rule.mdc --package=test-package" in the project directory
    And I run "rfh status" in the project directory
    Then I should see "test-package-1.0.0.tgz"
    And the command should exit with zero status

  Scenario: Status with multiple staged packages
    Given I am in an empty directory
    And RFH is initialized in the directory
    And I have a rule file "rule1.mdc" with content "# Rule 1"
    And I have a rule file "rule2.mdc" with content "# Rule 2"
    When I run "rfh pack rule1.mdc --package=package1" in the project directory
    And I run "rfh pack rule2.mdc --package=package2" in the project directory
    And I run "rfh status" in the project directory
    Then I should see "package1-1.0.0.tgz"
    And I should see "package2-1.0.0.tgz"
    And the command should exit with zero status