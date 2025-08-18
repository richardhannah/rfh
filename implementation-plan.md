# Implementation Plan: `rfh add` Command

## Questions Status
- [x] Question 1: What should the add command do exactly?
- [x] Question 2: Where should downloaded packages be stored?
- [x] Question 3: What should happen if a package is already installed?
- [x] Question 4: Integration with the API
- [x] Question 5: Expected behavior

## Implementation Tasks

### Phase 1: Core Infrastructure
- [ ] **Project root detection** - Find project root by looking for `rulestack.json`
- [ ] **Package name/version parsing** - Parse `package@version` format, handle defaults
- [ ] **API client methods** - Add methods to client.go for:
  - Get package metadata (via `/v1/packages/{scope}/{name}`)
  - Download package archive (via `/v1/blobs/{sha256}`)
  - No authentication required
- [ ] **Package extraction logic** - Extract .tgz files to `.rulestack/{packagename}/`
- [ ] **Directory management** - Create `.rulestack/` directory structure in project root
- [ ] **Dual manifest system** - Create/update both:
  - `rulestack.json` (user-editable dependencies)
  - `rulestack.lock.json` (actual installed packages)
- [ ] **Manifest structure** - Include project root reference and CLI version (1.0.0)

### Phase 2: Add Command Implementation  
- [ ] **CLI interface** - Implement the add command in `add.go`
- [ ] **Version validation** - Require `package@version` format, reject version-less requests
- [ ] **Conflict detection** - Check if package already installed in `.rulestack/`
- [ ] **User confirmation** - Prompt user when package exists, allow override
- [ ] **Error handling** - Handle network errors, file conflicts, missing packages
- [ ] **Verbose output** - Add progress indicators and detailed logging

### Phase 3: Testing & Integration
- [ ] **Unit tests** - Test parsing, extraction, error cases
- [ ] **Integration test** - Update test scripts to verify add functionality
- [ ] **Documentation** - Update help text and examples

## Dependencies to Analyze
- `internal/client/client.go` - HTTP client for API calls
- `internal/pkg/archive.go` - Archive extraction utilities  
- `internal/config/cli.go` - Registry configuration
- API endpoints: `/v1/packages/{scope}/{name}` and `/v1/blobs/{sha256}`

## Questions Impact on Implementation

### Question 1 Answer: ✅ ANSWERED
**Requirements:**
- Download and unpack specified package from registry
- Create `.rulestack/` directory if not exists
- Structure: `.rulestack/{packagename}/rules.md`
- Create/update package manifest in `.rulestack/` root
- Manifest tracks installed rule packages

### Question 2 Answer: ✅ ANSWERED
**Requirements:**
- `.rulestack/` folder in project root (find root by looking for existing `rulestack.json`)
- If run from subfolder, still update root `.rulestack/` folder
- Package structure: `.rulestack/{packagename}/` (multiple files in future, single `rules.md` for now)
- Two manifest files:
  - `rulestack.json` - User-editable, specifies desired packages
  - `rulestack.lock.json` - Auto-generated, tracks actually installed packages
- Manifests include:
  - Project root reference
  - CLI version field (start with 1.0.0)

### Question 3 Answer: ✅ ANSWERED
**Requirements:**
- If package already installed, prompt user for confirmation to reinstall
- Require specific versions for now (no automatic version resolution)
- User must specify exact version with `package@version` format
- Future: Will add version resolution (latest, semver ranges, etc.)

### Question 4 Answer: ✅ ANSWERED
**Requirements:**
- Use existing `getPackageHandler` and `downloadBlobHandler` endpoints
- API endpoints: `/v1/packages/{scope}/{name}` and `/v1/blobs/{sha256}`
- No authentication required for package downloads (proof of concept)
- No authentication needed anywhere in current setup
- Keep API contract consistent, internal implementation flexible

### Question 5 Answer: ✅ ANSWERED
**Requirements:**
- Very basic functionality: `rfh add {packagename@version}`
- No dependency resolution/installation for now
- Single package installation only
- Future: May add dependency support later
- Keep implementation simple and focused

---

*This plan will be updated as questions are answered and requirements clarified.*