@skip
Feature: Git Registry Search and Discovery
  Git registry should support package search and discovery operations

  Background:
    Given I have a clean test environment
    And I have initialized a new project

  Scenario: Search packages in Git registry
    Given I have a Git registry "test-git" configured
    And the Git repository contains test packages
    When I run "rfh search test"
    Then the command should succeed
    And the output should contain package results

  Scenario: Search with query filter
    Given I have a Git registry "test-git" configured
    And the Git repository contains test packages
    When I run "rfh search --query security"
    Then the command should succeed
    And the output should only contain packages matching "security"

  Scenario: Search with tag filter
    Given I have a Git registry "test-git" configured
    And the Git repository contains test packages
    When I run "rfh search --tag auth"
    Then the command should succeed
    And the output should only contain packages tagged with "auth"

  Scenario: Search with limit
    Given I have a Git registry "test-git" configured
    And the Git repository contains multiple packages
    When I run "rfh search --limit 2"
    Then the command should succeed
    And the output should contain at most 2 packages

  Scenario: Get specific package information
    Given I have a Git registry "test-git" configured
    And the Git repository contains package "test-package"
    When I run "rfh get test-package"
    Then the command should succeed
    And the output should contain package details for "test-package"
    And the output should contain version information

  Scenario: Get specific package version
    Given I have a Git registry "test-git" configured
    And the Git repository contains package "test-package" version "1.0.0"
    When I run "rfh get test-package@1.0.0"
    Then the command should succeed
    And the output should contain version details for "test-package@1.0.0"
    And the output should contain SHA256 hash
    And the output should contain dependencies

  Scenario: Get nonexistent package
    Given I have a Git registry "test-git" configured
    When I run "rfh get nonexistent-package"
    Then the command should fail
    And the error should contain "package not found"

  Scenario: Get nonexistent version
    Given I have a Git registry "test-git" configured
    And the Git repository contains package "test-package"
    When I run "rfh get test-package@99.99.99"
    Then the command should fail
    And the error should contain "version not found"

  Scenario: Download package archive
    Given I have a Git registry "test-git" configured
    And the Git repository contains package "test-package" version "1.0.0" with hash "abc123"
    When I run "rfh download abc123 ./test-archive.tar.gz"
    Then the command should succeed
    And the file "./test-archive.tar.gz" should exist
    And the file should have the correct SHA256 hash

  Scenario: Registry without index rebuilds automatically
    Given I have a Git registry "test-git-no-index" configured without index file
    When I run "rfh search test"
    Then the command should succeed
    And the output should contain "Index not found, attempting to rebuild"
    And the output should contain package results

  Scenario: Invalid registry structure
    Given I have a Git registry "test-git-invalid" configured with no packages directory
    When I run "rfh search test"
    Then the command should fail
    And the error should contain "invalid registry"

  Scenario: Search with verbose output
    Given I have a Git registry "test-git" configured
    And the Git repository contains test packages
    When I run "rfh search test --verbose"
    Then the command should succeed
    And the output should contain "Searching packages with query"
    And the output should contain "Found" followed by "packages"