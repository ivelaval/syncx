package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// EnsureOutputDirectory validates and creates the output directory if needed
func EnsureOutputDirectory(outputPath string, logger *Logger) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for %s: %w", outputPath, err)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Directory doesn't exist, create it
		logger.Info("Creating output directory: %s", absPath)
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create output directory %s: %w", absPath, err)
		}
		logger.Success("Created output directory: %s", absPath)
	} else if err != nil {
		// Other error checking directory
		return "", fmt.Errorf("error checking output directory %s: %w", absPath, err)
	} else {
		// Directory exists, verify it's writable
		if !isWritable(absPath) {
			return "", fmt.Errorf("output directory %s is not writable", absPath)
		}
		logger.Info("Using output directory: %s", absPath)
	}

	return absPath, nil
}

// ValidateOutputPath checks if the provided path is valid for output
func ValidateOutputPath(outputPath string) error {
	// Check if it's an absolute path or relative path
	if !filepath.IsAbs(outputPath) {
		// It's relative, which is fine
		logger := NewLogger(true)
		logger.Info("Using relative output path: %s", outputPath)
	}

	// Check for invalid characters or patterns
	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	// Expand ~ to home directory if present
	if filepath.HasPrefix(outputPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		outputPath = filepath.Join(homeDir, outputPath[2:])
	}

	return nil
}

// isWritable checks if a directory is writable
func isWritable(path string) bool {
	// Try to create a temporary file in the directory
	testFile := filepath.Join(path, ".write_test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}

// GetSmartOutputDirectory provides intelligent defaults for output directory
func GetSmartOutputDirectory(executablePath string) string {
	if executablePath == "" {
		return "./repositories"
	}

	// Get the directory containing the executable
	execDir := filepath.Dir(executablePath)
	
	// Check if we're in a typical installation location
	if filepath.Base(execDir) == "bin" || 
	   strings.Contains(execDir, "/usr/local/bin") || 
	   strings.Contains(execDir, "/usr/bin") {
		// If installed system-wide, use user's home directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, "repositories")
		}
	}

	// Otherwise, use parent directory + repositories
	parentDir := filepath.Dir(execDir)
	return filepath.Join(parentDir, "repositories")
}

// ShowOutputDirectoryInfo displays helpful information about the output directory
func ShowOutputDirectoryInfo(outputDir string, logger *Logger) {
	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		absDir = outputDir
	}

	logger.Header("📂 Output Directory Information")
	color.New(color.FgCyan).Printf("   Output Path: %s\n", absDir)
	
	// Show relative path from current working directory
	if cwd, err := os.Getwd(); err == nil {
		if relPath, err := filepath.Rel(cwd, absDir); err == nil {
			color.New(color.FgCyan).Printf("   Relative Path: %s\n", relPath)
		}
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		color.New(color.FgYellow).Printf("   Status: Will be created\n")
	} else {
		color.New(color.FgGreen).Printf("   Status: Exists\n")
		
		// Count existing repositories if any
		if entries, err := os.ReadDir(absDir); err == nil {
			dirCount := 0
			for _, entry := range entries {
				if entry.IsDir() {
					dirCount++
				}
			}
			color.New(color.FgCyan).Printf("   Existing directories: %d\n", dirCount)
		}
	}
	
	fmt.Println()
}