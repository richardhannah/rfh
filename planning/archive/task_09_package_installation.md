# Task 10: Package Download and Installation (1 hour)

## Objective
Implement the `add` and `apply` commands for downloading and installing rulesets into editor-specific locations.

## Prerequisites
- Tasks 1-9 completed
- HTTP client working
- Search and publish functionality operational

## Checklist

### 1. Create Add Command (25 minutes)
Create `internal/cli/add.go`:
```go
package cli

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    
    "github.com/spf13/cobra"
    
    "rulestack/internal/client"
    "rulestack/internal/config"
    "rulestack/internal/pkg"
)

var (
    addTarget string
    addForce  bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
    Use:   "add <package>[@version]",
    Short: "Add (download) a ruleset package",
    Long: `Download and add a ruleset package to the current workspace.

This command will:
1. Resolve the package version (latest if not specified)
2. Download the package archive
3. Extract it to a cache location
4. Update the lockfile with package information

The package will be available for application to editors with the 'apply' command.

Examples:
  rfh add @acme/secure-coding
  rfh add @acme/secure-coding@1.2.0
  rfh add secure-linting --target=cursor`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return runAdd(args[0])
    },
}

// PackageSpec represents a parsed package specification
type PackageSpec struct {
    Scope   string
    Name    string
    Version string
    FullName string
}

// parsePackageSpec parses a package spec like "@scope/name@version" or "name@version"
func parsePackageSpec(spec string) (*PackageSpec, error) {
    // Regex to match: (optional @scope/)(name)(@optional version)
    re := regexp.MustCompile(`^(?:@([^/]+)/)?([^@]+)(?:@(.+))?$`)
    matches := re.FindStringSubmatch(spec)
    
    if len(matches) != 4 {
        return nil, fmt.Errorf("invalid package specification: %s", spec)
    }
    
    scope := matches[1]
    name := matches[2]
    version := matches[3]
    
    // Build full name
    fullName := name
    if scope != "" {
        fullName = "@" + scope + "/" + name
    }
    
    return &PackageSpec{
        Scope:    scope,
        Name:     name,
        Version:  version,
        FullName: fullName,
    }, nil
}

func runAdd(packageSpec string) error {
    // Parse package specification
    spec, err := parsePackageSpec(packageSpec)
    if err != nil {
        return err
    }
    
    if verbose {
        fmt.Printf("📦 Adding %s\n", spec.FullName)
        if spec.Version != "" {
            fmt.Printf("📌 Version: %s\n", spec.Version)
        } else {
            fmt.Printf("📌 Version: latest\n")
        }
    }
    
    // Get registry configuration
    cfg, err := config.LoadCLI()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Determine which registry to use
    registryName := cfg.Current
    if registry != "" {
        registryName = registry
    }
    
    if registryName == "" {
        return fmt.Errorf("no registry configured. Use 'rfh registry add' to add a registry")
    }
    
    reg, exists := cfg.Registries[registryName]
    if !exists {
        return fmt.Errorf("registry '%s' not found", registryName)
    }
    
    // Create client
    c := client.NewClient(reg.URL, reg.Token)
    c.SetVerbose(verbose)
    
    // Get package information
    fmt.Printf("🔍 Resolving %s...\n", spec.FullName)
    packageInfo, err := c.GetPackage(spec.Scope, spec.Name)
    if err != nil {
        return fmt.Errorf("failed to get package info: %w", err)
    }
    
    // TODO: Implement version resolution logic
    // For now, assume we get the right version info from the API
    resolvedVersion := spec.Version
    if resolvedVersion == "" {
        resolvedVersion = "latest" // This should be resolved from API
    }
    
    // Check if already installed
    lockfile, err := pkg.LoadLockfile("rfh.lock")
    if err != nil {
        return fmt.Errorf("failed to load lockfile: %w", err)
    }
    
    lockKey := fmt.Sprintf("%s@%s", spec.FullName, resolvedVersion)
    if lockfile.HasPackage(lockKey) && !addForce {
        fmt.Printf("✅ %s already installed\n", lockKey)
        fmt.Printf("💡 Use --force to reinstall\n")
        return nil
    }
    
    // Create cache directory
    cacheDir, err := getCacheDir()
    if err != nil {
        return fmt.Errorf("failed to get cache directory: %w", err)
    }
    
    packageCacheDir := filepath.Join(cacheDir, spec.FullName, resolvedVersion)
    if err := os.MkdirAll(packageCacheDir, 0o755); err != nil {
        return fmt.Errorf("failed to create cache directory: %w", err)
    }
    
    // Download archive (assuming we have SHA256 from package info)
    sha256 := packageInfo["sha256"].(string) // This should come from API response
    archivePath := filepath.Join(packageCacheDir, "archive.tgz")
    
    fmt.Printf("📥 Downloading %s@%s...\n", spec.FullName, resolvedVersion)
    if err := c.DownloadBlob(sha256, archivePath); err != nil {
        return fmt.Errorf("failed to download package: %w", err)
    }
    
    // Verify SHA256
    actualSHA256, err := pkg.CalculateSHA256(archivePath)\n    if err != nil {\n        return fmt.Errorf(\"failed to calculate SHA256: %w\", err)\n    }\n    \n    if actualSHA256 != sha256 {\n        return fmt.Errorf(\"SHA256 mismatch: expected %s, got %s\", sha256, actualSHA256)\n    }\n    \n    // Extract archive\n    extractDir := filepath.Join(packageCacheDir, \"extracted\")\n    if err := pkg.Unpack(archivePath, extractDir); err != nil {\n        return fmt.Errorf(\"failed to extract package: %w\", err)\n    }\n    \n    // Update lockfile\n    entry := pkg.LockfileEntry{\n        Version:     resolvedVersion,\n        SHA256:      sha256,\n        Targets:     []string{addTarget}, // This should come from manifest\n        InstallPath: extractDir,\n        Registry:    registryName,\n    }\n    \n    lockfile.AddPackage(lockKey, entry)\n    lockfile.Registry = reg.URL\n    \n    if err := lockfile.SaveLockfile(\"rfh.lock\"); err != nil {\n        return fmt.Errorf(\"failed to save lockfile: %w\", err)\n    }\n    \n    fmt.Printf(\"✅ Successfully added %s@%s\n\", spec.FullName, resolvedVersion)\n    fmt.Printf(\"📁 Cached at: %s\n\", extractDir)\n    fmt.Printf(\"💡 Apply with: rfh apply %s --target <editor>\n\", spec.FullName)\n    \n    return nil\n}\n\n// getCacheDir returns the cache directory path\nfunc getCacheDir() (string, error) {\n    home, err := os.UserHomeDir()\n    if err != nil {\n        return \"\", err\n    }\n    \n    return filepath.Join(home, \".rfh\", \"cache\"), nil\n}\n\nfunc init() {\n    addCmd.Flags().StringVar(&addTarget, \"target\", \"\", \"target editor (cursor, claude-code, etc.)\")\n    addCmd.Flags().BoolVar(&addForce, \"force\", false, \"force reinstall if already installed\")\n}\n```\n\n- [ ] Create add.go file\n- [ ] Test package spec parsing\n- [ ] Verify cache directory creation\n\n### 2. Create Apply Command (25 minutes)\nCreate `internal/cli/apply.go`:\n```go\npackage cli\n\nimport (\n    \"fmt\"\n    \"io\"\n    \"os\"\n    \"path/filepath\"\n    \"strings\"\n    \n    \"github.com/spf13/cobra\"\n    \n    \"rulestack/internal/pkg\"\n)\n\nvar (\n    applyTarget    string\n    applyWorkspace string\n)\n\n// applyCmd represents the apply command\nvar applyCmd = &cobra.Command{\n    Use:   \"apply <package>[@version]\",\n    Short: \"Apply a ruleset to an editor workspace\",\n    Long: `Apply a previously added ruleset to a specific editor workspace.\n\nThis command will:\n1. Locate the cached package files\n2. Copy rule files to the appropriate editor-specific directory\n3. Handle any editor-specific transformations\n\nSupported targets:\n- cursor: Copies to .cursor/rules/\n- claude-code: Copies to .claude/rules/\n- windsurf: Copies to .windsurf/rules/\n\nExamples:\n  rfh apply @acme/secure-coding --target cursor\n  rfh apply @acme/secure-coding@1.2.0 --target claude-code --workspace ./my-project`,\n    Args: cobra.ExactArgs(1),\n    RunE: func(cmd *cobra.Command, args []string) error {\n        return runApply(args[0])\n    },\n}\n\nfunc runApply(packageSpec string) error {\n    // Parse package specification\n    spec, err := parsePackageSpec(packageSpec)\n    if err != nil {\n        return err\n    }\n    \n    if applyTarget == \"\" {\n        return fmt.Errorf(\"target editor is required. Use --target flag\")\n    }\n    \n    // Validate target\n    validTargets := map[string]string{\n        \"cursor\":     \".cursor/rules\",\n        \"claude-code\": \".claude/rules\",\n        \"windsurf\":   \".windsurf/rules\",\n        \"copilot\":    \".github/copilot\",\n    }\n    \n    targetDir, isValid := validTargets[applyTarget]\n    if !isValid {\n        return fmt.Errorf(\"unsupported target: %s. Supported: %v\", applyTarget, getKeys(validTargets))\n    }\n    \n    if verbose {\n        fmt.Printf(\"🎯 Applying %s to %s\n\", spec.FullName, applyTarget)\n        fmt.Printf(\"📁 Target directory: %s\n\", targetDir)\n    }\n    \n    // Load lockfile to find installed package\n    lockfile, err := pkg.LoadLockfile(\"rfh.lock\")\n    if err != nil {\n        return fmt.Errorf(\"failed to load lockfile: %w\", err)\n    }\n    \n    // Find package in lockfile\n    var packageEntry pkg.LockfileEntry\n    var found bool\n    \n    // Try with specified version first\n    if spec.Version != \"\" {\n        lockKey := fmt.Sprintf(\"%s@%s\", spec.FullName, spec.Version)\n        packageEntry, found = lockfile.GetPackage(lockKey)\n    }\n    \n    // If not found, try to find any version\n    if !found {\n        for key, entry := range lockfile.Packages {\n            if strings.HasPrefix(key, spec.FullName+\"@\") {\n                packageEntry = entry\n                found = true\n                break\n            }\n        }\n    }\n    \n    if !found {\n        return fmt.Errorf(\"package %s not installed. Use 'rfh add %s' first\", spec.FullName, spec.FullName)\n    }\n    \n    // Get workspace directory\n    workspace := applyWorkspace\n    if workspace == \"\" {\n        workspace = \".\"\n    }\n    \n    // Create target directory\n    fullTargetDir := filepath.Join(workspace, targetDir, spec.FullName)\n    if err := os.MkdirAll(fullTargetDir, 0o755); err != nil {\n        return fmt.Errorf(\"failed to create target directory: %w\", err)\n    }\n    \n    // Copy files from cache to target\n    sourceDir := packageEntry.InstallPath\n    if sourceDir == \"\" {\n        return fmt.Errorf(\"no install path found for package %s\", spec.FullName)\n    }\n    \n    fmt.Printf(\"📋 Copying rules to %s...\n\", fullTargetDir)\n    \n    filesCopied, err := copyRuleFiles(sourceDir, fullTargetDir, applyTarget)\n    if err != nil {\n        return fmt.Errorf(\"failed to copy rule files: %w\", err)\n    }\n    \n    fmt.Printf(\"✅ Applied %s to %s\n\", spec.FullName, applyTarget)\n    fmt.Printf(\"📁 Location: %s\n\", fullTargetDir)\n    fmt.Printf(\"📄 Files copied: %d\n\", filesCopied)\n    \n    // Show next steps based on target\n    showTargetSpecificInstructions(applyTarget, fullTargetDir)\n    \n    return nil\n}\n\n// copyRuleFiles copies rule files from source to destination with target-specific handling\nfunc copyRuleFiles(sourceDir, targetDir, target string) (int, error) {\n    filesCopied := 0\n    \n    err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {\n        if err != nil {\n            return err\n        }\n        \n        // Skip directories\n        if info.IsDir() {\n            return nil\n        }\n        \n        // Get relative path\n        relPath, err := filepath.Rel(sourceDir, path)\n        if err != nil {\n            return err\n        }\n        \n        // Skip non-rule files (keep manifest, docs, etc. but focus on rules)\n        if !isRuleFile(relPath) {\n            return nil\n        }\n        \n        // Create destination path\n        destPath := filepath.Join(targetDir, relPath)\n        \n        // Create directory if needed\n        if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {\n            return err\n        }\n        \n        // Copy file\n        if err := copyFile(path, destPath); err != nil {\n            return fmt.Errorf(\"failed to copy %s: %w\", relPath, err)\n        }\n        \n        filesCopied++\n        \n        if verbose {\n            fmt.Printf(\"   📄 %s\n\", relPath)\n        }\n        \n        return nil\n    })\n    \n    return filesCopied, err\n}\n\n// isRuleFile checks if a file should be considered a rule file\nfunc isRuleFile(path string) bool {\n    // Common rule file extensions\n    ext := strings.ToLower(filepath.Ext(path))\n    switch ext {\n    case \".md\", \".txt\", \".yaml\", \".yml\", \".json\":\n        return true\n    }\n    \n    // Check if it's in a rules directory\n    return strings.Contains(strings.ToLower(path), \"rule\")\n}\n\n// copyFile copies a single file\nfunc copyFile(src, dst string) error {\n    source, err := os.Open(src)\n    if err != nil {\n        return err\n    }\n    defer source.Close()\n    \n    destination, err := os.Create(dst)\n    if err != nil {\n        return err\n    }\n    defer destination.Close()\n    \n    _, err = io.Copy(destination, source)\n    return err\n}\n\n// showTargetSpecificInstructions shows next steps for each editor\nfunc showTargetSpecificInstructions(target, targetDir string) {\n    switch target {\n    case \"cursor\":\n        fmt.Printf(\"\\n💡 Next steps for Cursor:\n\")\n        fmt.Printf(\"   1. Restart Cursor to load new rules\n\")\n        fmt.Printf(\"   2. Rules should appear in Cursor's AI context\n\")\n    case \"claude-code\":\n        fmt.Printf(\"\\n💡 Next steps for Claude Code:\n\")\n        fmt.Printf(\"   1. Rules are now available in your workspace\n\")\n        fmt.Printf(\"   2. Claude will use these rules automatically\n\")\n    case \"windsurf\":\n        fmt.Printf(\"\\n💡 Next steps for Windsurf:\n\")\n        fmt.Printf(\"   1. Restart Windsurf to load new rules\n\")\n        fmt.Printf(\"   2. Check Windsurf settings for rule integration\n\")\n    }\n}\n\n// getKeys returns keys from a string map\nfunc getKeys(m map[string]string) []string {\n    keys := make([]string, 0, len(m))\n    for k := range m {\n        keys = append(keys, k)\n    }\n    return keys\n}\n\nfunc init() {\n    applyCmd.Flags().StringVar(&applyTarget, \"target\", \"\", \"target editor (cursor, claude-code, windsurf, copilot)\")\n    applyCmd.Flags().StringVar(&applyWorkspace, \"workspace\", \"\", \"workspace directory (default: current directory)\")\n    \n    // Mark target as required\n    applyCmd.MarkFlagRequired(\"target\")\n}\n```\n\n- [ ] Create apply.go file\n- [ ] Test file copying logic\n- [ ] Verify target-specific directory creation\n\n### 3. Complete Package Publishing API Handler (10 minutes)\nUpdate `internal/api/handlers.go` to add the missing publishPackageHandler:\n```go\n// Add this function to internal/api/handlers.go\n\n// publishPackageHandler handles package publishing\nfunc (s *Server) publishPackageHandler(w http.ResponseWriter, r *http.Request) {\n    // Get token from context (set by auth middleware)\n    token := getTokenFromContext(r.Context())\n    if token == nil {\n        writeError(w, http.StatusUnauthorized, \"Authentication required\")\n        return\n    }\n    \n    // Parse multipart form\n    if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit\n        writeError(w, http.StatusBadRequest, \"Failed to parse form\")\n        return\n    }\n    \n    // Get manifest file\n    manifestFile, _, err := r.FormFile(\"manifest\")\n    if err != nil {\n        writeError(w, http.StatusBadRequest, \"Manifest file required\")\n        return\n    }\n    defer manifestFile.Close()\n    \n    // Parse manifest\n    var manifest struct {\n        Name        string   `json:\"name\"`\n        Version     string   `json:\"version\"`\n        Description string   `json:\"description\"`\n        Targets     []string `json:\"targets\"`\n        Tags        []string `json:\"tags\"`\n    }\n    \n    if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {\n        writeError(w, http.StatusBadRequest, \"Invalid manifest JSON\")\n        return\n    }\n    \n    // Get archive file\n    archiveFile, _, err := r.FormFile(\"archive\")\n    if err != nil {\n        writeError(w, http.StatusBadRequest, \"Archive file required\")\n        return\n    }\n    defer archiveFile.Close()\n    \n    // Calculate SHA256 and save archive\n    hasher := sha256.New()\n    archivePath := filepath.Join(s.Config.StoragePath, fmt.Sprintf(\"%s-%s.tgz\", manifest.Name, manifest.Version))\n    \n    outFile, err := os.Create(archivePath)\n    if err != nil {\n        writeError(w, http.StatusInternalServerError, \"Failed to save archive\")\n        return\n    }\n    defer outFile.Close()\n    \n    // Copy with hashing\n    teeReader := io.TeeReader(archiveFile, hasher)\n    size, err := io.Copy(outFile, teeReader)\n    if err != nil {\n        writeError(w, http.StatusInternalServerError, \"Failed to save archive\")\n        return\n    }\n    \n    sha256Hash := fmt.Sprintf(\"%x\", hasher.Sum(nil))\n    \n    // Parse package name and scope\n    var scope *string\n    packageName := manifest.Name\n    \n    if strings.HasPrefix(manifest.Name, \"@\") {\n        parts := strings.SplitN(manifest.Name[1:], \"/\", 2)\n        if len(parts) == 2 {\n            s := parts[0]\n            scope = &s\n            packageName = parts[1]\n        }\n    }\n    \n    // Create or get package\n    pkg, err := s.DB.GetOrCreatePackage(scope, packageName)\n    if err != nil {\n        writeError(w, http.StatusInternalServerError, \"Failed to create package\")\n        return\n    }\n    \n    // Create package version\n    version := db.PackageVersion{\n        PackageID:   pkg.ID,\n        Version:     manifest.Version,\n        Description: &manifest.Description,\n        Targets:     manifest.Targets,\n        Tags:        manifest.Tags,\n        SHA256:      &sha256Hash,\n        SizeBytes:   &[]int{int(size)}[0],\n        BlobPath:    &archivePath,\n    }\n    \n    createdVersion, err := s.DB.CreatePackageVersion(version)\n    if err != nil {\n        writeError(w, http.StatusConflict, \"Package version already exists or creation failed\")\n        return\n    }\n    \n    // Return success response\n    writeJSON(w, http.StatusCreated, map[string]interface{}{\n        \"name\":    manifest.Name,\n        \"version\": manifest.Version,\n        \"sha256\":  sha256Hash,\n        \"size\":    size,\n        \"id\":      createdVersion.ID,\n    })\n}\n\n// Add this function to internal/api/handlers.go\n\n// downloadBlobHandler handles blob downloads\nfunc (s *Server) downloadBlobHandler(w http.ResponseWriter, r *http.Request) {\n    vars := mux.Vars(r)\n    sha256 := vars[\"sha256\"]\n    \n    if sha256 == \"\" {\n        writeError(w, http.StatusBadRequest, \"SHA256 required\")\n        return\n    }\n    \n    // Find package version by SHA256\n    var blobPath string\n    err := s.DB.Get(&blobPath, \"SELECT blob_path FROM package_versions WHERE sha256 = $1\", sha256)\n    if err != nil {\n        writeError(w, http.StatusNotFound, \"Blob not found\")\n        return\n    }\n    \n    // Open file\n    file, err := os.Open(blobPath)\n    if err != nil {\n        writeError(w, http.StatusInternalServerError, \"Failed to read blob\")\n        return\n    }\n    defer file.Close()\n    \n    // Get file info\n    info, err := file.Stat()\n    if err != nil {\n        writeError(w, http.StatusInternalServerError, \"Failed to get file info\")\n        return\n    }\n    \n    // Set headers\n    w.Header().Set(\"Content-Type\", \"application/gzip\")\n    w.Header().Set(\"Content-Length\", fmt.Sprintf(\"%d\", info.Size()))\n    w.Header().Set(\"Content-Disposition\", fmt.Sprintf(\"attachment; filename=\\\"%s.tgz\\\"\", sha256[:8]))\n    \n    // Stream file\n    http.ServeContent(w, r, \"\", info.ModTime(), file)\n}\n```\n\n- [ ] Add publishPackageHandler to handlers.go\n- [ ] Add downloadBlobHandler to handlers.go\n- [ ] Test file upload and download\n\n## Validation\nTest the complete package flow:\n```bash\n# Build everything\ngo build -o rfh ./cmd/cli\ngo build -o rulestack-api ./cmd/api\n\n# Start API server\n./rulestack-api\n\n# In another terminal:\n# Create and publish a package\n./rfh init\n./rfh pack\n./rfh registry add local http://localhost:8080 test-token\n./rfh publish\n\n# Search and install\n./rfh search example\n./rfh add @acme/example-rules\n./rfh apply @acme/example-rules --target cursor\n./rfh list\n```\n\n## Acceptance Criteria\n- [ ] Add command downloads and caches packages correctly\n- [ ] SHA256 verification works during download\n- [ ] Apply command copies files to correct editor directories\n- [ ] Lockfile is properly maintained\n- [ ] Package spec parsing handles scoped/unscoped packages\n- [ ] Target validation prevents invalid editor names\n- [ ] File copying preserves directory structure\n- [ ] Error messages guide users on next steps\n- [ ] Cache directory is created in user's home\n- [ ] API handlers support publish and download flows\n\n## Time Estimate: ~60 minutes\n\n## Next Task\nTask 10: End-to-End Testing and Polish"
        ]
    }

