#!/usr/bin/env pwsh
# Debug script to investigate the manifest loading issue

Write-Host "=== Debug: Manifest Loading Issue ===" -ForegroundColor Cyan

$TestDir = "D:\temp\rfh-manifest-debug"

# Clean up and create test directory
Write-Host "Creating test environment..." -ForegroundColor Yellow
if (Test-Path $TestDir) {
    Remove-Item $TestDir -Recurse -Force
}
New-Item -ItemType Directory -Path $TestDir | Out-Null
Set-Location $TestDir

# Test 1: Complete clean directory (should work)
Write-Host "`n=== Test 1: Clean directory ===" -ForegroundColor Cyan
"# Test Rule`nThis is a test rule." | Set-Content "example-rule.md"

Write-Host "Directory contents:" -ForegroundColor Yellow
Get-ChildItem | Format-Table Name, Length

Write-Host "Running pack command:" -ForegroundColor Yellow
& "D:\projects\render.com\rfh\dist\rfh.exe" pack .\example-rule.md --verbose 2>&1

# Test 2: After init (should work with new init)
Write-Host "`n=== Test 2: After init ===" -ForegroundColor Cyan
Remove-Item * -Force
"# Test Rule`nThis is a test rule." | Set-Content "example-rule.md"

Write-Host "Running init:" -ForegroundColor Yellow
& "D:\projects\render.com\rfh\dist\rfh.exe" init

Write-Host "Generated rulestack.json:" -ForegroundColor Yellow
if (Test-Path "rulestack.json") {
    Get-Content "rulestack.json" | Write-Host -ForegroundColor Gray
} else {
    Write-Host "No rulestack.json found" -ForegroundColor Red
}

Write-Host "Running pack command after init:" -ForegroundColor Yellow
& "D:\projects\render.com\rfh\dist\rfh.exe" pack .\example-rule.md --verbose 2>&1

# Test 3: Corrupted manifest (simulate the issue)
Write-Host "`n=== Test 3: Corrupted manifest ===" -ForegroundColor Cyan
Remove-Item * -Force
"# Test Rule`nThis is a test rule." | Set-Content "example-rule.md"

# Create a manifest with missing name field
@"
{
  "version": "0.1.0",
  "description": "Test ruleset",
  "files": ["*.md"]
}
"@ | Set-Content "rulestack.json"

Write-Host "Corrupted rulestack.json (missing name):" -ForegroundColor Yellow
Get-Content "rulestack.json" | Write-Host -ForegroundColor Gray

Write-Host "Running pack command with corrupted manifest:" -ForegroundColor Yellow
& "D:\projects\render.com\rfh\dist\rfh.exe" pack .\example-rule.md --verbose 2>&1

Write-Host "`n=== Debug Complete ===" -ForegroundColor Cyan