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

# Check if Docker is available and handle API server
$dockerAvailable = $false
try {
    docker info *> $null
    if ($LASTEXITCODE -eq 0) {
        $dockerAvailable = $true
        Write-Host "Docker is available - managing test infrastructure..." -ForegroundColor Cyan
        
        # Clean up any existing test containers and start fresh
        Write-Host "Cleaning up existing test containers..." -ForegroundColor Yellow
        docker-compose -f docker-compose.test.yml down -v *> $null
        
        # Build and start test infrastructure
        Write-Host "Building and starting test API server..." -ForegroundColor Yellow
        docker-compose -f docker-compose.test.yml up --build -d *> $null
        
        # Wait for API to be healthy
        Write-Host -NoNewline "Waiting for test API to be ready"
        $apiReady = $false
        for ($i = 0; $i -lt 60; $i++) {
            try {
                $response = Invoke-WebRequest -Uri "http://localhost:8081/v1/health" -Method Get -ErrorAction SilentlyContinue
                if ($response.StatusCode -eq 200) {
                    $apiReady = $true
                    Write-Host " OK" -ForegroundColor Green
                    break
                }
            } catch {}
            Write-Host -NoNewline "."
            Start-Sleep -Seconds 1
        }
        
        if (-not $apiReady) {
            Write-Host ""
            Write-Host "Warning: API server not responding - some tests may fail" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "Docker not found or not running - skipping API server setup" -ForegroundColor Yellow
    Write-Host "Some registry/auth tests may fail without the API server" -ForegroundColor Yellow
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

# Clean up test infrastructure if Docker was used
if ($dockerAvailable) {
    Write-Host ""
    Write-Host "Cleaning up test infrastructure..." -ForegroundColor Cyan
    docker-compose -f docker-compose.test.yml down -v *> $null
    Write-Host "Test containers cleaned up" -ForegroundColor Green
}

exit $TestExitCode