### Create Apply Command (25 minutes)
Create `internal/cli/apply.go`:
```go
package cli

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/spf13/cobra"
    
    "rulestack/internal/pkg"
)

var (
    applyTarget    string
    applyWorkspace string
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
    Use:   "apply <package>[@version]",
    Short: "Apply a ruleset to an editor workspace",
    Long: `Apply a previously added ruleset to a specific editor workspace.

This command will:
1. Locate the cached package files
2. Copy rule files to the appropriate editor-specific directory
3. Handle any editor-specific transformations

Supported targets:
- cursor: Copies to .cursor/rules/
- claude-code: Copies to .claude/rules/
- windsurf: Copies to .windsurf/rules/

Examples:
  rfh apply @acme/secure-coding --target cursor
  rfh apply @acme/secure-coding@1.2.0 --target claude-code --workspace ./my-project`,
    Args: cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        return runApply(args[0])
    },
}

func runApply(packageSpec string) error {
    // Parse package specification
    spec, err := parsePackageSpec(packageSpec)
    if err != nil {
        return err
    }
    
    if applyTarget == "" {
        return fmt.Errorf("target editor is required. Use --target flag")
    }
    
    // Validate target
    validTargets := map[string]string{
        "cursor":     ".cursor/rules",
        "claude-code": ".claude/rules",
        "windsurf":   ".windsurf/rules",
        "copilot":    ".github/copilot",
    }
    
    targetDir, isValid := validTargets[applyTarget]
    if !isValid {
        return fmt.Errorf("unsupported target: %s. Supported: %v", applyTarget, getKeys(validTargets))
    }
    
    if verbose {
        fmt.Printf("🎯 Applying %s to %s\n", spec.FullName, applyTarget)
        fmt.Printf("📁 Target directory: %s\n", targetDir)
    }
    
    // Load lockfile to find installed package
    lockfile, err := pkg.LoadLockfile("rfh.lock")
    if err != nil {
        return fmt.Errorf("failed to load lockfile: %w", err)
    }
    
    // Find package in lockfile
    var packageEntry pkg.LockfileEntry
    var found bool
    
    // Try with specified version first
    if spec.Version != "" {
        lockKey := fmt.Sprintf("%s@%s", spec.FullName, spec.Version)
        packageEntry, found = lockfile.GetPackage(lockKey)
    }
    
    // If not found, try to find any version
    if !found {
        for key, entry := range lockfile.Packages {
            if strings.HasPrefix(key, spec.FullName+"@") {
                packageEntry = entry
                found = true
                break
            }
        }
    }
    
    if !found {
        return fmt.Errorf("package %s not installed. Use 'rfh add %s' first", spec.FullName, spec.FullName)
    }
    
    // Get workspace directory
    workspace := applyWorkspace
    if workspace == "" {
        workspace = "."
    }
    
    // Create target directory
    fullTargetDir := filepath.Join(workspace, targetDir, spec.FullName)
    if err := os.MkdirAll(fullTargetDir, 0o755); err != nil {
        return fmt.Errorf("failed to create target directory: %w", err)
    }
    
    // Copy files from cache to target
    sourceDir := packageEntry.InstallPath
    if sourceDir == "" {
        return fmt.Errorf("no install path found for package %s", spec.FullName)
    }
    
    fmt.Printf("📋 Copying rules to %s...\n", fullTargetDir)
    
    filesCopied, err := copyRuleFiles(sourceDir, fullTargetDir, applyTarget)
    if err != nil {
        return fmt.Errorf("failed to copy rule files: %w", err)
    }
    
    fmt.Printf("✅ Applied %s to %s\n", spec.FullName, applyTarget)
    fmt.Printf("📁 Location: %s\n", fullTargetDir)
    fmt.Printf("📄 Files copied: %d\n", filesCopied)
    
    // Show next steps based on target
    showTargetSpecificInstructions(applyTarget, fullTargetDir)
    
    return nil
}

// copyRuleFiles copies rule files from source to destination with target-specific handling
func copyRuleFiles(sourceDir, targetDir, target string) (int, error) {
    filesCopied := 0
    
    err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // Skip directories
        if info.IsDir() {
            return nil
        }
        
        // Get relative path
        relPath, err := filepath.Rel(sourceDir, path)
        if err != nil {
            return err
        }
        
        // Skip non-rule files (keep manifest, docs, etc. but focus on rules)
        if !isRuleFile(relPath) {
            return nil
        }
        
        // Create destination path
        destPath := filepath.Join(targetDir, relPath)
        
        // Create directory if needed
        if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
            return err
        }
        
        // Copy file
        if err := copyFile(path, destPath); err != nil {
            return fmt.Errorf("failed to copy %s: %w", relPath, err)
        }
        
        filesCopied++
        
        if verbose {
            fmt.Printf("   📄 %s\n", relPath)
        }
        
        return nil
    })
    
    return filesCopied, err
}

// isRuleFile checks if a file should be considered a rule file
func isRuleFile(path string) bool {
    // Common rule file extensions
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".md", ".txt", ".yaml", ".yml", ".json":
        return true
    }
    
    // Check if it's in a rules directory
    return strings.Contains(strings.ToLower(path), "rule")
}

// copyFile copies a single file
func copyFile(src, dst string) error {
    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()
    
    destination, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destination.Close()
    
    _, err = io.Copy(destination, source)
    return err
}

// showTargetSpecificInstructions shows next steps for each editor
func showTargetSpecificInstructions(target, targetDir string) {
    switch target {
    case "cursor":
        fmt.Printf("\n💡 Next steps for Cursor:\n")
        fmt.Printf("   1. Restart Cursor to load new rules\n")
        fmt.Printf("   2. Rules should appear in Cursor's AI context\n")
    case "claude-code":
        fmt.Printf("\n💡 Next steps for Claude Code:\n")
        fmt.Printf("   1. Rules are now available in your workspace\n")
        fmt.Printf("   2. Claude will use these rules automatically\n")
    case "windsurf":
        fmt.Printf("\n💡 Next steps for Windsurf:\n")
        fmt.Printf("   1. Restart Windsurf to load new rules\n")
        fmt.Printf("   2. Check Windsurf settings for rule integration\n")
    }
}

// getKeys returns keys from a string map
func getKeys(m map[string]string) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

func init() {
    applyCmd.Flags().StringVar(&applyTarget, "target", "", "target editor (cursor, claude-code, windsurf, copilot)")
    applyCmd.Flags().StringVar(&applyWorkspace, "workspace", "", "workspace directory (default: current directory)")
    
    // Mark target as required
    applyCmd.MarkFlagRequired("target")
}
```

