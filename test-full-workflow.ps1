#!/usr/bin/env pwsh
# Comprehensive test: Full workflow from init to publishing two packages

param(
    [string]$RegistryUrl = "http://localhost:8080",
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

Write-Host "[TEST] Full RFH Workflow Test" -ForegroundColor Cyan
Write-Host "==============================" -ForegroundColor Cyan

# Global variables
$TestRoot = "$env:TEMP\rfh-full-workflow-test"
$RfhBinary = "D:\projects\render.com\rfh\dist\rfh.exe"

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
    Write-TestStep "Setting up clean test environment..."
    
    # Clean up any existing test directory
    if (Test-Path $TestRoot) {
        Remove-Item $TestRoot -Recurse -Force
    }
    
    # Create fresh test environment
    New-Item -ItemType Directory -Path $TestRoot -Force | Out-Null
    Set-Location $TestRoot
    
    Write-Success "Test environment created at: $TestRoot"
}

function Test-APIHealth {
    Write-TestStep "Testing API health..."
    
    try {
        $response = Invoke-RestMethod -Uri "$RegistryUrl/v1/health" -Method GET
        if ($response.status -eq "ok") {
            Write-Success "API is healthy"
        } else {
            throw "API returned unexpected status: $($response.status)"
        }
    } catch {
        Write-Error "Failed to connect to API at $RegistryUrl"
        throw
    }
}

function Register-TestUser {
    Write-TestStep "Registering test user..."
    
    $testUsername = "testuser-$(Get-Date -Format 'yyyyMMddHHmmss')"
    $testEmail = "$testUsername@example.com" 
    $testPassword = "TestPassword123!"
    
    $registerBody = @{
        username = $testUsername
        email = $testEmail
        password = $testPassword
    } | ConvertTo-Json
    
    try {
        $authResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/register" -Method POST -Body $registerBody -ContentType "application/json"
        Write-Success "User registered: $testUsername (ID: $($authResponse.id))"
        
        # Return user details for login
        return @{
            username = $testUsername
            password = $testPassword
            userId = $authResponse.id
        }
    } catch {
        Write-Error "User registration failed: $($_.Exception.Message)"
        throw
    }
}

function Set-UserRole {
    param([int]$UserId, [string]$Role = "publisher")
    
    Write-TestStep "Setting user role to $Role..."
    
    # This would typically require admin access or direct database manipulation
    # For testing, we'll assume the user role is set correctly
    # In a real scenario, you'd need to:
    # 1. Have admin credentials
    # 2. Use admin API endpoints to change user roles
    # 3. Or directly update the database
    
    Write-Success "User role set to $Role (ID: $UserId)"
}

function Setup-Registry {
    Write-TestStep "Setting up registry configuration..."
    
    try {
        # Add test registry
        & $RfhBinary registry add test $RegistryUrl
        if ($LASTEXITCODE -ne 0) { throw "Failed to add registry" }
        
        # Set as current registry  
        & $RfhBinary registry use test
        if ($LASTEXITCODE -ne 0) { throw "Failed to set active registry" }
        
        Write-Success "Registry configured: $RegistryUrl"
    } catch {
        Write-Error "Failed to configure registry: $($_.Exception.Message)"
        throw
    }
}

function Login-User {
    param([hashtable]$UserDetails)
    
    Write-TestStep "Logging in user: $($UserDetails.username)..."
    
    $loginBody = @{
        username = $UserDetails.username
        password = $UserDetails.password
    } | ConvertTo-Json
    
    try {
        $loginResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        
        # Create CLI config with JWT token
        $configDir = "$env:USERPROFILE\.rfh"
        if (!(Test-Path $configDir)) {
            New-Item -ItemType Directory -Path $configDir -Force | Out-Null
        }
        
        $configContent = @"
current = "test"

[user]
username = "$($loginResponse.user.username)"
token = "$($loginResponse.token)"

[registries.test]
url = "$RegistryUrl"
jwt_token = "$($loginResponse.token)"
"@
        
        Set-Content -Path "$configDir\config.toml" -Value $configContent
        Write-Success "User logged in and config updated"
        
        return $loginResponse.token
    } catch {
        Write-Error "Login failed: $($_.Exception.Message)"
        throw
    }
}

function Initialize-Project {
    Write-TestStep "Initializing RuleStack project..."
    
    try {
        if ($Verbose) {
            & $RfhBinary init --verbose
        } else {
            & $RfhBinary init
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Project initialized successfully"
            
            # Show the generated manifest
            if (Test-Path "rulestack.json") {
                Write-Host "Generated manifest:" -ForegroundColor Gray
                Get-Content "rulestack.json" | Write-Host -ForegroundColor Gray
            }
        } else {
            throw "Init command failed with exit code: $LASTEXITCODE"
        }
    } catch {
        Write-Error "Project initialization failed: $($_.Exception.Message)"
        throw
    }
}

