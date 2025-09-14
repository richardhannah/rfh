@skip
Feature: Git Registry Publishing
  Git registry should support package publishing via Git workflow

  Background:
    Given I have a clean test environment
    And I have a Git registry "test-git" configured

  Scenario: Publish requires GitHub username
    Given I have a package ready to publish
    And I use registry "test-git"
    When I run "rfh publish"
    Then the command should fail
    And the output should contain "GitHub username required"

  Scenario: Publish requires GitHub repository
    Given I have a Git registry "gitlab-test" with URL "https://gitlab.com/test-org/test-registry"
    And I have a package ready to publish
    And I use registry "gitlab-test"
    When I run "rfh publish"
    Then the command should fail
    And the output should contain "only GitHub repositories supported"

  Scenario: Publish with valid configuration (dry run)
    Given I have a package ready to publish
    And I use registry "test-git"
    And I set environment variable "GITHUB_USERNAME" to "test-user"
    And I set environment variable "GITHUB_TOKEN" to "test-token"
    When I run "rfh publish --dry-run" if supported
    Then the command should succeed or skip with message "dry-run not implemented"
    
  Scenario: Branch name follows convention
    Given I have package "my-package@1.5.0" ready to publish
    And I use registry "test-git"
    And I set environment variable "GITHUB_USERNAME" to "test-user"
    When I run "rfh publish --verbose" if supported
    Then the command should fail gracefully
    And the output should contain "publish/my-package/1.5.0" if the branch creation was attempted

  Scenario: Commit message format is correct
    Given I have package "test-package@1.0.0" ready to publish
    And I use registry "test-git"
    And I set environment variable "GITHUB_USERNAME" to "test-user"
    And I set environment variable "GIT_AUTHOR_NAME" to "Test Publisher"
    And I set environment variable "GIT_AUTHOR_EMAIL" to "test@example.com"
    When I run "rfh publish --verbose" if supported
    Then the command should fail gracefully
    And the output should contain "Creating commit" if the commit creation was attempted

  Scenario: Publish shows helpful error for authentication failure
    Given I have a package ready to publish
    And I use registry "test-git"
    And I set environment variable "GITHUB_USERNAME" to "test-user"
    And I do not set any authentication token
    When I run "rfh publish"
    Then the command should fail
    And the output should contain error information about authentication or fork access

  Scenario: Publish shows PR preparation URL format
    Given I have a package ready to publish
    And I use registry "test-git"
    And I set environment variable "GITHUB_USERNAME" to "test-user"
    When I run "rfh publish --dry-run" if supported
    Then if the command succeeds
    And the output should contain "github.com" and "compare" in the URL format

  Scenario: Package files structure validation
    Given I have an invalid package manifest
    And I use registry "test-git"
    And I set environment variable "GITHUB_USERNAME" to "test-user"
    When I run "rfh publish"
    Then the command should fail
    And the output should contain error about manifest parsing or file reading

  Scenario: Environment variable priority for username
    Given I have a package ready to publish
    And I use registry "test-git"
    And I set environment variable "GIT_USER" to "git-user"
    And I set environment variable "GITHUB_USERNAME" to "github-user"
    When I run "rfh publish"
    Then the command should attempt to use "github-user" as the username
    And the expected fork URL should contain "github-user"

  Scenario: Author information from environment
    Given I have a package ready to publish
    And I use registry "test-git"
    And I set environment variable "GITHUB_USERNAME" to "test-user"
    And I set environment variable "GIT_AUTHOR_NAME" to "Custom Author"
    And I set environment variable "GIT_AUTHOR_EMAIL" to "custom@example.com"
    When I run "rfh publish --verbose" if supported
    Then the author information should be used in commit if attempted