# RFH Cucumber Test Runner (PowerShell)
param(
    [string]$TestTarget = "all"
)

Write-Host "RFH Cucumber Test Runner" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan

# Check if we're in the correct directory
if (-not (Test-Path "go.mod") -or -not (Test-Path "cucumber-testing")) {
    Write-Host "Error: Please run this script from the RFH root directory" -ForegroundColor Red
    exit 1
}

# Build RFH binary
Write-Host "Building RFH binary..." -ForegroundColor Yellow
go build -o dist/rfh.exe ./cmd/cli
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build RFH binary" -ForegroundColor Red
    exit 1
}
Write-Host "RFH binary built successfully" -ForegroundColor Green

# Navigate to cucumber testing directory
Set-Location cucumber-testing

# Check if dependencies are installed
if (-not (Test-Path "node_modules")) {
    Write-Host "Installing test dependencies..." -ForegroundColor Yellow
    npm install
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to install dependencies" -ForegroundColor Red
        Set-Location ..
        exit 1
    }
}

Write-Host "Running Cucumber tests..." -ForegroundColor Cyan

# Run tests based on target
switch ($TestTarget) {
    "actual" {
        Write-Host "Running actual behavior tests only..." -ForegroundColor Yellow
        npx cucumber-js features/init-actual-behavior.feature --format progress
        break
    }
    "init" {
        Write-Host "Running rfh init tests only..." -ForegroundColor Yellow
        npx cucumber-js "features/init-*.feature" --format progress
        break
    }
    "working" {
        Write-Host "Running only working scenarios..." -ForegroundColor Yellow
        npx cucumber-js features/init-actual-behavior.feature --format progress
        break
    }
    default {
        Write-Host "Running all available tests..." -ForegroundColor Yellow
        npm test
        break
    }
}

$TestExitCode = $LASTEXITCODE
Set-Location ..

Write-Host "Test execution completed!" -ForegroundColor Green
Write-Host "Check output above for results"

exit $TestExitCode