- [ ] Create apply.go file
- [ ] Test file copying logic
- [ ] Verify target-specific directory creation

### 3. Complete Package Publishing API Handler (10 minutes)
Update `internal/api/handlers.go` to add the missing publishPackageHandler:
```go
// Add this function to internal/api/handlers.go

// publishPackageHandler handles package publishing
func (s *Server) publishPackageHandler(w http.ResponseWriter, r *http.Request) {
    // Get token from context (set by auth middleware)
    token := getTokenFromContext(r.Context())
    if token == nil {
        writeError(w, http.StatusUnauthorized, "Authentication required")
        return
    }
    
    // Parse multipart form
    if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
        writeError(w, http.StatusBadRequest, "Failed to parse form")
        return
    }
    
    // Get manifest file
    manifestFile, _, err := r.FormFile("manifest")
    if err != nil {
        writeError(w, http.StatusBadRequest, "Manifest file required")
        return
    }
    defer manifestFile.Close()
    
    // Parse manifest
    var manifest struct {
        Name        string   `json:"name"`
        Version     string   `json:"version"`
        Description string   `json:"description"`
        Targets     []string `json:"targets"`
        Tags        []string `json:"tags"`
    }
    
    if err := json.NewDecoder(manifestFile).Decode(&manifest); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid manifest JSON")
        return
    }
    
    // Get archive file
    archiveFile, _, err := r.FormFile("archive")
    if err != nil {
        writeError(w, http.StatusBadRequest, "Archive file required")
        return
    }
    defer archiveFile.Close()
    
    // Calculate SHA256 and save archive
    hasher := sha256.New()
    archivePath := filepath.Join(s.Config.StoragePath, fmt.Sprintf("%s-%s.tgz", manifest.Name, manifest.Version))
    
    outFile, err := os.Create(archivePath)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to save archive")
        return
    }
    defer outFile.Close()
    
    // Copy with hashing
    teeReader := io.TeeReader(archiveFile, hasher)
    size, err := io.Copy(outFile, teeReader)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to save archive")
        return
    }
    
    sha256Hash := fmt.Sprintf("%x", hasher.Sum(nil))
    
    // Parse package name and scope
    var scope *string
    packageName := manifest.Name
    
    if strings.HasPrefix(manifest.Name, "@") {
        parts := strings.SplitN(manifest.Name[1:], "/", 2)
        if len(parts) == 2 {
            s := parts[0]
            scope = &s
            packageName = parts[1]
        }
    }
    
    // Create or get package
    pkg, err := s.DB.GetOrCreatePackage(scope, packageName)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to create package")
        return
    }
    
    // Create package version
    version := db.PackageVersion{
        PackageID:   pkg.ID,
        Version:     manifest.Version,
        Description: &manifest.Description,
        Targets:     manifest.Targets,
        Tags:        manifest.Tags,
        SHA256:      &sha256Hash,
        SizeBytes:   &[]int{int(size)}[0],
        BlobPath:    &archivePath,
    }
    
    createdVersion, err := s.DB.CreatePackageVersion(version)
    if err != nil {
        writeError(w, http.StatusConflict, "Package version already exists or creation failed")
        return
    }
    
    // Return success response
    writeJSON(w, http.StatusCreated, map[string]interface{}{
        "name":    manifest.Name,
        "version": manifest.Version,
        "sha256":  sha256Hash,
        "size":    size,
        "id":      createdVersion.ID,
    })
}

// Add this function to internal/api/handlers.go

// downloadBlobHandler handles blob downloads
func (s *Server) downloadBlobHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    sha256 := vars["sha256"]
    
    if sha256 == "" {
        writeError(w, http.StatusBadRequest, "SHA256 required")
        return
    }
    
    // Find package version by SHA256
    var blobPath string
    err := s.DB.Get(&blobPath, "SELECT blob_path FROM package_versions WHERE sha256 = $1", sha256)
    if err != nil {
        writeError(w, http.StatusNotFound, "Blob not found")
        return
    }
    
    // Open file
    file, err := os.Open(blobPath)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to read blob")
        return
    }
    defer file.Close()
    
    // Get file info
    info, err := file.Stat()
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to get file info")
        return
    }
    
    // Set headers
    w.Header().Set("Content-Type", "application/gzip")
    w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.tgz\"", sha256[:8]))
    
    // Stream file
    http.ServeContent(w, r, "", info.ModTime(), file)
}
```

- [ ] Add publishPackageHandler to handlers.go
- [ ] Add downloadBlobHandler to handlers.go
- [ ] Test file upload and download

## Validation
Test the complete package flow:
```bash
# Build everything
go build -o rfh ./cmd/cli
go build -o rulestack-api ./cmd/api

# Start API server
./rulestack-api

# In another terminal:
# Create and publish a package
./rfh init
./rfh pack
./rfh registry add local http://localhost:8080 test-token
./rfh publish

# Search and install
./rfh search example
./rfh add @acme/example-rules
./rfh apply @acme/example-rules --target cursor
./rfh list
```

## Acceptance Criteria
- [ ] Add command downloads and caches packages correctly
- [ ] SHA256 verification works during download
- [ ] Apply command copies files to correct editor directories
- [ ] Lockfile is properly maintained
- [ ] Package spec parsing handles scoped/unscoped packages
- [ ] Target validation prevents invalid editor names
- [ ] File copying preserves directory structure
- [ ] Error messages guide users on next steps
- [ ] Cache directory is created in user's home
- [ ] API handlers support publish and download flows

## Time Estimate: ~60 minutes

## Next Task
Task 11: End-to-End Testing and Polish
