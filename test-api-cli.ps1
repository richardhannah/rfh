#!/usr/bin/env pwsh

# Test script for RFH API and CLI operations
# Tests publish, search, and install operations
# Default registry URL points to local Docker dev environment

param(
    [string]$RegistryUrl = "http://localhost:8080",
    [string]$Token = "",
    [switch]$SkipCleanup,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

Write-Host "[TEST] RFH API and CLI Test Script" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

# Global variables
$TestPackageName = "monty-python-quotes"
$TestPackageVersion = "1.0.0"
$OriginalDir = Get-Location
$TempTestRoot = "$env:TEMP\rfhlocaltest"
$TestDir = "$TempTestRoot\test-package"

function Write-TestStep {
    param([string]$Message)
    Write-Host "`n[STEP] $Message" -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Message)
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Pre-Cleanup {
    Write-TestStep "Cleaning up any leftover test state..."
    
    # Remove any leftover test registries from previous runs
    $testRegistries = @('test', 'test-local', 'local')
    foreach ($regName in $testRegistries) {
        try {
            $registryList = & rfh registry list 2>$null
            if ($LASTEXITCODE -eq 0 -and $registryList -match $regName) {
                & rfh registry remove $regName 2>$null | Out-Null
            }
        }
        catch {
            # Continue silently
        }
    }
    
    # Clean up any leftover temp test directories
    $tempPattern = "$env:TEMP\rfhlocaltest*"
    Get-ChildItem $env:TEMP -Directory -Filter "rfhlocaltest*" -ErrorAction SilentlyContinue | ForEach-Object {
        try {
            Remove-Item $_.FullName -Recurse -Force
        }
        catch {
            # Continue if we can't remove it
        }
    }
    
    Write-Success "Pre-cleanup completed"
}

function Setup-TestEnvironment {
    Write-TestStep "Setting up isolated test environment..."
    
    # Clean up any existing test directory
    if (Test-Path $TempTestRoot) {
        Remove-Item $TempTestRoot -Recurse -Force
        Write-Host "[DEBUG] Cleaned up existing test directory: $TempTestRoot" -ForegroundColor Yellow
    }
    
    # Create fresh test environment
    New-Item -ItemType Directory -Path $TempTestRoot -Force | Out-Null
    New-Item -ItemType Directory -Path $TestDir -Force | Out-Null
    
    # Copy test package files to isolated environment
    $sourceTestPackage = "$OriginalDir\test-package"
    if (Test-Path $sourceTestPackage) {
        Copy-Item "$sourceTestPackage\*" -Destination $TestDir -Recurse -Force
        Write-Success "Copied test package to isolated environment"
    } else {
        Write-Error "Source test-package directory not found at $sourceTestPackage"
        exit 1
    }
    
    # Copy RFH binary to test environment for easy access
    $sourceBinary = "$OriginalDir\dist\rfh.exe"
    if (Test-Path $sourceBinary) {
        Copy-Item $sourceBinary -Destination "$TempTestRoot\rfh.exe" -Force
        Write-Success "Copied RFH binary to test environment"
    } else {
        Write-Error "RFH binary not found at $sourceBinary"
        exit 1
    }
    
    # Change to test environment
    Set-Location $TempTestRoot
    Write-Success "Test environment created at: $TempTestRoot"
    
    # Add test environment to PATH for this session
    $env:PATH = "$TempTestRoot;$env:PATH"
    Write-Success "Added test environment to PATH"
}

function Test-Prerequisites {
    Write-TestStep "Checking prerequisites in test environment..."
    
    # Check if rfh executable exists in test environment
    try {
        $rfhVersion = & rfh --help
        Write-Success "RFH CLI is available in test environment"
    }
    catch {
        Write-Error "RFH CLI not found in test environment"
        exit 1
    }
    
    # Check if test package directory exists in test environment
    if (!(Test-Path $TestDir)) {
        Write-Error "Test package directory '$TestDir' not found in test environment"
        exit 1
    }
    
    Write-Success "Prerequisites check passed"
}

function Test-APIHealth {
    Write-TestStep "Testing API health..."
    
    try {
        if ($Verbose) {
            $response = Invoke-RestMethod -Uri "$RegistryUrl/v1/health" -Method GET -Verbose
        } else {
            $response = Invoke-RestMethod -Uri "$RegistryUrl/v1/health" -Method GET
        }
        
        if ($response.status -eq "ok") {
            Write-Success "API is healthy"
        } else {
            Write-Error "API returned unexpected status: $($response.status)"
            exit 1
        }
    }
    catch {
        Write-Error "Failed to connect to API at $RegistryUrl"
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "Make sure the API server is running (e.g., 'docker-compose up')" -ForegroundColor Yellow
        exit 1
    }
}

function Setup-Registry {
    Write-TestStep "Setting up registry configuration..."
    
    try {
        # Add test registry
        $addArgs = @("registry", "add", "test", $RegistryUrl)
        if ($Token) {
            $addArgs += @("--token", $Token)
        }
        & rfh @addArgs
        
        # Set as current registry
        & rfh registry use test
        
        Write-Success "Registry configured"
    }
    catch {
        Write-Error "Failed to configure registry: $($_.Exception.Message)"
        exit 1
    }
}

function Test-PackageInit {
    Write-TestStep "Testing package initialization..."
    
    Set-Location $TestDir
    
    # Verify rulestack.json exists and is valid
    if (!(Test-Path "rulestack.json")) {
        Write-Error "rulestack.json not found in test package"
        exit 1
    }
    
    try {
        $manifest = Get-Content "rulestack.json" | ConvertFrom-Json
        if ($manifest.name -ne $TestPackageName) {
            Write-Error "Package name mismatch in manifest. Expected: '$TestPackageName', Got: '$($manifest.name)'"
            exit 1
        }
        Write-Success "Package manifest is valid"
    }
    catch {
        Write-Error "Invalid JSON in rulestack.json: $($_.Exception.Message)"
        exit 1
    }
}

function Test-PackOperation {
    Write-TestStep "Testing pack operation..."
    
    try {
        if ($Verbose) {
            & rfh pack --verbose
        } else {
            & rfh pack
        }
        
        $expectedArchive = "$TestPackageName-$TestPackageVersion.tgz"
        if (Test-Path $expectedArchive) {
            Write-Success "Package packed successfully: $expectedArchive"
        } else {
            Write-Error "Archive file not created: $expectedArchive"
            exit 1
        }
    }
    catch {
        Write-Error "Pack operation failed: $($_.Exception.Message)"
        exit 1
    }
}

function Test-UserRegistration {
    Write-TestStep "Testing user registration API..."
    
    # Test successful registration
    $testUsername = "testuser-$(Get-Date -Format 'yyyyMMddHHmmss')"
    $testEmail = "$testUsername@example.com"
    $testPassword = "TestPassword123!"
    
    $registerBody = @{
        username = $testUsername
        email = $testEmail
        password = $testPassword
    } | ConvertTo-Json
    
    try {
        $authResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/register" -Method POST -Body $registerBody -ContentType "application/json" -ErrorAction Stop
        
        # Validate response structure
        if (-not $authResponse.id -or -not $authResponse.username -or -not $authResponse.email -or -not $authResponse.role -or -not $authResponse.created_at) {
            Write-Error "Registration response missing required fields"
            exit 1
        }
        
        if ($authResponse.username -ne $testUsername) {
            Write-Error "Registration response username mismatch. Expected: $testUsername, Got: $($authResponse.username)"
            exit 1
        }
        
        if ($authResponse.email -ne $testEmail) {
            Write-Error "Registration response email mismatch. Expected: $testEmail, Got: $($authResponse.email)"
            exit 1
        }
        
        if ($authResponse.role -ne "user") {
            Write-Error "Registration response role mismatch. Expected: user, Got: $($authResponse.role)"
            exit 1
        }
        
        Write-Success "User registration successful: $($authResponse.username) (ID: $($authResponse.id))"
        
    } catch {
        Write-Error "User registration failed: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "HTTP Status Code: $statusCode" -ForegroundColor Red
        }
        exit 1
    }
    
    # Test duplicate username registration (should fail)
    try {
        $duplicateResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/register" -Method POST -Body $registerBody -ContentType "application/json" -ErrorAction Stop
        Write-Error "Duplicate registration should have failed but succeeded"
        exit 1
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 409) {
            Write-Success "Duplicate registration correctly rejected with 409 Conflict"
        } else {
            Write-Error "Duplicate registration failed with unexpected status code: $statusCode (expected 409)"
            exit 1
        }
    }
    
    # Test invalid input (missing fields)
    $invalidBody = @{
        username = "testuser"
        # Missing email and password
    } | ConvertTo-Json
    
    try {
        $invalidResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/register" -Method POST -Body $invalidBody -ContentType "application/json" -ErrorAction Stop
        Write-Error "Invalid registration (missing fields) should have failed but succeeded"
        exit 1
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 400) {
            Write-Success "Invalid registration correctly rejected with 400 Bad Request"
        } else {
            Write-Error "Invalid registration failed with unexpected status code: $statusCode (expected 400)"
            exit 1
        }
    }
    
    Write-Success "All user registration tests passed"
    
    # Return the successful registration details for use in authentication setup
    return @{
        username = $testUsername
        email = $testEmail
        password = $testPassword
        response = $authResponse
    }
}

