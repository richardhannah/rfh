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

  Scenario: Warning appears for status command when logged in as root
    Given I am logged in as "root" user
    When I run "rfh status"
    Then I should see "🚨 SECURITY WARNING 🚨"
    And I should see "YOU ARE LOGGED IN AS ROOT USER!"

  Scenario: No warning appears for status command when logged in as admin
    Given I am logged in as "admin" user
    When I run "rfh status"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"

  Scenario: No warning appears for status command when logged in as user
    Given I am logged in as "regular-user" user
    When I run "rfh init"
    Then I should not see "🚨 SECURITY WARNING 🚨"
    And I should not see "YOU ARE LOGGED IN AS ROOT USER!"
    But I should see "Initializing RuleStack project"

