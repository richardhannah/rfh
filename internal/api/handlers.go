package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"rulestack/internal/db"
)

// healthHandler returns API health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.DB.Health(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "Database connection failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "rulestack-api",
		"version": "1.0.0",
	})
}

// searchPackagesHandler searches for packages
func (s *Server) searchPackagesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	tag := r.URL.Query().Get("tag")
	target := r.URL.Query().Get("target")

	// Parse limit parameter
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	results, err := s.DB.SearchPackages(query, tag, target, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Search failed")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// getPackageHandler gets package information
func (s *Server) getPackageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	pkg, err := s.DB.GetPackage(name)
	if err != nil {
		writeError(w, http.StatusNotFound, "Package not found")
		return
	}

	writeJSON(w, http.StatusOK, pkg)
}

// getPackageVersionHandler gets specific package version
func (s *Server) getPackageVersionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	version := vars["version"]

	fmt.Printf("[DEBUG] getPackageVersionHandler called with name='%s', version='%s'\n", name, version)

	pkgVersion, err := s.DB.GetPackageVersion(name, version)
	if err != nil {
		fmt.Printf("[ERROR] GetPackageVersion failed: %v\n", err)
		writeError(w, http.StatusNotFound, "Package version not found")
		return
	}

	fmt.Printf("[DEBUG] Found package version: %+v\n", pkgVersion)
	writeJSON(w, http.StatusOK, pkgVersion)
}

// publishPackageHandler handles package publishing
func (s *Server) publishPackageHandler(w http.ResponseWriter, r *http.Request) {
	// Authentication is now handled by middleware based on route metadata
	// Check if user is authenticated (works for both JWT and legacy tokens)
	user := getUserFromContext(r.Context())
	if user == nil {
		fmt.Fprintf(os.Stderr, "DEBUG PUBLISH: No user found in context\n")
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	fmt.Fprintf(os.Stderr, "DEBUG PUBLISH: User authenticated: %s (ID: %d, Role: %s)\n", user.Username, user.ID, user.Role)

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
	// Sanitize filename by replacing invalid characters
	safeName := strings.ReplaceAll(manifest.Name, "/", "-")
	safeName = strings.ReplaceAll(safeName, "@", "")
	archivePath := filepath.Join(s.Config.StoragePath, fmt.Sprintf("%s-%s.tgz", safeName, manifest.Version))

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

	// Use package name directly (no scope support)
	packageName := manifest.Name

	// Create or get package
	pkg, err := s.DB.GetOrCreatePackage(packageName)
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