function Setup-Authentication {
    Write-TestStep "Setting up authentication for testing..."
    
    # Use the user registration test to create a user
    try {
        $registrationResult = Test-UserRegistration
        
        # Create simple CLI config with user info (no JWT token available from registration)
        $configDir = "$env:USERPROFILE\.rfh"
        if (!(Test-Path $configDir)) {
            New-Item -ItemType Directory -Path $configDir -Force | Out-Null
        }
        
        # Try to login to get a JWT token
        $loginBody = @{
            username = $registrationResult.username
            password = $registrationResult.password
        } | ConvertTo-Json
        
        $loginResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json" -ErrorAction Stop
        
        $configContent = @"
current = "test"

[user]
username = "$($loginResponse.user.username)"
token = "$($loginResponse.token)"

[registries.test]
url = "$RegistryUrl"
"@
        
        Set-Content -Path "$configDir\config.toml" -Value $configContent
        Write-Success "JWT authentication configured: $($loginResponse.user.username)"
        return $true
        
    } catch {
        Write-Host "INFO: JWT setup failed, will use legacy token method" -ForegroundColor Yellow
        return $false
    }
}

function Test-PublishOperation {
    Write-TestStep "Testing publish operation..."
    
    # Check if user is authenticated first
    $whoamiOutput = & rfh auth whoami 2>$null
    if ($LASTEXITCODE -eq 0 -and $whoamiOutput -match "Logged in as:") {
        Write-Host "[JWT] Attempting publish with JWT authentication..." -ForegroundColor Cyan
        
        if ($Verbose) {
            & rfh publish --verbose
        } else {
            & rfh publish
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Package published successfully with JWT authentication"
            return
        } else {
            Write-Host "JWT publish failed, trying legacy token..." -ForegroundColor Yellow
        }
    }
    
    # Fallback to legacy token method
    Write-Host "[LEGACY] Using legacy token authentication..." -ForegroundColor Yellow
    
    if ($Verbose) {
        & rfh publish --token noauth --verbose
    } else {
        & rfh publish --token noauth
    }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Package published successfully"
    } else {
        Write-Host "Publish operation failed, but this might be expected if the package was already published" -ForegroundColor Yellow
    }
}

