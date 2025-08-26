#!/usr/bin/env pwsh
# Register a test user and provide SQL for role assignment

$RegistryUrl = "http://localhost:8080"
$username = "testuser-$(Get-Date -Format 'yyyyMMddHHmmss')"
$email = "$username@example.com"
$password = "TestPassword123!"

$registerBody = @{
    username = $username
    email = $email
    password = $password
} | ConvertTo-Json

Write-Host "Registering user..." -ForegroundColor Yellow
$authResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/auth/register" -Method POST -Body $registerBody -ContentType "application/json"

Write-Host "✅ User registered successfully!" -ForegroundColor Green
Write-Host "Username: $username" -ForegroundColor Cyan
Write-Host "Email: $email" -ForegroundColor Cyan
Write-Host "User ID: $($authResponse.id)" -ForegroundColor Cyan
Write-Host "Current Role: $($authResponse.role)" -ForegroundColor Cyan

Write-Host "`nTo set publisher role, run this SQL command:" -ForegroundColor Yellow
Write-Host "UPDATE rulestack.users SET role = 'publisher' WHERE id = $($authResponse.id);" -ForegroundColor White

Write-Host "`nUser credentials for testing:" -ForegroundColor Yellow
Write-Host "Username: $username" -ForegroundColor White
Write-Host "Password: $password" -ForegroundColor White