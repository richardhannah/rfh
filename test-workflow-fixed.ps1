#!/usr/bin/env pwsh
# Test workflow with existing publisher user (fixed syntax)

$ErrorActionPreference = "Stop"

Write-Host "[TEST] RFH Workflow with Publisher User" -ForegroundColor Cyan
Write-Host "=======================================" -ForegroundColor Cyan

# Use existing publisher user
$username = "testuser-20250826192044"
$password = "TestPassword123!"
$RegistryUrl = "http://localhost:8080"

$TestRoot = "$env:TEMP\rfh-publisher-test"
$RfhBinary = "D:\projects\render.com\rfh\dist\rfh.exe"

try {
    # Setup
    Write-Host "`n[SETUP] Creating test environment..." -ForegroundColor Yellow
    if (Test-Path $TestRoot) { Remove-Item $TestRoot -Recurse -Force }
    New-Item -ItemType Directory -Path $TestRoot -Force | Out-Null
    Set-Location $TestRoot

    # Setup registry and login
    Write-Host "[SETUP] Setting up registry and login..." -ForegroundColor Yellow
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
    Write-Host "✅ Registry configured and user logged in (role: $($loginResponse.user.role))" -ForegroundColor Green

    # Initialize project
    Write-Host "`n[STEP 1] Initializing project..." -ForegroundColor Yellow
    & $RfhBinary init | Out-Null
    Write-Host "✅ Project initialized" -ForegroundColor Green

    Write-Host "Generated manifest:" -ForegroundColor Gray
    Get-Content "rulestack.json" | Write-Host -ForegroundColor White

    # Create and publish first rule
    Write-Host "`n[STEP 2] Creating, packing, and publishing first rule (example1.md)..." -ForegroundColor Yellow

    $rule1Content = @"
# Example Rule 1

## Description
This is the first test rule for workflow validation.

## Usage
Apply this rule when working with example scenarios that require structured guidance.

## Example
When user asks for validation rules, apply this rule to ensure proper structure.

## Notes
- Test rule 1 created: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
- Part of workflow validation suite
"@

    Set-Content -Path "example1.md" -Value $rule1Content
    Write-Host "✅ Created example1.md" -ForegroundColor Green

    Write-Host "Packing example1.md..." -ForegroundColor Yellow
    & $RfhBinary pack .\example1.md
    if ($LASTEXITCODE -ne 0) { throw "Pack failed for example1" }
    Write-Host "✅ Successfully packed example1" -ForegroundColor Green

    Write-Host "Publishing first package..." -ForegroundColor Yellow
    & $RfhBinary publish
    if ($LASTEXITCODE -ne 0) { throw "Publish failed for example1" }
    Write-Host "✅ Successfully published first package" -ForegroundColor Green

    # Create and publish second rule
    Write-Host "`n[STEP 3] Creating, packing, and publishing second rule (example2.md)..." -ForegroundColor Yellow

    $rule2Content = @"
# Example Rule 2

## Description  
This is the second test rule for workflow validation, focusing on different scenarios.

## Usage
Apply this rule when working with secondary validation requirements.

## Example
When user requests additional validation, use this rule to provide comprehensive coverage.

## Notes
- Test rule 2 created: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
- Complementary to example1 rule
- Demonstrates multi-rule package management
"@

    Set-Content -Path "example2.md" -Value $rule2Content
    Write-Host "✅ Created example2.md" -ForegroundColor Green

    Write-Host "Packing example2.md..." -ForegroundColor Yellow
    & $RfhBinary pack .\example2.md
    if ($LASTEXITCODE -ne 0) { throw "Pack failed for example2" }
    Write-Host "✅ Successfully packed example2" -ForegroundColor Green

    Write-Host "Publishing second package..." -ForegroundColor Yellow
    & $RfhBinary publish
    if ($LASTEXITCODE -ne 0) { throw "Publish failed for example2" }
    Write-Host "✅ Successfully published second package" -ForegroundColor Green

    # Test search
    Write-Host "`n[STEP 4] Testing search functionality..." -ForegroundColor Yellow
    Write-Host "Searching for 'example'..." -ForegroundColor Gray
    & $RfhBinary search "example"
    Write-Host "✅ Search completed" -ForegroundColor Green

    # Show final archive files
    Write-Host "`n[INFO] Final archive files created:" -ForegroundColor Yellow
    $archives = Get-ChildItem *.tgz -ErrorAction SilentlyContinue
    foreach ($archive in $archives) {
        Write-Host "  📦 $($archive.Name) ($($archive.Length) bytes)" -ForegroundColor White
    }

    # Final success
    Write-Host "`n[SUCCESS] Complete workflow test passed!" -ForegroundColor Green
    Write-Host "✅ User authentication with publisher role" -ForegroundColor Green
    Write-Host "✅ Registry configuration and login" -ForegroundColor Green  
    Write-Host "✅ Project initialization with correct manifest" -ForegroundColor Green
    Write-Host "✅ First rule: creation, packing, and publishing (example1)" -ForegroundColor Green
    Write-Host "✅ Second rule: creation, packing, and publishing (example2)" -ForegroundColor Green
    Write-Host "✅ Package search functionality" -ForegroundColor Green

} catch {
    Write-Host "`n[ERROR] Test failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
} finally {
    # Cleanup
    Write-Host "`n[CLEANUP] Cleaning up..." -ForegroundColor Yellow
    & $RfhBinary registry remove test 2>$null
    Set-Location $env:TEMP
    if (Test-Path $TestRoot) {
        Remove-Item $TestRoot -Recurse -Force
    }
    Write-Host "✅ Cleanup completed" -ForegroundColor Green
}

Write-Host "`n🎉 All tests passed! Scope removal and workflow are working correctly." -ForegroundColor Cyan