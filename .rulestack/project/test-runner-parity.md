# Test Runner Script Parity Rule

## Rule: Maintain Complete Functional Parity Between Test Runners

### Overview
The RFH project maintains two test runner scripts to support cross-platform testing:
- `run-tests.ps1` (PowerShell for Windows)
- `run-tests.sh` (Bash for Unix/Linux/Mac)

**CRITICAL**: These scripts MUST maintain complete functional parity. Any change to one script requires updating the other to match.

### Requirements

#### When Editing Either Script
1. **Always update both scripts** - Never modify just one test runner
2. **Match functionality exactly** - Both scripts must provide identical features
3. **Test on both platforms** if possible, or document changes clearly for testing

#### Key Features That Must Remain Synchronized

##### Infrastructure Management
- Docker availability checking
- API server building and deployment via Docker Compose
- Health check waiting logic for API readiness
- Container cleanup messages and instructions

##### Build Steps
- CLI binary building (`dist/rfh.exe` on Windows, `dist/rfh` on Unix)
- API server building (when Docker is available)
- Error handling for build failures

##### Test Execution
- Dependency installation checks (node_modules)
- Test target selection (all, init, actual, working, etc.)
- Cucumber test execution commands
- Exit code preservation and propagation

##### User Feedback
- Consistent status messages and progress indicators
- Warning messages for missing dependencies
- Docker status notifications
- Test completion summaries
- Debugging hints and tips

### Implementation Guidelines

#### Adding New Features
When adding a new feature to either script:

1. **Design the feature** considering both PowerShell and Bash syntax
2. **Implement in one script** with clear comments
3. **Port to the other script** maintaining identical behavior
4. **Test both scripts** to ensure parity
5. **Document any platform-specific differences** if unavoidable

#### Common Patterns

##### PowerShell to Bash Equivalents
```powershell
# PowerShell
if ($LASTEXITCODE -ne 0) { exit 1 }

# Bash equivalent
if [ $? -ne 0 ]; then exit 1; fi
```

```powershell
# PowerShell
Write-Host "Message" -ForegroundColor Green

# Bash equivalent
echo -e "\033[32mMessage\033[0m"  # or use emoji indicators
```

```powershell
# PowerShell
docker info *> $null

# Bash equivalent
docker info &> /dev/null
```

##### Handling Docker Commands
Both scripts should:
- Check Docker availability the same way
- Use identical docker-compose commands
- Implement the same retry/timeout logic
- Provide equivalent error messages

### Testing Checklist

When modifying test runners, verify:

- [ ] Both scripts start Docker containers (if Docker available)
- [ ] Both scripts build the CLI binary
- [ ] Both scripts handle missing dependencies identically
- [ ] Both scripts support the same command-line arguments
- [ングBoth scripts provide equivalent user feedback
- [ ] Both scripts return the same exit codes
- [ ] Both scripts handle errors consistently

### Examples of Required Parity

#### Docker Setup (Current Implementation)
Both scripts now:
1. Check if Docker is available
2. Run `docker-compose down` to clean up
3. Run `docker-compose up --build -d` to start services
4. Wait up to 30 seconds for API health
5. Warn if API is not responding

#### Test Target Selection
Both scripts support:
- `all` or no argument - runs all tests
- `init` - runs init-related tests only
- `actual` - runs actual behavior tests only
- `working` - runs only passing scenarios

### Maintenance Notes

#### Version History
- Initial scripts created with basic test running
- Added Docker support to bash script
- Added Docker support to PowerShell script (maintaining parity)
- Both scripts now have complete feature parity as of the latest update

#### Known Differences (Acceptable)
- File extension: `.ps1` vs `.sh`
- Executable name: `rfh.exe` vs `rfh` (platform-specific)
- Shell-specific syntax while maintaining identical functionality
- Color output methods (PowerShell -ForegroundColor vs bash escape codes/emoji)

### Enforcement

This rule should be enforced during code review. Any PR that modifies either test runner script should:

1. Update both scripts
2. Document the changes in the PR description
3. Confirm testing on available platforms
4. Note any platform-specific considerations

### Benefits of Maintaining Parity

1. **Consistent Developer Experience** - Developers on any platform get the same features
2. **Reliable CI/CD** - Tests behave identically in different environments
3. **Easier Maintenance** - Features and fixes need to be designed once
4. **Better Documentation** - One set of behaviors to document
5. **Reduced Bugs** - Prevents platform-specific issues from being overlooked

---

**Remember**: The test runners are critical infrastructure. They must work reliably and consistently across all platforms to ensure the quality of the RFH project.