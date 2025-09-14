@skip
Feature: Gitea Registry Integration
  Test Gitea-specific Git registry operations

  Background:
    Given I have a clean test environment
    And Gitea server is running at localhost:3000

  Scenario: Pack and push rules to Gitea registry
    Given I have a Git registry "gitea-test" configured at "http://localhost:3000/rfh-admin/rfh-test-registry-public.git"
    And I have initialized a new project
    And I use registry "gitea-test"
    And I create a test rule file "test-gitea-rule.mdc" with content:
      """
      # Test Gitea Rule
      This is a test rule for Gitea registry.
      
      ## Rule Content
      - Test rule for Gitea integration
      - Should work with git push workflow
      """
    When I run "rfh pack --file=test-gitea-rule.mdc --package=gitea-test-package"
    Then the command should succeed
    And the output should contain "Package created successfully"
    # Publishing to Git registries requires GitHub username
    Given I set environment variable "GITHUB_USERNAME" to "test-user"
    When I run "rfh publish"
    Then the command should fail gracefully
    # This is expected to fail since we don't have proper auth setup for Gitea
    And the output should contain one of:
      | GitHub username required |
      | authentication |
      | only GitHub repositories supported |
      | failed to publish |

  Scenario: Search packages in Gitea registry
    Given I have a Git registry "gitea-test" configured at "http://localhost:3000/rfh-admin/rfh-test-registry-public.git"
    And I use registry "gitea-test"
    When I run "rfh search test"
    Then the command should succeed or fail with "not implemented"
    # Git registry search is implemented in Phase 5
    And if successful the output should contain package results

  Scenario: Clone Gitea repository locally
    Given I have a Git registry "gitea-test" configured at "http://localhost:3000/rfh-admin/rfh-test-registry-public.git"
    When I run "rfh registry use gitea-test"
    Then the command should succeed
    When I run "rfh registry info"
    Then the command should succeed
    And the output should contain "gitea-test"
    And the output should contain "git"