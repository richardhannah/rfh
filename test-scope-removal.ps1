#!/usr/bin/env pwsh
# Quick test of scope removal functionality

Write-Host "[TEST] Testing scope removal features..." -ForegroundColor Cyan

$RegistryUrl = "http://localhost:8080"

# Test 1: API health 
Write-Host "[STEP] Testing API health..."
try {
    $response = Invoke-RestMethod -Uri "$RegistryUrl/v1/health" -Method GET
    if ($response.status -eq "ok") {
        Write-Host "[OK] API is healthy" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] API returned unexpected status: $($response.status)" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "[ERROR] API health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: Scoped API endpoints should return 404
Write-Host "[STEP] Testing scoped API endpoint rejection..."
try {
    $scopedResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/packages/@myorg/test-package" -Method GET -ErrorAction Stop
    Write-Host "[ERROR] Scoped API endpoint should have returned 404" -ForegroundColor Red
    exit 1
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 404) {
        Write-Host "[OK] Scoped API endpoint correctly returns 404" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Scoped API endpoint returned unexpected status: $statusCode (expected 404)" -ForegroundColor Red
        exit 1
    }
}

# Test 3: Scoped version endpoint should also return 404
Write-Host "[STEP] Testing scoped version API endpoint rejection..."
try {
    $scopedVersionResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/packages/@myorg/test-package/versions/1.0.0" -Method GET -ErrorAction Stop
    Write-Host "[ERROR] Scoped version API endpoint should have returned 404" -ForegroundColor Red
    exit 1
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 404) {
        Write-Host "[OK] Scoped version API endpoint correctly returns 404" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Scoped version API endpoint returned unexpected status: $statusCode (expected 404)" -ForegroundColor Red
        exit 1
    }
}

# Test 4: Search API should work and exclude scope fields
Write-Host "[STEP] Testing search API response format..."
try {
    $searchResponse = Invoke-RestMethod -Uri "$RegistryUrl/v1/packages?q=test" -Method GET -ErrorAction Stop
    Write-Host "[OK] Search API works correctly" -ForegroundColor Green
    
    # Check that search results don't contain scope field
    if ($searchResponse.Count -gt 0) {
        foreach ($result in $searchResponse) {
            if ($result.PSObject.Properties.Name -contains "scope") {
                Write-Host "[ERROR] Search result contains 'scope' field but shouldn't" -ForegroundColor Red
                exit 1
            }
        }
        Write-Host "[OK] Search results correctly exclude scope field (found $($searchResponse.Count) results)" -ForegroundColor Green
    } else {
        Write-Host "[OK] Search completed with no results (scope field validation skipped)" -ForegroundColor Green
    }
} catch {
    Write-Host "[ERROR] Search API failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host "`n[SUCCESS] All scope removal API tests passed!" -ForegroundColor Cyan
Write-Host "[INFO] The key scope removal functionality is working correctly:" -ForegroundColor Yellow
Write-Host "  ✅ Scoped package endpoints return 404" -ForegroundColor Green
Write-Host "  ✅ Search API excludes scope fields" -ForegroundColor Green
Write-Host "  ✅ API health and basic operations work" -ForegroundColor Green