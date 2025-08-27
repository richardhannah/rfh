package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"rulestack/internal/manifest"
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
func promptPackageSelection(manifests manifest.ManifestFile) (int, error) {
	if len(manifests) == 0 {
		return -1, fmt.Errorf("no existing packages found")
	}
	
	fmt.Println("\nExisting packages:")
	for i, m := range manifests {
		fmt.Printf("  %d) %s (v%s) - %s\n", i+1, m.Name, m.Version, m.Description)
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Printf("Select package (1-%d): ", len(manifests))
		if !scanner.Scan() {
			return -1, fmt.Errorf("failed to read input")
		}
		
		choice, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || choice < 1 || choice > len(manifests) {
			fmt.Printf("Please enter a number between 1 and %d\n", len(manifests))
			continue
		}
		
		return choice - 1, nil
	}
}

// promptNewVersion prompts user for new version number with validation
func promptNewVersion(currentVersion string) (string, error) {
	nextPatch, err := incrementPatchVersion(currentVersion)
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
		
		if err := validateVersionIncrease(currentVersion, input); err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		return input, nil
	}
}

// incrementPatchVersion increments the patch version (x.y.z -> x.y.z+1)
func incrementPatchVersion(version string) (string, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}
	
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("invalid patch version: %s", parts[2])
	}
	
	return fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch+1), nil
}

// validateVersionIncrease ensures new version is greater than current version
func validateVersionIncrease(currentVersion, newVersion string) error {
	currentParts := strings.Split(currentVersion, ".")
	newParts := strings.Split(newVersion, ".")
	
	if len(currentParts) != 3 || len(newParts) != 3 {
		return fmt.Errorf("versions must be in semantic version format (x.y.z)")
	}
	
	for i := 0; i < 3; i++ {
		current, err := strconv.Atoi(currentParts[i])
		if err != nil {
			return fmt.Errorf("invalid current version: %s", currentVersion)
		}
		
		new, err := strconv.Atoi(newParts[i])
		if err != nil {
			return fmt.Errorf("invalid new version: %s", newVersion)
		}
		
		if new > current {
			return nil // New version is higher
		} else if new < current {
			return fmt.Errorf("new version %s must be greater than current version %s", newVersion, currentVersion)
		}
		// If equal, continue to next part
	}
	
	return fmt.Errorf("new version %s must be greater than current version %s", newVersion, currentVersion)
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