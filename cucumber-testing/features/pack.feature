Feature: Package Creation
  As a developer
  I want to pack my ruleset files into a distributable archive
  So that I can publish them to a registry

  Background:
    Given RFH is installed and accessible

  Scenario: Pack with existing rulestack.json manifest
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "test-rules" and version "1.0.0"
    And I have a rule file "security-rules.md" with content "# Security Rules"
    When I run "rfh pack" in the project directory
    Then I should see "✅ Successfully packed test-rules"
    And I should see "📦 Archive: test-rules-1.0.0.tgz"
    And the archive file "test-rules-1.0.0.tgz" should exist
    And the command should exit with zero status

  Scenario: Pack with custom manifest file
    Given I have a temporary project directory
    And I have a custom manifest "custom-rules.json" with name "custom-package" and version "2.1.0"
    And I have a rule file "rules.md" with content "# Custom Rules"
    When I run "rfh pack --manifest custom-rules.json" in the project directory
    Then I should see "✅ Successfully packed custom-package"
    And I should see "📦 Archive: custom-package-2.1.0.tgz"
    And the archive file "custom-package-2.1.0.tgz" should exist

  Scenario: Pack single file with auto-manifest creation
    Given I have a temporary project directory
    And I have a rule file "my-security-rules.md" with content "# My Security Rules"
    When I run "rfh pack --file my-security-rules.md --verbose" in the project directory
    Then I should see "📝 No manifest found, creating auto-manifest from file"
    And I should see "📦 Packing my-security-rules v1.0.0"
    And I should see "✅ Successfully packed my-security-rules"
    And the archive file "my-security-rules-1.0.0.tgz" should exist

  Scenario: Pack with custom output path
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "output-test" and version "1.0.0"
    And I have a rule file "rules.md" with content "# Output Test Rules"
    When I run "rfh pack --output custom-output.tgz" in the project directory
    Then I should see "✅ Successfully packed output-test"
    And I should see "📦 Archive: custom-output.tgz"
    And the archive file "custom-output.tgz" should exist

  Scenario: Pack from different project directory
    Given I have a temporary project directory at "remote-project"
    And I have a rulestack.json manifest with name "remote-rules" and version "1.5.0" in "remote-project"
    And I have a rule file "remote.md" with content "# Remote Rules" in "remote-project"
    When I run "rfh pack remote-project --verbose"
    Then I should see "📁 Working directory: remote-project"
    And I should see "📦 Packing remote-rules v1.5.0"
    And I should see "✅ Successfully packed remote-rules"

  Scenario: Pack command help text
    When I run "rfh pack --help"
    Then I should see "Creates a tar.gz archive containing all files"
    And I should see "--manifest string   path to manifest file"
    And I should see "--file string       override single file to pack"
    And I should see "--output string     output archive path"

  Scenario: Pack with missing manifest
    Given I have a temporary project directory
    And I have a rule file "orphan-rules.md" with content "# Orphan Rules"
    When I run "rfh pack" in the project directory
    Then I should see "failed to load manifest"
    And I should see "The system cannot find the file specified"
    And the command should exit with non-zero status

  Scenario: Pack with missing file in manifest
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "broken-rules" and version "1.0.0"
    And the manifest includes file "missing-file.md"
    When I run "rfh pack" in the project directory
    Then I should see "failed to pack files"
    And I should see "no files matched the specified patterns"
    And the command should exit with non-zero status

  Scenario: Pack verbose output
    Given I have a temporary project directory
    And I have a rulestack.json manifest with name "verbose-test" and version "1.0.0"
    And I have a rule file "verbose-rules.md" with content "# Verbose Rules"
    When I run "rfh pack --verbose" in the project directory
    Then I should see "📄 Manifest file:"
    And I should see "📦 Packing verbose-test v1.0.0"
    And I should see "📋 Files included:"
    And I should see "✅ Successfully packed verbose-test"