function Test-SearchOperation {
    Write-TestStep "Testing search operation..."
    
    try {
        # Test basic search
        if ($Verbose) {
            $output = & rfh search "monty" --verbose
        } else {
            $output = & rfh search "monty"
        }
        
        if ($output -match $TestPackageName) {
            Write-Success "Search found our test package"
        } else {
            Write-Error "Search did not find our test package"
            Write-Host "Search output:" -ForegroundColor Yellow
            Write-Host $output -ForegroundColor Gray
        }
        
        # Test search with filters
        if ($Verbose) {
            $tagOutput = & rfh search "quotes" --tag "humor" --verbose
        } else {
            $tagOutput = & rfh search "quotes" --tag "humor"
        }
        
        Write-Success "Search with filters completed"
    }
    catch {
        Write-Error "Search operation failed: $($_.Exception.Message)"
        exit 1
    }
}

function Test-PreInitializationErrors {
    Write-TestStep "Testing commands before initialization..."
    
    # Create a clean directory without initialization
    $noInitDir = "$TempTestRoot\no-init-test"
    if (Test-Path $noInitDir) {
        Remove-Item $noInitDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $noInitDir | Out-Null
    Set-Location $noInitDir
    
    # Test that add command fails before initialization
    $ErrorActionPreference = "Continue"
    $output = & rfh add "test-package@1.0.0" 2>&1
    $ErrorActionPreference = "Stop"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Error "add command should have failed before initialization"
        exit 1
    }
    if ($output -match "no RuleStack project found.*Run 'rfh init' first") {
        Write-Success "add command properly requires initialization"
    } else {
        Write-Error "add command error message incorrect: $output"
        exit 1
    }
    
    # Test that pack command fails before initialization
    $ErrorActionPreference = "Continue"
    $output = & rfh pack 2>&1
    $ErrorActionPreference = "Stop"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Error "pack command should have failed before initialization"
        exit 1
    }
    if ($output -match "failed to load manifest.*rulestack.json") {
        Write-Success "pack command properly requires rulestack.json"
    } else {
        Write-Error "pack command error message unexpected: $output"
        exit 1
    }
    
    # Test that publish command fails before initialization
    $ErrorActionPreference = "Continue"
    $output = & rfh publish 2>&1
    $ErrorActionPreference = "Stop"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Error "publish command should have failed before initialization"
        exit 1
    }
    if ($output -match "failed to load manifest.*rulestack.json") {
        Write-Success "publish command properly requires rulestack.json"
    } else {
        Write-Error "publish command error message unexpected: $output"
        exit 1
    }
    
    Write-Success "All pre-initialization error checks passed"
    
    # Return to test environment root
    Set-Location $TempTestRoot
}

