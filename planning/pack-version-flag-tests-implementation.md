# Pack Version Flag Test Coverage Implementation Plan

## Feature Description
Add comprehensive Cucumber test coverage for the `--version` flag in the `rfh pack` command.

## Current Status
- The `--version` flag exists and is functional in the code
- Zero test coverage for this functionality
- Risk of regressions without verification

## Implementation Approach
Add test scenarios to `cucumber-testing/features/04-pack.feature` to cover all `--version` flag use cases.

## Test Scenarios to Add

### 1. Help Text Verification
- Update existing help text scenario to verify `--version` flag appears
- Should show: `--version string   package version (default: 1.0.0)`

### 2. Custom Version Functionality  
- Test: `rfh pack --file=test.mdc --package=test-pkg --version=1.2.3`
- Verify correct archive filename: `test-pkg-1.2.3.tgz`
- Verify correct package directory: `.rulestack/test-pkg.1.2.3/`

### 3. Default Version Behavior
- Test pack without `--version` flag uses default "1.0.0"
- Confirm archive naming follows default pattern

### 4. Integration Testing
- Verify `rfh status` correctly displays archives with custom versions
- Test that custom versions work with publish workflow

## Files to Modify
- `cucumber-testing/features/04-pack.feature` - Add new test scenarios

## Expected Test Structure
```gherkin
Scenario: Pack with custom version number
  Given I have a rule file "version-test.mdc" with content "# Version Test"
  When I run "rfh pack --file=version-test.mdc --package=version-test --version=2.1.5" in the project directory
  Then I should see "✅ Created new package: version-test v2.1.5"
  And the archive file ".rulestack/staged/version-test-2.1.5.tgz" should exist
  And the directory ".rulestack/version-test.2.1.5" should exist
```

## Benefits
- Ensures `--version` flag works correctly
- Prevents regressions in version handling
- Validates integration with status and publish commands
- Documents expected behavior for developers

## Implementation Status
Ready to implement comprehensive test coverage.