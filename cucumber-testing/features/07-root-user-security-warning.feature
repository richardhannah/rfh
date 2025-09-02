Feature: Root User Security Warning
  As a security-conscious system
  I want to warn users when they are logged in as root
  So that they are encouraged to use regular admin accounts for daily operations

  Background:
    Given I am in an empty directory
    And RFH is installed and accessible

  Scenario: Warning appears when logged in as root user
    Given I am logged in as "root" user
    When I run "rfh init"
    Then I should see "🚨 SECURITY WARNING 🚨"
    And I should see "YOU ARE LOGGED IN AS ROOT USER!"
    And I should see "Create a regular user account: rfh auth register"
    And I should see "Grant admin privileges to your user account"
    And I should see "Disable or change the root account password"
    And I should see "Use your regular account for daily operations"

  Scenario: Warning appears for pack command when logged in as root
    Given I am logged in as "root" user
    And I have a rule file "test.mdc" with content "# Test Rule"
    When I run "rfh pack --file=test.mdc --package=test-package"
    Then I should see "🚨 SECURITY WARNING 🚨"
    And I should see "YOU ARE LOGGED IN AS ROOT USER!"

  Scenario: Warning appears for status command when logged in as root
    Given I am logged in as "root" user
    When I run "rfh status"
    Then I should see "🚨 SECURITY WARNING 🚨"
    And I should see "YOU ARE LOGGED IN AS ROOT USER!"

  Scenario: No warning appears for regular users
    Given I am logged in as "regular-user" user
    When I run "rfh init"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"
    But I should see "Initializing RuleStack project"

  Scenario: No warning appears for admin users
    Given I am logged in as "admin" user
    When I run "rfh status"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"

  Scenario: Case insensitive root user detection
    Given I am logged in as "ROOT" user
    When I run "rfh status"
    Then I should see "🚨 SECURITY WARNING 🚨"
    And I should see "YOU ARE LOGGED IN AS ROOT USER!"

  Scenario: No warning for auth login command
    Given I am logged in as "root" user
    When I run "rfh auth login --help"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"
    But I should see "Login to your user account"

  Scenario: No warning for auth register command
    Given I am logged in as "root" user
    When I run "rfh auth register --help"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"
    But I should see "Register a new user account"

  Scenario: No warning for auth whoami command
    Given I am logged in as "root" user
    When I run "rfh auth whoami"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"

  Scenario: No warning for auth logout command
    Given I am logged in as "root" user
    When I run "rfh auth logout"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"

  Scenario: No warning when not authenticated
    Given I am not logged in to any registry
    When I run "rfh init"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"
    But I should see "Initializing RuleStack project"

  Scenario: No warning when no active registry is configured
    Given I have no active registry configured
    When I run "rfh init"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"
    But I should see "Initializing RuleStack project"

  Scenario: Warning appears for mixed case root variants
    Given I am logged in as "Root" user
    When I run "rfh status"
    Then I should see "🚨 SECURITY WARNING 🚨"
    And I should see "YOU ARE LOGGED IN AS ROOT USER!"