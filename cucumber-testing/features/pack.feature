Feature: Package Creation
  As a developer
  I want to pack my ruleset files into a distributable archive
  So that I can publish them to a registry

  Background:
    Given RFH is installed and accessible

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

  Scenario: Pack with custom output path
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "output-test" and version "1.0.0"
    And I have a rule file "rules.mdc" with content "# Output Test Rules"
    When I run "rfh pack rules.mdc --package=output-test --output custom-output.tgz" in the project directory
    Then I should see "✅ Created new package: output-test v1.0.0"
    And I should see "📦 Archive: .rulestack/staged/output-test-1.0.0.tgz"
    And the archive file ".rulestack/staged/output-test-1.0.0.tgz" should exist

  Scenario: Pack from different project directory
    Given I have a temporary project directory at "remote-project"
    And I have a rulestack.json manifest with name "remote-rules" and version "1.5.0" in "remote-project"
    And I have a rule file "remote.mdc" with content "# Remote Rules" in "remote-project"
    When I run "rfh pack remote-project/remote.mdc --package=remote-rules --verbose"
    Then I should see "RFH version: 1.0.0"
    And I should see "✅ Created new package: remote-rules v1.5.0"

  Scenario: Pack command help text
    When I run "rfh pack --help"
    Then I should see "Creates a tar.gz archive containing ruleset files"
    And I should see "--file string       override single file to pack"
    And I should see "--output string     output archive path"
    And I should see "--package string    package name (enables non-interactive mode)"

  Scenario: Pack with missing manifest
    Given I have a temporary project directory
    And I have a rule file "orphan-rules.mdc" with content "# Orphan Rules"
    When I run "rfh pack orphan-rules.mdc" in the project directory
    Then I should see "failed to load manifest"
    And I should see "no such file or directory"
    And the command should exit with non-zero status

  Scenario: Pack with missing file in manifest
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "broken-rules" and version "1.0.0"
    And the manifest includes file "missing-file.mdc"
    And I have a rule file "exists.mdc" with content "# Exists"
    When I run "rfh pack exists.mdc --package=broken-rules" in the project directory
    Then I should see "failed to pack files"
    And I should see "no files matched the specified patterns"
    And the command should exit with non-zero status

  Scenario: Pack verbose output
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "verbose-test" and version "1.0.0"
    And I have a rule file "verbose-rules.mdc" with content "# Verbose Rules"
    When I run "rfh pack verbose-rules.mdc --package=verbose-test --verbose" in the project directory
    Then I should see "RFH version: 1.0.0"
    And I should see "Config file:"
    And I should see "✅ Created new package: verbose-test v1.0.0"