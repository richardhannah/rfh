Feature: Git Client Basic Operations
  As a developer
  I want to use Git-based registries 
  So that I can access packages stored in Git repositories

  Background:
    Given RFH is installed and accessible
    And I have a clean config file

  Scenario: Add Git registry
    When I run "rfh registry add test-git http://localhost:3000/rfh-admin/rfh-test-registry-public.git --type git"
    Then I should see "Added registry 'test-git'"
    And I should see "Type: git"
    And I should see "Set git_token in config or use GITHUB_TOKEN environment variable"
    And the config should contain registry "test-git" with type "git"
    And the config should contain registry "test-git" with URL ending in ".git"


  Scenario: Git client supports multiple Git providers
    Given I have a GitLab registry "gitlab-repo" configured at "https://gitlab.com/org/registry"
    And I have a Bitbucket registry "bitbucket-repo" configured at "https://bitbucket.org/org/registry" 
    When I check authentication methods for both registries
    Then GitLab should use "oauth2" as username
    And Bitbucket should use "x-token-auth" as username
    And GitHub should use "token" as username

  Scenario: Git registry URL normalization
    When I add a Git registry with URL "http://localhost:3000/rfh-admin/rfh-test-registry-public"
    Then the stored URL should be "http://localhost:3000/rfh-admin/rfh-test-registry-public"
    When I add a Git registry with URL "http://localhost:3000/rfh-admin/rfh-test-registry-public/"
    Then the stored URL should be "http://localhost:3000/rfh-admin/rfh-test-registry-public/"

