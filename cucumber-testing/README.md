# RFH Cucumber Testing

This directory contains Behavior-Driven Development (BDD) test scenarios written in Cucumber/Gherkin format for the RFH (RuleStack File Handler) CLI tool.

## Structure

- `usage-scenarios.md` - Master document with all feature scenarios
- `*.feature` - Individual feature files for specific functionality

## RFH Init Test Files

### `init-new-project.feature`
Tests the basic functionality of initializing a new RuleStack project:
- Default project initialization
- Manifest structure validation
- Directory structure creation
- Success message format verification

### `init-existing-project.feature` 
Tests behavior when initializing in directories with existing projects:
- Warning messages for existing projects
- Force flag override behavior
- Partial project file handling
- User confirmation prompts

### `init-custom-name.feature`
Tests custom project name functionality:
- Custom name parameter handling
- Name validation rules
- Invalid character rejection
- Directory name auto-suggestion
- Name sanitization

### `init-scope-removal.feature`
Tests the scope removal initiative implementation:
- Default simple name generation
- Validation against legacy scoped names
- Consistent naming across all files
- Migration hints for existing scoped projects

## Running Tests

These Cucumber scenarios are designed to be:
1. **Documentation** - Clear specification of expected behavior
2. **Test Cases** - Can be implemented with step definitions for automated testing
3. **Validation** - Manual verification checklist for functionality

## Test Implementation

To implement automated testing:
1. Choose a BDD framework (e.g., Cucumber for Go, Godog)
2. Create step definitions for each Given/When/Then statement
3. Set up test environments and fixtures
4. Implement assertions for expected outcomes

## Status

- ✅ **rfh init** - All scenarios defined and documented
- 🚧 **rfh pack** - Scenarios pending based on testing results
- 🚧 **rfh publish** - Scenarios pending based on testing results
- ✅ **rfh registry** - Scenarios defined in usage-scenarios.md
- ✅ **rfh auth** - Scenarios defined in usage-scenarios.md