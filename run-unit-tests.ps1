param(
    [switch]$Coverage,
    [switch]$Verbose,
    [switch]$Race,
    [string]$Package = "./...",
    [string]$TestName = "",
    [switch]$Short
)

# Colors for output
$Red = [System.ConsoleColor]::Red
$Green = [System.ConsoleColor]::Green
$Yellow = [System.ConsoleColor]::Yellow
$Blue = [System.ConsoleColor]::Blue

function Write-ColorOutput {
    param(
        [string]$Message,
        [System.ConsoleColor]$ForegroundColor = [System.ConsoleColor]::White
    )
    $currentColor = $Host.UI.RawUI.ForegroundColor
    $Host.UI.RawUI.ForegroundColor = $ForegroundColor
    Write-Output $Message
    $Host.UI.RawUI.ForegroundColor = $currentColor
}

# Check if Go is installed
try {
    $goVersion = go version 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Go not found"
    }
    Write-ColorOutput "Using: $goVersion" $Blue
} catch {
    Write-ColorOutput "Error: Go is not installed or not in PATH" $Red
    exit 1
}

# Check if we're in the correct directory
if (-not (Test-Path "go.mod")) {
    Write-ColorOutput "Error: Please run this script from the RFH root directory" $Red
    exit 1
}

# Build test command
$testArgs = @("test")
$testArgs += $Package

if ($Coverage) {
    $testArgs += "-cover"
    $testArgs += "-coverprofile=coverage.out"
}

if ($Verbose) {
    $testArgs += "-v"
}

if ($Race) {
    $testArgs += "-race"
}

if ($Short) {
    $testArgs += "-short"
}

if ($TestName -ne "") {
    $testArgs += "-run"
    $testArgs += $TestName
}

$testArgs += "-timeout"
$testArgs += "30s"

Write-ColorOutput "Running Go unit tests..." $Blue
Write-ColorOutput "Command: go $($testArgs -join ' ')" $Yellow

$startTime = Get-Date
& go @testArgs
$exitCode = $LASTEXITCODE
$endTime = Get-Date
$duration = $endTime - $startTime

if ($exitCode -eq 0) {
    Write-ColorOutput "Tests passed in $($duration.TotalSeconds.ToString('F2')) seconds" $Green
    
    if ($Coverage -and (Test-Path "coverage.out")) {
        Write-ColorOutput "Coverage Report:" $Blue
        & go tool cover "-func=coverage.out"
        
        Write-ColorOutput "Generating HTML coverage report..." $Blue
        & go tool cover "-html=coverage.out" "-o=coverage.html"
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "HTML coverage report generated: coverage.html" $Green
        }
    }
} else {
    Write-ColorOutput "Tests failed after $($duration.TotalSeconds.ToString('F2')) seconds" $Red
    Write-ColorOutput "Exit code: $exitCode" $Red
}

if ($Coverage -and (Test-Path "coverage.out")) {
    Write-ColorOutput "Coverage file: coverage.out" $Yellow
}

exit $exitCode