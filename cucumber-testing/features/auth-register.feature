Feature: User Registration Command
  As a developer
  I want to register a new user account
  So that I can authenticate and publish packages

  Background:
    Given RFH is installed and accessible
    And I have a clean config file

  Scenario: Registration command availability
    Given the test registry is configured
    When I attempt to register interactively
    Then I should see "Registering new account at http://localhost:8081"
    And I should see "Username:"

  Scenario: Registration with no active registry
    Given I have a clean config file with no registries
    When I register with username "testuser", email "test@example.com", and password "password123"
    Then I should see "no active registry configured"
    And I should see "Use 'rfh registry add' to add one"

  Scenario: Registration when active registry not found
    Given I have a config with current registry "missing-registry"  
    When I register with username "testuser", email "test@example.com", and password "password123"
    Then I should see "active registry 'missing-registry' not found"

  Scenario: Auth register help text
    When I run "rfh auth register --help"
    Then I should see "Register a new user account with username, email, and password"
    And I should see "After successful registration, you'll be automatically logged in"

  Scenario: Auth command group availability
    When I run "rfh auth --help"
    Then I should see "Authentication commands"
    And I should see "register    Register a new user account"
    And I should see "login       Login to your user account" 
    And I should see "logout      Logout from your user account"
    And I should see "whoami      Show current user information"

  Scenario: Successful non-interactive registration
    Given the test registry is configured
    When I register with username "newuser", email "newuser@example.com", and password "password123" using flags
    Then I should see "Using provided credentials for newuser"
    And I should see "Registering new account at http://localhost:8081"

  Scenario: Non-interactive registration with validation
    Given the test registry is configured
    When I register with username "testuser", email "test@domain.com", and password "securepass" using flags
    Then I should see "Using provided credentials for testuser (test@domain.com)"
    And I should see "Registering new account at http://localhost:8081"