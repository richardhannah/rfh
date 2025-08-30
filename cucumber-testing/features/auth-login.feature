Feature: User Login
  As a developer
  I want to login to my user account
  So that I can authenticate and access protected functionality

  Background:
    Given RFH is installed and accessible
    And I have a clean config file

  Scenario: Login command availability
    Given I have a registry "test-registry" configured at "http://localhost:8081"
    And "test-registry" is the active registry  
    When I attempt to login interactively
    Then I should see "Logging in to http://localhost:8081"
    And I should see "Username:"

  Scenario: Login with no active registry
    Given I have a clean config file with no registries
    When I login with username "testuser" and password "password123"
    Then I should see "no active registry configured"
    And I should see "Use 'rfh registry add' to add one"

  Scenario: Login when active registry not found
    Given I have a config with current registry "missing-registry"
    When I login with username "testuser" and password "password123"
    Then I should see "active registry 'missing-registry' not found"

  Scenario: Auth login help text
    When I run "rfh auth login --help"
    Then I should see "Login to your user account with username and password"
    And I should see "Your JWT token will be saved locally for future API calls"

  Scenario: Login command validation
    When I run "rfh auth --help"
    Then I should see "login       Login to your user account"

  Scenario: Login with registry connection failure
    Given I have a registry "offline-registry" configured at "http://localhost:9999"
    And "offline-registry" is the active registry
    When I login with username "testuser" and password "password123"
    Then I should see "Logging in to http://localhost:9999"

  Scenario: Successful non-interactive login with valid credentials
    Given I have a registry "test-registry" configured at "http://localhost:8081"
    And "test-registry" is the active registry
    When I login with username "validuser" and password "validpass"
    Then I should see "Using provided credentials for validuser"
    And I should see "Logging in to http://localhost:8081"

  Scenario: Non-interactive login with invalid credentials  
    Given I have a registry "test-registry" configured at "http://localhost:8081"
    And "test-registry" is the active registry
    When I login with username "invaliduser" and password "wrongpass"
    Then I should see "Using provided credentials for invaliduser"
    And I should see "login failed"