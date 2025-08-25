#!/usr/bin/env pwsh
# Debug script to help identify where @acme/example-rules is coming from

Write-Host "=== Debug: Pack Issue Investigation ===" -ForegroundColor Cyan

$TestDir = "D:\temp\rfh-debug-test"

# Clean up and create test directory
Write-Host "Creating clean test environment..." -ForegroundColor Yellow
if (Test-Path $TestDir) {
    Remove-Item $TestDir -Recurse -Force
}
New-Item -ItemType Directory -Path $TestDir | Out-Null

# Create simple test file
@"
# Example Rule

This is a simple test rule.
"@ | Set-Content "$TestDir\example-rule.md"

Set-Location $TestDir

Write-Host "`nDirectory contents before pack:" -ForegroundColor Yellow
Get-ChildItem -Force | Format-Table Name, Length

Write-Host "`nChecking for any JSON files:" -ForegroundColor Yellow
Get-ChildItem *.json -Force -ErrorAction SilentlyContinue | Format-Table Name, Length

Write-Host "`nRunning rfh pack with verbose output:" -ForegroundColor Yellow
& "D:\projects\render.com\rfh\dist\rfh.exe" pack .\example-rule.md --verbose

Write-Host "`nDirectory contents after pack:" -ForegroundColor Yellow
Get-ChildItem -Force | Format-Table Name, Length

Write-Host "`nIf rulestack.json was created, here's its content:" -ForegroundColor Yellow
if (Test-Path "rulestack.json") {
    Write-Host "--- rulestack.json ---" -ForegroundColor Gray
    Get-Content "rulestack.json" | Write-Host -ForegroundColor Gray
    Write-Host "--- end ---" -ForegroundColor Gray
} else {
    Write-Host "No rulestack.json found" -ForegroundColor Gray
}

Write-Host "`n=== Debug Complete ===" -ForegroundColor Cyan