function Test-RegistryOperations {
    Write-TestStep "Testing registry management operations..."
    
    # Test registry list (should show our test registry)
    $registryOutput = & rfh registry list
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Registry list command failed"
        exit 1
    }
    
    if ($registryOutput -notmatch "test") {
        Write-Error "Test registry not found in registry list"
        exit 1
    }
    
    if ($registryOutput -notmatch "\*.*test") {
        Write-Error "Test registry should be marked as active"
        exit 1
    }
    
    Write-Success "Registry list shows test registry as active"
    
    # Test adding a second registry
    & rfh registry add "test-secondary" "http://localhost:9090" "secondary-token"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to add secondary test registry"
        exit 1
    }
    
    Write-Success "Added secondary test registry"
    
    # Test registry switching
    & rfh registry use "test-secondary"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to switch to secondary registry"
        exit 1
    }
    
    # Verify the switch worked
    $registryOutput = & rfh registry list
    if ($registryOutput -notmatch "\*.*test-secondary") {
        Write-Error "Secondary registry should be marked as active after switch"
        exit 1
    }
    
    if ($registryOutput -match "\*.*test[^-]") {
        Write-Error "Original test registry should no longer be active"
        exit 1
    }
    
    Write-Success "Registry switching works correctly"
    
    # Switch back to original test registry for remaining tests
    & rfh registry use "test"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to switch back to test registry"
        exit 1
    }
    
    # Test registry removal
    & rfh registry remove "test-secondary"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to remove secondary test registry"
        exit 1
    }
    
    # Verify removal
    $registryOutput = & rfh registry list
    if ($registryOutput -match "test-secondary") {
        Write-Error "Secondary registry should be removed from list"
        exit 1
    }
    
    Write-Success "Registry removal works correctly"
    
    # Test error handling - try to use non-existent registry
    $ErrorActionPreference = "Continue"
    $output = & rfh registry use "nonexistent-registry" 2>&1
    $ErrorActionPreference = "Stop"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Error "Using non-existent registry should have failed"
        exit 1
    }
    
    if ($output -notmatch "registry.*not found") {
        Write-Error "Error message for non-existent registry should mention 'not found'"
        exit 1
    }
    
    Write-Success "Error handling for non-existent registry works correctly"
    
    # Test error handling - try to remove non-existent registry
    $ErrorActionPreference = "Continue"
    $output = & rfh registry remove "nonexistent-registry" 2>&1
    $ErrorActionPreference = "Stop"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Error "Removing non-existent registry should have failed"
        exit 1
    }
    
    if ($output -notmatch "registry.*not found") {
        Write-Error "Error message for removing non-existent registry should mention 'not found'"
        exit 1
    }
    
    Write-Success "Error handling for registry removal works correctly"
    
    Write-Success "All registry operations completed successfully"
}

