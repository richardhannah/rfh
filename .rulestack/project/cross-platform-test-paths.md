# Cross-Platform Path Assertions in Cucumber Tests

## Rule: Handle OS-Specific Path Separators in Test Assertions

### Overview
Cucumber tests must work correctly on both Windows and Unix-like systems (Linux/Mac). File paths use different separators on different operating systems, and test assertions must account for these differences.

**CRITICAL**: Never hard-code path assertions that will fail on different operating systems due to path separator differences.

### Path Separator Differences

| Operating System | Path Separator | Example Path |
|-----------------|---------------|--------------|
| Windows | Backslash `\` | `.rulestack\staged\package-1.0.0.tgz` |
| Linux/Mac | Forward slash `/` | `.rulestack/staged/package-1.0.0.tgz` |

### Implementation Requirements

#### 1. Flexible Path Assertions
When writing step definitions that check for paths in output:

```javascript
// BAD - Will fail on Windows
Then('I should see {string}', function (expectedText) {
  expect(output).to.include('.rulestack/staged/package.tgz');
});

// GOOD - Works on all platforms
Then('I should see {string}', function (expectedText) {
  let found = output.includes(expectedText);
  
  // If not found and text contains paths, try OS-specific version
  if (!found && expectedText.includes('/')) {
    const windowsPath = expectedText.replace(/\//g, '\\');
    found = output.includes(windowsPath);
  }
  
  if (!found) {
    throw new Error(`Expected text not found: ${expectedText}`);
  }
});
```

#### 2. Writing Feature Files
In `.feature` files, use forward slashes as the canonical format:

```gherkin
# GOOD - Use forward slashes in feature files
Then I should see "📁 Package directory: .rulestack/security-rules.1.0.0"
And I should see "📦 Archive: .rulestack/staged/security-rules-1.0.0.tgz"
```

The step definitions should handle converting these to the appropriate OS format.

#### 3. File System Operations
When checking for actual file existence:

```javascript
// Use path.join() for cross-platform paths
const archivePath = path.join('.rulestack', 'staged', 'package-1.0.0.tgz');

// Or normalize paths when comparing
const normalizedExpected = path.normalize(expectedPath);
const normalizedActual = path.normalize(actualPath);
```

### Common Patterns

#### Pattern 1: Output Assertion with Path
```javascript
// Step definition that handles both separators
Then('I should see {string}', function (expectedText) {
  const output = this.lastCommandOutput + this.lastCommandError;
  
  // Check for exact match first
  let found = output.includes(expectedText);
  
  // If not found and contains forward slashes, try Windows format
  if (!found && expectedText.includes('/')) {
    const windowsPath = expectedText.replace(/\//g, '\\');
    found = output.includes(windowsPath);
  }
  
  if (!found) {
    // Provide detailed error with actual output
    throw new Error(`
Expected text not found.
EXPECTED: "${expectedText}"
ACTUAL OUTPUT:
${output}
`);
  }
});
```

#### Pattern 2: File Existence Check
```javascript
// Check if file exists at path (cross-platform)
Then('the file {string} should exist', async function (filePath) {
  // Convert forward slashes to OS-specific separator
  const normalizedPath = path.normalize(filePath);
  const fullPath = path.join(this.tempProjectDir, normalizedPath);
  
  const exists = await fs.pathExists(fullPath);
  expect(exists).to.be.true;
});
```

#### Pattern 3: Path in JSON Validation
```javascript
// When validating paths in JSON files
Then('the manifest should reference {string}', async function (expectedPath) {
  const manifest = await fs.readJSON('rulestack.json');
  
  // Normalize both paths for comparison
  const normalizedExpected = path.normalize(expectedPath);
  const normalizedActual = path.normalize(manifest.files[0]);
  
  expect(normalizedActual).to.equal(normalizedExpected);
});
```

### OS-Specific Error Messages

Different operating systems also produce different error messages:

| Error Type | Windows | Linux/Mac |
|-----------|---------|-----------|
| File not found | `The system cannot find the file specified` | `no such file or directory` |
| Path not found | `The system cannot find the path specified` | `cannot access: No such file or directory` |

Handle these with flexible assertions:

```javascript
Then('I should see a file not found error', function () {
  const output = this.lastCommandOutput + this.lastCommandError;
  
  const hasFileNotFoundError = 
    output.includes('no such file or directory') || 
    output.includes('The system cannot find the file specified') ||
    output.includes('cannot find the path specified') ||
    output.includes('No such file or directory');
    
  expect(hasFileNotFoundError).to.be.true;
});
```

### Testing Guidelines

#### When Adding New Tests
1. **Always use forward slashes** in feature files
2. **Test on both platforms** if possible
3. **Use path.join()** for constructing paths in step definitions
4. **Avoid regex patterns** that assume specific separators

#### When Debugging Failed Tests
1. **Check the actual output** to see which separator is used
2. **Verify the step definition** handles both separators
3. **Ensure error messages** are checked in an OS-agnostic way

### Examples from Current Tests

#### Good Example - Flexible Path Checking
```javascript
// From init_steps.js - handles both separators
Then('I should see {string}', function (expectedText) {
  const output = this.lastCommandOutput + this.lastCommandError;
  let found = output.includes(expectedText);
  
  if (!found && expectedText.includes('/')) {
    const windowsPath = expectedText.replace(/\//g, '\\');
    found = output.includes(windowsPath);
  }
  
  if (!found) {
    throw new Error(`Expected text not found: ${expectedText}`);
  }
});
```

#### Good Example - OS-Agnostic Error Handling
```javascript
// From pack_steps.js - handles different error messages
Then('I should see a file not found error', function () {
  const output = this.lastCommandOutput + this.lastCommandError;
  const hasError = 
    output.includes('no such file or directory') || 
    output.includes('The system cannot find the file specified');
  expect(hasError).to.be.true;
});
```

### Enforcement

During code review, verify that:
1. New test assertions handle both path separators
2. Error message checks account for OS differences
3. Path construction uses appropriate Node.js path utilities
4. Feature files use consistent forward slash convention

### Benefits

1. **Cross-platform reliability** - Tests pass on all developer machines
2. **CI/CD compatibility** - Tests work regardless of CI platform
3. **Reduced debugging time** - Fewer false failures due to path issues
4. **Better contributor experience** - New contributors aren't blocked by OS-specific failures

---

**Remember**: A test suite that only works on one operating system is only half as valuable. Always write tests that work everywhere!