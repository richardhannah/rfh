#!/usr/bin/env pwsh

Write-Host "[SECURITY TEST] RuleStack Security Validation" -ForegroundColor Cyan
Write-Host "Testing security features with unit tests..." -ForegroundColor White

# Run the security tests  
Write-Host "`nRunning security validation tests:" -ForegroundColor Yellow
$testResult = go test ./internal/security -v 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nAll security tests PASSED!" -ForegroundColor Green
    Write-Host "`nSecurity features successfully protect against:" -ForegroundColor White
    Write-Host "- Path traversal attacks" -ForegroundColor Gray  
    Write-Host "- Executable files (.sh, .exe, .dll, etc.)" -ForegroundColor Gray
    Write-Host "- Malicious markdown (XSS, scripts, iframes)" -ForegroundColor Gray
    Write-Host "- Binary executables (ELF, PE headers)" -ForegroundColor Gray
    Write-Host "- Oversized files and archives" -ForegroundColor Gray
    Write-Host "- Invalid UTF-8 and NUL bytes" -ForegroundColor Gray
    Write-Host "- Symlinks and special file types" -ForegroundColor Gray
    Write-Host "- Too many files in archives" -ForegroundColor Gray
} else {
    Write-Host "`nSecurity tests FAILED!" -ForegroundColor Red
    Write-Host $testResult
}

Write-Host "`nTesting that normal packages still work:" -ForegroundColor Yellow
$normalTest = powershell -File test-api-cli.ps1 2>&1 | Select-String "SUCCESS|FAILED"

if ($normalTest -match "SUCCESS") {
    Write-Host "Normal packages work correctly with security validation" -ForegroundColor Green
} else {
    Write-Host "Normal packages are being blocked inappropriately" -ForegroundColor Red
}

Write-Host "`n[SECURITY VALIDATION COMPLETE]" -ForegroundColor Cyan