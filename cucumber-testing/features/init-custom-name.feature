Feature: Initialize with custom project name
  As a developer
  I want to initialize a project with a custom name
  So that I can create packages with meaningful names for my use case

  Background:
    Given I am in an empty directory
    And RFH is installed and accessible

  Scenario: Initialize with simple custom name
    When I run "rfh init --name my-custom-rules"
    Then I should see "Initialized RuleStack project"
    And the manifest should have name "my-custom-rules"
    And the name should not contain scope characters "@" or "/"
    And all other fields should use default values

  Scenario: Initialize with kebab-case name
    When I run "rfh init --name validation-rules"
    Then the manifest should have name "validation-rules"
    And the project should initialize successfully

  Scenario: Initialize with snake_case name
    When I run "rfh init --name my_validation_rules"
    Then the manifest should have name "my_validation_rules"
    And the project should initialize successfully

  Scenario: Reject scoped package names
    When I run "rfh init --name @myorg/my-rules"
    Then I should see an error "Package names with scopes are not supported"
    And I should see "Please use a simple name without '@' or '/'"
    And no "rulestack.json" should be created
    And the command should exit with non-zero status

  Scenario: Reject names with forward slashes
    When I run "rfh init --name myorg/my-rules"
    Then I should see an error "Package names cannot contain '/'"
    And I should see "Please use a simple name"
    And no project files should be created

  Scenario: Reject names with invalid characters
    When I run "rfh init --name 'my rules with spaces'"
    Then I should see an error "Package name contains invalid characters"
    And I should see "Use only letters, numbers, hyphens, and underscores"
    And no project files should be created

  Scenario: Validate name length limits
    When I run "rfh init --name a"
    Then I should see an error "Package name too short (minimum 2 characters)"
    
    When I run "rfh init --name" followed by a 100-character string
    Then I should see an error "Package name too long (maximum 50 characters)"

  Scenario: Initialize with name matching directory
    Given I am in a directory named "awesome-ai-rules"
    When I run "rfh init --name awesome-ai-rules"
    Then the manifest should have name "awesome-ai-rules"
    And the project should initialize successfully

  Scenario: Auto-suggest name based on directory
    Given I am in a directory named "my-awesome-rules"
    When I run "rfh init"
    Then I should see "Detected directory name: my-awesome-rules"
    And I should see "Use directory name? (Y/n)"
    When I respond "y"
    Then the manifest should have name "my-awesome-rules"

  Scenario: Sanitize directory name for auto-suggestion
    Given I am in a directory named "My Rules with Spaces"
    When I run "rfh init"
    Then I should see "Suggested name: my-rules-with-spaces"
    And I should see "Use suggested name? (Y/n)"
    When I respond "y"
    Then the manifest should have name "my-rules-with-spaces"