Feature: Git Client Basic Operations
  As a developer
  I want to use Git-based registries 
  So that I can access packages stored in Git repositories

  Background:
    Given RFH is installed and accessible
    And I have a clean config file

  Scenario: Add Git registry with Gitea URL
    When I run "rfh registry add test-public http://localhost:3000/rfh-admin/rfh-test-registry-public.git --type git"
    Then I should see "Added registry 'test-public'"
    And I should see "Type: git"
    And I should see "Set git_token in config or use GITHUB_TOKEN environment variable"
    And the config should contain registry "test-public" with type "git"

  Scenario: Add Git registry with .git suffix
    When I run "rfh registry add test-dotgit http://localhost:3000/rfh-admin/rfh-test-registry-public.git --type git"
    Then I should see "Added registry 'test-dotgit'"
    And I should see "Type: git"
    And the config should contain registry "test-dotgit" with URL ending in ".git"

  Scenario: Git registry health check fails without authentication for private repo
    Given I have a Git registry "private-repo" configured at "http://localhost:3000/rfh-admin/rfh-test-registry-private.git"
    And the Git token is not configured
    When I run "rfh registry health-check private-repo"
    Then the command should exit with non-zero status
    And I should see an error about authentication being required
    And I should see "provide a Git token for private repositories"

  Scenario: Git registry health check with public repository
    Given I have a Git registry "public-repo" configured at "http://localhost:3000/rfh-admin/rfh-test-registry-public.git"
    And the repository contains valid package structure
    When I run "rfh registry health-check public-repo"
    Then the command should exit with zero status
    And I should see "Git registry is healthy"

  Scenario: Git registry health check detects invalid structure
    Given I have a Git registry "invalid-repo" configured at "http://localhost:3000/rfh-admin/rfh-test-invalid-registry.git"
    And the repository does not contain packages directory or index.json
    When I run "rfh registry health-check invalid-repo"
    Then the command should exit with non-zero status
    And I should see an error about invalid registry structure
    And I should see "neither packages directory nor index.json found"

  Scenario: Git client caches repository locally
    Given I have a Git registry "cached-repo" configured at "http://localhost:3000/rfh-admin/rfh-test-registry-public.git"
    When I run "rfh registry health-check cached-repo --verbose"
    Then I should see "Cloning repository" in verbose output
    And I should see "Cache directory:" in verbose output
    And a cached repository should exist in the user's .rfh directory

  Scenario: Git client uses cached repository on subsequent operations
    Given I have a Git registry "cached-repo" configured at "http://localhost:3000/rfh-admin/rfh-test-registry-public.git"
    And the repository is already cached locally
    When I run "rfh registry health-check cached-repo --verbose"
    Then I should see "Opening cached repository" in verbose output
    And I should see "Pulling latest changes" in verbose output
    And I should not see "Cloning repository" in git output

  Scenario: Git client supports multiple Git providers
    Given I have a GitLab registry "gitlab-repo" configured at "https://gitlab.com/org/registry"
    And I have a Bitbucket registry "bitbucket-repo" configured at "https://bitbucket.org/org/registry" 
    When I check authentication methods for both registries
    Then GitLab should use "oauth2" as username
    And Bitbucket should use "x-token-auth" as username
    And GitHub should use "token" as username

  Scenario: Git registry URL normalization
    When I add a Git registry with URL "http://localhost:3000/rfh-admin/rfh-test-registry-public"
    Then the stored URL should be "http://localhost:3000/rfh-admin/rfh-test-registry-public.git"
    When I add a Git registry with URL "http://localhost:3000/rfh-admin/rfh-test-registry-public/"
    Then the stored URL should be "http://localhost:3000/rfh-admin/rfh-test-registry-public.git"

  Scenario: Git client placeholder methods return not implemented errors
    Given I have a Git registry "test-repo" configured
    When I try to search packages in the Git registry
    Then I should see an error "not yet implemented - see Phase 5" in git operation
    When I try to get a package from the Git registry
    Then I should see an error "not yet implemented - see Phase 5" in git operation
    When I try to publish to the Git registry
    Then I should see an error "not yet implemented - see Phase 6" in git operation