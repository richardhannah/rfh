package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"rulestack/internal/manifest"
	"rulestack/internal/version"
)

// promptUserChoice prompts user for a yes/no choice
func promptUserChoice(question string) (bool, error) {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Printf("%s (y/n): ", question)
		if !scanner.Scan() {
			return false, fmt.Errorf("failed to read input")
		}
		
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		switch response {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Println("Please enter 'y' or 'n'")
		}
	}
}

// promptUserInput prompts user for text input
func promptUserInput(question string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Printf("%s: ", question)
	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}
	
	return strings.TrimSpace(scanner.Text()), nil
}

// promptPackageSelection shows existing packages and prompts user to select one
func promptPackageSelection(packageManifests manifest.PackageManifestFile) (int, error) {
	if len(packageManifests) == 0 {
		return -1, fmt.Errorf("no existing packages found")
	}
	
	fmt.Println("\nExisting packages:")
	for i, m := range packageManifests {
		fmt.Printf("  %d) %s (v%s) - %s\n", i+1, m.Name, m.Version, m.Description)
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Printf("Select package (1-%d): ", len(packageManifests))
		if !scanner.Scan() {
			return -1, fmt.Errorf("failed to read input")
		}
		
		choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || choice < 1 || choice > len(packageManifests) {
			fmt.Printf("Please enter a number between 1 and %d\n", len(packageManifests))
			continue
		}
		
		return choice - 1, nil
	}
}

// promptNewVersion prompts user for new version number with validation
func promptNewVersion(currentVersion string) (string, error) {
	nextPatch, err := version.IncrementPatchVersion(currentVersion)
	if err != nil {
		return "", err
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Printf("Enter new version (current: %s, default: %s): ", currentVersion, nextPatch)
		if !scanner.Scan() {
			return "", fmt.Errorf("failed to read input")
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			return nextPatch, nil
		}
		
		if err := version.ValidateVersionIncrease(currentVersion, input); err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		return input, nil
	}
}



// isValidMdcFile checks if a file is a valid .mdc rules file
func isValidMdcFile(filePath string) bool {
	if !strings.HasSuffix(strings.ToLower(filePath), ".mdc") {
		return false
	}
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	
	return true
}

// ensureDirectoryExists creates directory if it doesn't exist
func ensureDirectoryExists(dirPath string) error {
	return os.MkdirAll(dirPath, 0o755)
}

// getPackageDirectory returns the package directory path in .rulestack
func getPackageDirectory(packageName, version string) string {
	return filepath.Join(".rulestack", fmt.Sprintf("%s.%s", packageName, version))
}

// getStagingDirectory returns the staging directory path
func getStagingDirectory() string {
	return filepath.Join(".rulestack", "staged")
}