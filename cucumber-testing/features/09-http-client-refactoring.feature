Feature: HTTP Client Refactoring
  HTTP registry operations should continue working after refactoring to new interface
  All existing functionality must work with the new RegistryClient interface

  Background:
    Given RFH is installed and accessible
    And I have a clean config file
    And I have a running HTTP registry at "http://localhost:8080"

  Scenario: Search works with refactored HTTP client
    Given I have a registry "local" configured at "http://localhost:8080" with type "remote-http"
    And "local" is the active registry
    When I run "rfh search test"
    Then the command should succeed
    And the HTTP registry should receive a valid search request

  Scenario: Get package works with refactored HTTP client
    Given I have a registry "local" configured at "http://localhost:8080" with type "remote-http"
    And "local" is the active registry
    And the registry has a package "test-rules" available
    When I run "rfh get test-rules"
    Then the command should succeed
    And the HTTP registry should receive a valid get package request

  Scenario: Publish works with refactored HTTP client
    Given I have a registry "local" configured at "http://localhost:8080" with type "remote-http"
    And "local" is the active registry
    And I am authenticated to the registry
    And I have a rule file "publish-test.mdc" with content "# Publish Test Rules"
    When I run "rfh pack --file=publish-test.mdc --package=refactor-test" in the project directory
    And I run "rfh publish" in the project directory
    Then the command should succeed
    And the HTTP registry should receive a valid publish request with proper authentication
    And I should see the package version and SHA256 in the output

  Scenario: Context timeout is respected in HTTP operations
    Given I have a registry "slow" configured at "http://localhost:8080" with type "remote-http"
    And "slow" is the active registry
    And the registry responds with 5 second delays
    When I run "rfh search test" with a 1 second timeout
    Then the command should fail with a timeout error
    And I should see "context deadline exceeded" or "timeout" in the output

  Scenario: Error types are properly converted from HTTP responses
    Given I have a registry "local" configured at "http://localhost:8080" with type "remote-http"
    And "local" is the active registry
    When I run "rfh get nonexistent-package"
    Then the command should fail
    And I should see "package not found" or similar error message
    And the command should exit with non-zero status

  Scenario: Package struct fields are properly displayed in search results
    Given I have a registry "local" configured at "http://localhost:8080" with type "remote-http"  
    And "local" is the active registry
    And the registry has packages with various metadata
    When I run "rfh search example"
    Then I should see package names and latest versions
    And I should see package descriptions where available
    And I should see tags when packages have them
    And the output format should be consistent with old behavior

  Scenario: Legacy client wrapper maintains compatibility
    Given I have a registry "local" configured at "http://localhost:8080" with type "remote-http"
    And "local" is the active registry
    And I have existing code that uses the old Client interface
    When the legacy wrapper is used for search operations
    Then it should return results in the old map-based format
    And all existing functionality should work unchanged

  Scenario: Authentication headers are properly set with new client
    Given I have a registry "auth-test" configured at "http://localhost:8080" with type "remote-http"
    And "auth-test" is the active registry
    And I have a JWT token configured for the registry
    When I run "rfh search authenticated"
    Then the HTTP registry should receive the request with proper Authorization header
    And the Bearer token should match the configured JWT token

  Scenario: Health check works with context support
    Given I have a registry "health" configured at "http://localhost:8080" with type "remote-http"
    And "health" is the active registry
    When the health check is performed
    Then it should use the new context-aware Health method
    And it should succeed when the registry is healthy
    And it should fail appropriately when the registry is down

  Scenario: Download operations use context properly
    Given I have a registry "download" configured at "http://localhost:8080" with type "remote-http"
    And "download" is the active registry
    And I have a package "download-test@1.0.0" available in the registry
    When I run "rfh add download-test@1.0.0"
    Then the download should use the new DownloadBlob method with context
    And it should respect timeout settings
    And the package should be installed successfully

  Scenario: Multiple HTTP registries work with refactored client
    Given I have a registry "prod" configured at "http://prod.example.com" with type "remote-http"
    And I have a registry "staging" configured at "http://staging.example.com" with type "remote-http"
    When I run "rfh registry use prod"
    Then "prod" should be the active registry
    When I run "rfh search test"
    Then the request should go to the prod registry
    When I run "rfh registry use staging"
    And I run "rfh search test"
    Then the request should go to the staging registry

  Scenario: HTTP client factory creates correct client type
    Given I have a registry "factory-test" configured at "http://localhost:8080" with type "remote-http"
    When the registry client is created via the factory
    Then it should return an HTTPClient instance
    And the client should implement the RegistryClient interface
    And the client Type() method should return "remote-http"

  Scenario: Verbose mode works with refactored HTTP client
    Given I have a registry "verbose" configured at "http://localhost:8080" with type "remote-http"
    And "verbose" is the active registry
    When I run "rfh search test --verbose"
    Then I should see HTTP request details in verbose output
    And I should see response status codes
    And I should see authentication header information (redacted)
    And the verbose output should maintain the same format as before