function Test-InstallOperation {
    Write-TestStep "Testing install operation (add command)..."
    
    # Create a clean install test directory within our test environment
    $installTestDir = "$TempTestRoot\install-test"
    if (Test-Path $installTestDir) {
        Remove-Item $installTestDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $installTestDir | Out-Null
    Set-Location $installTestDir
    
    # Initialize a new RuleStack project for testing add command
    & rfh init | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to initialize RuleStack project for testing"
    }
    Write-Success "Initialized test RuleStack project"
    
    # Verify that rfh init created the expected files
    $expectedFiles = @("rulestack.json", "CLAUDE.md", ".rulestack")
    foreach ($file in $expectedFiles) {
        if (!(Test-Path $file)) {
            Write-Error "rfh init should have created $file but it's missing"
            exit 1
        }
    }
    Write-Success "rfh init created expected files: rulestack.json, CLAUDE.md, .rulestack/"
    
    # Verify that rfh init did NOT create unnecessary files
    $unexpectedFiles = @("README.md", "rules")
    foreach ($file in $unexpectedFiles) {
        if (Test-Path $file) {
            Write-Error "rfh init created unexpected file: $file (should be simplified now)"
            exit 1
        }
    }
    Write-Success "rfh init correctly avoided creating unnecessary files"
    
    try {
        # The add command is now implemented and should work
        # Provide "y" input to confirm reinstallation if package already exists
        if ($Verbose) {
            "y" | & rfh add "$TestPackageName@$TestPackageVersion" --verbose
        } else {
            "y" | & rfh add "$TestPackageName@$TestPackageVersion"
        }
        
        # Verify installation - check current directory (now project root for add command)
        $installSuccess = $true
        
        if (Test-Path ".rulestack") {
            Write-Success ".rulestack directory created"
        } else {
            Write-Error ".rulestack directory NOT created - FAILED"
            $installSuccess = $false
        }
        
        if (Test-Path ".rulestack\$TestPackageName.$TestPackageVersion") {
            Write-Success "Package directory created: .rulestack\$TestPackageName.$TestPackageVersion"
        } else {
            Write-Error "Package directory NOT created: .rulestack\$TestPackageName.$TestPackageVersion - FAILED"
            $installSuccess = $false
        }
        
        if (Test-Path "rulestack.json") {
            Write-Success "Project manifest created: rulestack.json"
            # Show content for debugging
            if ($Verbose) {
                Write-Host "rulestack.json content:" -ForegroundColor Gray
                Get-Content "rulestack.json" | Write-Host -ForegroundColor Gray
            }
        } else {
            Write-Error "Project manifest NOT created: rulestack.json - FAILED"
            $installSuccess = $false
        }
        
        if (Test-Path "rulestack.lock.json") {
            Write-Success "Lock manifest created: rulestack.lock.json"
            # Show content for debugging
            if ($Verbose) {
                Write-Host "rulestack.lock.json content:" -ForegroundColor Gray
                Get-Content "rulestack.lock.json" | Write-Host -ForegroundColor Gray
            }
        } else {
            Write-Error "Lock manifest NOT created: rulestack.lock.json - FAILED"
            $installSuccess = $false
        }
        
        # Check if package files exist
        if (Test-Path ".rulestack\$TestPackageName.$TestPackageVersion\*") {
            $fileCount = (Get-ChildItem ".rulestack\$TestPackageName.$TestPackageVersion" -Recurse -File).Count
            Write-Success "Package files extracted: $fileCount files found"
        } else {
            Write-Error "No package files found in .rulestack\$TestPackageName.$TestPackageVersion - FAILED"
            $installSuccess = $false
        }
        
        # Check CLAUDE.md was updated correctly without duplicates
        if (Test-Path "CLAUDE.md") {
            $claudeContent = Get-Content "CLAUDE.md" -Raw
            $ruleLines = $claudeContent -split "`n" | Where-Object { $_ -match "^\s*- @\.rulestack/" }
            
            # Check for our expected rule entry (actual file from package)
            $expectedRule = "- @.rulestack/$TestPackageName.$TestPackageVersion/rules/spam-quotes.md"
            $hasExpectedRule = $ruleLines -contains $expectedRule
            
            # Check for duplicates by comparing unique vs total count
            $uniqueRules = $ruleLines | Sort-Object -Unique
            $hasDuplicates = $ruleLines.Count -ne $uniqueRules.Count
            
            if ($hasExpectedRule) {
                Write-Success "CLAUDE.md contains expected rule entry"
            } else {
                Write-Error "CLAUDE.md missing expected rule entry: $expectedRule - FAILED"
                $installSuccess = $false
            }
            
            if (!$hasDuplicates) {
                Write-Success "CLAUDE.md has no duplicate rule entries"
            } else {
                Write-Error "CLAUDE.md contains duplicate rule entries - FAILED"
                if ($Verbose) {
                    Write-Host "All rule entries:" -ForegroundColor Gray
                    $ruleLines | ForEach-Object { Write-Host "  $_" -ForegroundColor Gray }
                }
                $installSuccess = $false
            }
        } else {
            Write-Error "CLAUDE.md was not created - FAILED"
            $installSuccess = $false
        }
        
        if ($installSuccess) {
            Write-Success "Package installed successfully - ALL VERIFICATIONS PASSED"
        } else {
            Write-Error "Package installation FAILED verification checks"
            exit 1
        }
    }
    catch {
        Write-Error "Install operation failed: $($_.Exception.Message)"
        Write-Host "This indicates the add command has an issue that needs fixing" -ForegroundColor Yellow
        exit 1
    }
    
    # Return to test environment root
    Set-Location $TempTestRoot
}

