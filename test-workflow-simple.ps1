#!/usr/bin/env pwsh
# Simplified workflow test with manual role setup instructions

param(
    [string]$RegistryUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"

Write-Host "[TEST] Simplified RFH Workflow Test" -ForegroundColor Cyan
Write-Host "====================================" -ForegroundColor Cyan

$TestRoot = "$env:TEMP\rfh-simple-workflow"
$RfhBinary = "D:\projects\render.com\rfh\dist\rfh.exe"

# Setup
Write-Host "`n[SETUP] Creating test environment..." -ForegroundColor Yellow
if (Test-Path $TestRoot) { Remove-Item $TestRoot -Recurse -Force }
New-Item -ItemType Directory -Path $TestRoot -Force | Out-Null
Set-Location $TestRoot

# Check API
Write-Host "[SETUP] Checking API health..." -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$RegistryUrl/v1/health" -Method GET
if ($response.status -ne "ok") { throw "API not healthy" }
Write-Host "✅ API is healthy" -ForegroundColor Green

# Register user
Write-Host "`n[STEP 1] Registering user..." -ForegroundColor Yellow
$username = "testuser-$(Get-Date -Format 'yyyyMMddHHmmss')"
$email = "$username@example.com"
$password = "TestPassword123!"

$registerBody = @{
    username = $username
    email = $email
    password = $password
} | ConvertTo-Json

$authResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/register" -Method POST -Body $registerBody -ContentType "application/json"
Write-Host "✅ User registered: $username (ID: $($authResponse.id))" -ForegroundColor Green

# Manual role assignment instruction
Write-Host "`n[MANUAL ACTION REQUIRED]" -ForegroundColor Red
Write-Host "You need to manually set the user role to 'publisher' in the database." -ForegroundColor Yellow
Write-Host "User ID: $($authResponse.id)" -ForegroundColor Yellow
Write-Host "Username: $username" -ForegroundColor Yellow
Write-Host ""
Write-Host "SQL Command to run:" -ForegroundColor Cyan
Write-Host "UPDATE rulestack.users SET role = 'publisher' WHERE id = $($authResponse.id);" -ForegroundColor White
Write-Host ""
Read-Host "Press Enter after updating the user role in the database"

# Setup registry and login
Write-Host "`n[STEP 2] Setting up registry and login..." -ForegroundColor Yellow
& $RfhBinary registry add test $RegistryUrl | Out-Null
& $RfhBinary registry use test | Out-Null

$loginBody = @{
    username = $username
    password = $password
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"

$configDir = "$env:USERPROFILE\.rfh"
if (!(Test-Path $configDir)) { New-Item -ItemType Directory -Path $configDir -Force | Out-Null }

$configContent = @"
current = "test"

[registries.test]
url = "$RegistryUrl"
jwt_token = "$($loginResponse.token)"
"@

Set-Content -Path "$configDir\config.toml" -Value $configContent
Write-Host "✅ Registry configured and user logged in" -ForegroundColor Green

# Initialize project
Write-Host "`n[STEP 3] Initializing project..." -ForegroundColor Yellow
& $RfhBinary init | Out-Null
Write-Host "✅ Project initialized" -ForegroundColor Green

Write-Host "Generated manifest:" -ForegroundColor Gray
Get-Content "rulestack.json" | Write-Host -ForegroundColor White

# Create first rule
Write-Host "`n[STEP 4] Creating and packing first rule (example1.md)..." -ForegroundColor Yellow
@"
# Example Rule 1

## Description
This is the first test rule for workflow validation.

## Usage
This rule should be applied when working with test scenarios.

## Example
``````
Example usage of this rule in practice.
``````

## Notes
- This is a test rule created for workflow validation
- Rule created: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
"@ | Set-Content "example1.md"

Write-Host "✅ Created example1.md" -ForegroundColor Green

# Pack first rule
Write-Host "Packing example1.md..." -ForegroundColor Yellow
& $RfhBinary pack .\example1.md
if ($LASTEXITCODE -ne 0) { throw "Pack failed for example1" }
Write-Host "✅ Successfully packed example1" -ForegroundColor Green

# Publish first rule
Write-Host "Publishing first package..." -ForegroundColor Yellow
& $RfhBinary publish --verbose
if ($LASTEXITCODE -ne 0) { throw "Publish failed for example1" }
Write-Host "✅ Successfully published first package" -ForegroundColor Green

# Create second rule
Write-Host "`n[STEP 5] Creating and packing second rule (example2.md)..." -ForegroundColor Yellow
@"
# Example Rule 2

## Description  
This is the second test rule for workflow validation.

## Usage
This rule should be applied when working with additional test scenarios.

## Example
``````
Example usage of this second rule in practice.
``````

## Notes
- This is the second test rule created for workflow validation
- Rule created: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
"@ | Set-Content "example2.md"

Write-Host "✅ Created example2.md" -ForegroundColor Green

# Pack second rule
Write-Host "Packing example2.md..." -ForegroundColor Yellow
& $RfhBinary pack .\example2.md
if ($LASTEXITCODE -ne 0) { throw "Pack failed for example2" }
Write-Host "✅ Successfully packed example2" -ForegroundColor Green

# Publish second rule
Write-Host "Publishing second package..." -ForegroundColor Yellow
& $RfhBinary publish --verbose
if ($LASTEXITCODE -ne 0) { throw "Publish failed for example2" }
Write-Host "✅ Successfully published second package" -ForegroundColor Green

# Test search
Write-Host "`n[STEP 6] Testing search functionality..." -ForegroundColor Yellow
& $RfhBinary search "example"
Write-Host "✅ Search completed" -ForegroundColor Green

# Final success
Write-Host "`n[SUCCESS] Complete workflow test passed!" -ForegroundColor Green
Write-Host "✅ User registration and manual role setup" -ForegroundColor Green
Write-Host "✅ Registry configuration and login" -ForegroundColor Green  
Write-Host "✅ Project initialization" -ForegroundColor Green
Write-Host "✅ Rule creation, packing, and publishing (example1)" -ForegroundColor Green
Write-Host "✅ Rule creation, packing, and publishing (example2)" -ForegroundColor Green
Write-Host "✅ Package search functionality" -ForegroundColor Green

# Cleanup
Write-Host "`n[CLEANUP] Cleaning up..." -ForegroundColor Yellow
& $RfhBinary registry remove test 2>$null
Set-Location $env:TEMP
Remove-Item $TestRoot -Recurse -Force
Write-Host "✅ Cleanup completed" -ForegroundColor Green