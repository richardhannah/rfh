Feature: Registry Management
  As a developer
  I want to manage multiple registry configurations
  So that I can work with different package repositories

  Background:
    Given RFH is installed and accessible
    And I have a clean config file

  Scenario: Add first registry
    When I run "rfh registry add production https://registry.example.com"
    Then I should see "Added registry 'production'"
    And I should see "Set as active registry"
    And the config should contain registry "production" with URL "https://registry.example.com"
    And "production" should be the current active registry

  Scenario: Add additional registry
    Given I have a registry "production" configured at "https://registry.example.com"
    When I run "rfh registry add staging https://staging.example.com"
    Then I should see "Added registry 'staging'"
    And the config should contain both registries
    And "production" should remain the active registry

  Scenario: Add registry with token
    When I run "rfh registry add private https://private.example.com my-secret-token"
    Then I should see "Added registry 'private'"
    And the config should contain registry "private" with URL "https://private.example.com"
    And the config should store the token for "private"

  Scenario: List registries when none configured
    When I run "rfh registry list"
    Then I should see "No registries configured"

  Scenario: List single registry
    Given I have a registry "production" configured at "https://registry.example.com"
    When I run "rfh registry list"
    Then I should see "production" in the registry list
    And I should see "https://registry.example.com" in the registry list
    And I should see "production" marked as active

  Scenario: List multiple registries
    Given I have a registry "production" configured at "https://registry.example.com"
    And I have a registry "staging" configured at "https://staging.example.com"
    And "staging" is the active registry
    When I run "rfh registry list"
    Then I should see both registries in the list
    And I should see "staging" marked as active
    And I should not see "production" marked as active

  Scenario: Switch active registry
    Given I have a registry "production" configured at "https://registry.example.com"
    And I have a registry "staging" configured at "https://staging.example.com"
    And "production" is the active registry
    When I run "rfh registry use staging"
    Then I should see "Set 'staging' as active registry"
    And "staging" should be the current active registry

  Scenario: Remove non-active registry
    Given I have a registry "production" configured at "https://registry.example.com"
    And I have a registry "staging" configured at "https://staging.example.com"
    And "staging" is the active registry
    When I run "rfh registry remove production"
    Then I should see "Removed registry 'production'"
    And "staging" should remain the active registry
    And the config should not contain registry "production"

  Scenario: Remove active registry
    Given I have a registry "production" configured at "https://registry.example.com"
    And I have a registry "staging" configured at "https://staging.example.com"
    And "staging" is the active registry
    When I run "rfh registry remove staging"
    Then I should see "Removed active registry"
    And I should see a warning about setting a new active registry
    And no registry should be active

  Scenario: Handle invalid registry operations
    When I run "rfh registry use nonexistent"
    Then I should see an error about registry not found
    And the command should exit with non-zero status

  Scenario: Handle duplicate registry name
    Given I have a registry "production" configured at "https://registry.example.com"
    When I run "rfh registry add production https://different.example.com"
    Then I should see "Added registry 'production'"
    And the config should contain registry "production" with URL "https://different.example.com"