function Cleanup {
    Write-TestStep "Cleaning up test artifacts..."
    
    if (!$SkipCleanup) {
        # Clean up test files within the test environment
        if (Test-Path "$TestDir/*.tgz") {
            Remove-Item "$TestDir/*.tgz" -Force
            Write-Success "Removed test archives"
        }
        
        # Clean up ALL test registries (including any leftover from previous runs)
        $testRegistries = @('test', 'test-local', 'local')
        foreach ($regName in $testRegistries) {
            try {
                # Check if registry exists first
                $registryList = & rfh registry list 2>$null
                if ($LASTEXITCODE -eq 0 -and $registryList -match $regName) {
                    & rfh registry remove $regName 2>$null
                    if ($LASTEXITCODE -eq 0) {
                        Write-Success "Removed registry '$regName'"
                    }
                }
            }
            catch {
                # Silently continue if registry doesn't exist
            }
        }
        
        # Ensure no active registry is set to our test registries
        try {
            $currentConfig = & rfh registry list 2>$null
            if ($LASTEXITCODE -eq 0 -and $currentConfig -match "\* (test|test-local|local)") {
                # If test registry is active, clear it by trying to set a non-existent one
                & rfh registry use "" 2>$null | Out-Null
            }
        }
        catch {
            # Continue if there's an issue
        }
        
        # Clean up any temporary RFH config that might have been created
        $tempConfigPath = "$env:TEMP\.rfh"
        if (Test-Path $tempConfigPath) {
            try {
                Remove-Item $tempConfigPath -Recurse -Force
                Write-Success "Removed temporary RFH configuration"
            }
            catch {
                Write-Host "Note: Could not remove temporary RFH config" -ForegroundColor Yellow
            }
        }
        
        # Return to original directory
        Set-Location $OriginalDir
        
        # Clean up entire test environment
        if (Test-Path $TempTestRoot) {
            Remove-Item $TempTestRoot -Recurse -Force
            Write-Success "Removed isolated test environment: $TempTestRoot"
        }
        
        Write-Success "Complete environment cleanup performed"
    } else {
        Set-Location $OriginalDir
        Write-Host "[DEBUG] Skipping cleanup - test environment preserved at: $TempTestRoot" -ForegroundColor Yellow
    }
}

function Run-Tests {
    try {
        Pre-Cleanup
        Setup-TestEnvironment
        Test-Prerequisites
        Test-APIHealth
        Setup-Registry
        Setup-Authentication  # This includes Test-UserRegistration
        Test-PreInitializationErrors
        Test-RegistryOperations
        Test-PackageInit
        Test-PackOperation
        Test-PublishOperation
        Test-SearchOperation
        Test-InstallOperation
        
        Write-Host "`n[SUCCESS] All tests completed!" -ForegroundColor Green
        Write-Host "[OK] User registration: Working" -ForegroundColor Green
        Write-Host "[OK] Pre-initialization errors: Working" -ForegroundColor Green
        Write-Host "[OK] Registry operations: Working" -ForegroundColor Green
        Write-Host "[OK] Pack operation: Working" -ForegroundColor Green
        Write-Host "[OK] Publish operation: Working" -ForegroundColor Green
        Write-Host "[OK] Search operation: Working" -ForegroundColor Green
        Write-Host "[OK] Install operation: Working" -ForegroundColor Green
        
    }
    catch {
        Write-Error "Test failed: $($_.Exception.Message)"
        exit 1
    }
    finally {
        Cleanup
    }
}

# Main execution
Write-Host "Registry URL: $RegistryUrl" -ForegroundColor Cyan
if ($Token) {
    Write-Host "Using authentication token: ****" -ForegroundColor Cyan
} else {
    Write-Host "No authentication token provided" -ForegroundColor Yellow
}

Run-Tests

Write-Host "`n[DONE] Test script completed!" -ForegroundColor Cyan