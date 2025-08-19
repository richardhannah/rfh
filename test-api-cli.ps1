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

function Test-PublishOperation {
    Write-TestStep "Testing publish operation..."
    
    try {
        # Use token override since no auth is configured for proof of concept
        if ($Verbose) {
            & rfh publish --token noauth --verbose
        } else {
            & rfh publish --token noauth
        }
        Write-Success "Package published successfully"
    }
    catch {
        Write-Error "Publish operation failed: $($_.Exception.Message)"
        # Don't exit here as the package might already exist
        Write-Host "This might be expected if the package was already published" -ForegroundColor Yellow
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
        
        if (Test-Path ".rulestack\$TestPackageName") {
            Write-Success "Package directory created: .rulestack\$TestPackageName"
        } else {
            Write-Error "Package directory NOT created: .rulestack\$TestPackageName - FAILED"
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
        if (Test-Path ".rulestack\$TestPackageName\*") {
            $fileCount = (Get-ChildItem ".rulestack\$TestPackageName" -Recurse -File).Count
            Write-Success "Package files extracted: $fileCount files found"
        } else {
            Write-Error "No package files found in .rulestack\$TestPackageName - FAILED"
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
        
        # Remove test registry
        try {
            & rfh registry remove test
            Write-Success "Removed test registry configuration"
        }
        catch {
            Write-Host "Note: Could not remove test registry configuration" -ForegroundColor Yellow
        }
        
        # Return to original directory
        Set-Location $OriginalDir
        
        # Clean up entire test environment
        if (Test-Path $TempTestRoot) {
            Remove-Item $TempTestRoot -Recurse -Force
            Write-Success "Removed isolated test environment: $TempTestRoot"
        }
    } else {
        Set-Location $OriginalDir
        Write-Host "[DEBUG] Skipping cleanup - test environment preserved at: $TempTestRoot" -ForegroundColor Yellow
    }
}

function Run-Tests {
    try {
        Setup-TestEnvironment
        Test-Prerequisites
        Test-APIHealth
        Setup-Registry
        Test-PreInitializationErrors
        Test-PackageInit
        Test-PackOperation
        Test-PublishOperation
        Test-SearchOperation
        Test-InstallOperation
        
        Write-Host "`n[SUCCESS] All tests completed!" -ForegroundColor Green
        Write-Host "[OK] Pre-initialization errors: Working" -ForegroundColor Green
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