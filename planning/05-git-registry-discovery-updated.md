# Phase 5: Git Registry Search and Discovery (Updated)

## Overview
Implement package discovery operations for Git-based registries, including search, package information retrieval, and archive downloading. This phase focuses on reading and parsing the registry structure with consistent error handling and caching strategies.

## Changes from Original Planning
- Updated to use `os` package instead of deprecated `ioutil`
- Clarified that Git-specific types can be added to existing `git.go` or separate file
- Added missing imports for sha256 and hex encoding
- Simplified `copyFile` implementation
- Confirmed all helper methods already exist in codebase

## Scope
- Implement optimized registry index parsing with fallback
- Add efficient package search functionality
- Implement package and version information retrieval with proper error types
- Add robust archive download capability with verification
- Create comprehensive registry structure validation
- Ensure consistency with HTTP client patterns

## Prerequisites
- Phase 4: Git Client Basic Operations completed ✅
- Understanding of registry structure patterns ✅
- Error handling types established ✅
- Helper methods already implemented ✅

## Expected Repository Structure

```
repo-root/
├── index.json              # Registry metadata and package listing
├── packages/
│   ├── package-name/
│   │   ├── metadata.json   # Package metadata
│   │   ├── versions/
│   │   │   ├── v1.0.0/
│   │   │   │   ├── manifest.json
│   │   │   │   └── archive.tar.gz
│   │   │   └── v1.1.0/
│   │   │       ├── manifest.json
│   │   │       └── archive.tar.gz
│   │   └── latest.json     # Points to latest version
│   └── another-package/
└── README.md               # Optional registry documentation
```

## Implementation Plan

### Step 1: Add Git-Specific Types
Location: Either `internal/client/git.go` or new file `internal/client/git_types.go`

### Step 2: Implement Index Loading
- Load and parse `index.json`
- Fallback to rebuilding from packages directory
- Cache index for performance

### Step 3: Implement Search Functionality
- Use loaded index for efficient searching
- Support query, tag, and target filters
- Apply limit parameter

### Step 4: Implement Package Retrieval
- `GetPackage`: Load package metadata
- `GetPackageVersion`: Load version manifest
- Use existing helper methods

### Step 5: Implement Archive Download
- Find archive by SHA256 hash
- Copy to destination path
- Verify integrity

### Step 6: Add Validation Helpers
- Validate registry structure
- Rebuild index when missing
- Handle corrupt/missing files gracefully

## Success Criteria
- Can search and filter packages from Git registry
- Can retrieve package and version information
- Can download archives by SHA256 hash
- Handles missing or corrupt index gracefully
- Validates repository structure correctly
- Efficient searching with index caching

## Testing Requirements
- Unit tests for index parsing and search logic
- Integration tests with test Git repository
- Test index rebuild functionality
- Test error handling for missing/corrupt files
- Verify performance with large package lists

## Next Phase
Phase 6: Git Registry Publishing - Implement package publishing via Git commits and pull requests