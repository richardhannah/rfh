#!/usr/bin/env pwsh

Write-Host "Installing CLI..." -ForegroundColor Green

# Create dist folder if it doesn't exist
if (!(Test-Path "dist")) {
    Write-Host "Creating dist folder..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path "dist" | Out-Null
}

# Build the CLI executable
Write-Host "Building CLI executable..." -ForegroundColor Yellow
$buildResult = & go build -o "dist/rfh.exe" cmd/cli/main.go

if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build CLI executable" -ForegroundColor Red
    exit 1
}

Write-Host "CLI executable built successfully at dist/rfh.exe" -ForegroundColor Green

# Get the full path to the dist directory
$distPath = (Resolve-Path "dist").Path

# Check if dist path is already in PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -split [IO.Path]::PathSeparator -contains $distPath) {
    Write-Host "Dist folder is already in PATH" -ForegroundColor Yellow
} else {
    Write-Host "Adding dist folder to PATH..." -ForegroundColor Yellow
    
    try {
        # Add to user PATH permanently
        $newPath = $currentPath + [IO.Path]::PathSeparator + $distPath
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        
        # Add to current session PATH
        $env:PATH += [IO.Path]::PathSeparator + $distPath
        
        Write-Host "Successfully added $distPath to user PATH" -ForegroundColor Green
        Write-Host "You may need to restart your terminal for the changes to take effect" -ForegroundColor Yellow
    }
    catch {
        Write-Host "Failed to add to permanent PATH. You can manually add: $distPath" -ForegroundColor Red
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host "You can now run 'rfh' from anywhere in your terminal" -ForegroundColor Green
Write-Host ""