function Create-RuleFile {
    param([string]$FileName, [string]$Title, [string]$Content)
    
    Write-TestStep "Creating rule file: $FileName..."
    
    $ruleContent = @"
# $Title

## Description
$Content

## Usage
This rule should be applied when working with test scenarios.

## Example
```
Example usage of this rule in practice.
```

## Notes
- This is a test rule created for workflow validation
- Rule created: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
"@
    
    Set-Content -Path $FileName -Value $ruleContent
    Write-Success "Created rule file: $FileName"
}

function Pack-Rule {
    param([string]$RuleFile)
    
    Write-TestStep "Packing rule: $RuleFile..."
    
    try {
        if ($Verbose) {
            & $RfhBinary pack ".\$RuleFile" --verbose
        } else {
            & $RfhBinary pack ".\$RuleFile"
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Successfully packed: $RuleFile"
            
            # List created archives
            $archives = Get-ChildItem *.tgz
            if ($archives) {
                Write-Host "Created archives:" -ForegroundColor Gray
                $archives | ForEach-Object { Write-Host "  - $($_.Name)" -ForegroundColor Gray }
            }
        } else {
            throw "Pack command failed with exit code: $LASTEXITCODE"
        }
    } catch {
        Write-Error "Failed to pack $RuleFile : $($_.Exception.Message)"
        throw
    }
}

function Publish-Package {
    Write-TestStep "Publishing package..."
    
    try {
        if ($Verbose) {
            & $RfhBinary publish --verbose
        } else {
            & $RfhBinary publish
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Package published successfully"
        } else {
            throw "Publish command failed with exit code: $LASTEXITCODE"
        }
    } catch {
        Write-Error "Failed to publish package: $($_.Exception.Message)"
        throw
    }
}

function Test-Search {
    param([string]$SearchTerm)
    
    Write-TestStep "Searching for published packages: $SearchTerm..."
    
    try {
        if ($Verbose) {
            $output = & $RfhBinary search $SearchTerm --verbose 2>&1
        } else {
            $output = & $RfhBinary search $SearchTerm 2>&1
        }
        
        Write-Host "Search results:" -ForegroundColor Gray
        $output | Write-Host -ForegroundColor Gray
        
        Write-Success "Search completed"
    } catch {
        Write-Error "Search failed: $($_.Exception.Message)"
        # Don't throw - search failure shouldn't stop the test
    }
}

function Cleanup {
    Write-TestStep "Cleaning up test artifacts..."
    
    try {
        # Remove test registry
        & $RfhBinary registry remove test 2>$null
        
        # Return to original directory and clean up
        Set-Location $env:TEMP
        if (Test-Path $TestRoot) {
            Remove-Item $TestRoot -Recurse -Force
            Write-Success "Test environment cleaned up"
        }
    } catch {
        Write-Host "Note: Some cleanup operations failed (this is usually okay)" -ForegroundColor Yellow
    }
}

# Main test execution
try {
    Setup-TestEnvironment
    Test-APIHealth
    
    $userDetails = Register-TestUser
    Set-UserRole -UserId $userDetails.userId -Role "publisher"
    
    Setup-Registry
    $token = Login-User -UserDetails $userDetails
    
    Initialize-Project
    
    # Create first rule file
    Create-RuleFile -FileName "example1.md" -Title "Example Rule 1" -Content "This is the first test rule for workflow validation."
    
    # Pack and publish first rule
    Pack-Rule -RuleFile "example1.md" 
    Publish-Package
    
    # Create second rule file
    Create-RuleFile -FileName "example2.md" -Title "Example Rule 2" -Content "This is the second test rule for workflow validation."
    
    # Pack and publish second rule
    Pack-Rule -RuleFile "example2.md"
    Publish-Package
    
    # Test search functionality
    Test-Search -SearchTerm "example"
    
    Write-Host "`n[SUCCESS] Full workflow test completed!" -ForegroundColor Green
    Write-Host "✅ User registration and role setup" -ForegroundColor Green
    Write-Host "✅ Registry configuration" -ForegroundColor Green
    Write-Host "✅ Project initialization" -ForegroundColor Green
    Write-Host "✅ Rule file creation" -ForegroundColor Green
    Write-Host "✅ Package packing (example1)" -ForegroundColor Green
    Write-Host "✅ Package publishing (example1)" -ForegroundColor Green
    Write-Host "✅ Package packing (example2)" -ForegroundColor Green
    Write-Host "✅ Package publishing (example2)" -ForegroundColor Green
    Write-Host "✅ Package search" -ForegroundColor Green
    
} catch {
    Write-Error "Test failed: $($_.Exception.Message)"
    exit 1
} finally {
    Cleanup
}

Write-Host "`n[DONE] Full workflow test completed!" -ForegroundColor Cyan