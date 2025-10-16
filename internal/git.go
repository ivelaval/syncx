package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// formatGitURL converts base URL to proper git clone URL based on protocol
func FormatGitURL(baseURL, protocol string) string {
	switch protocol {
	case "ssh":
		if strings.HasPrefix(baseURL, "git@") {
			return baseURL
		}
		if strings.HasPrefix(baseURL, "gitlab.com:") {
			return "git@" + baseURL
		}
		return "git@" + baseURL
	case "http":
		if strings.HasPrefix(baseURL, "https://") {
			return baseURL
		}
		if strings.HasPrefix(baseURL, "git@") {
			return strings.Replace(strings.Replace(baseURL, "git@", "https://", 1), ":", "/", 1)
		}
		if strings.HasPrefix(baseURL, "gitlab.com:") {
			return "https://" + strings.Replace(baseURL, ":", "/", 1)
		}
		return "https://" + baseURL
	default:
		panic(fmt.Sprintf("Unsupported protocol: %s", protocol))
	}
}

// extractDirectoryPath extracts directory path from git URL
func ExtractDirectoryPath(gitURL string) string {
	if strings.Contains(gitURL, "gitlab.com") {
		parts := strings.Split(gitURL, "gitlab.com")
		if len(parts) > 1 {
			path := parts[1]
			if strings.HasPrefix(path, ":") || strings.HasPrefix(path, "/") {
				path = path[1:]
			}
			if strings.HasSuffix(path, ".git") {
				path = path[:len(path)-4]
			}
			return path
		}
	}
	return ""
}

// CreateProjectLocalPath creates the correct local path for a project based on its URL and base directory
func CreateProjectLocalPath(baseDir, projectURL string, group string) string {
	// Extract the path from the URL
	dirPath := ExtractDirectoryPath(projectURL)
	if dirPath == "" {
		// Fallback: create path from group and project name
		if group != "" && group != "Standalone" {
			// Convert group path to directory structure
			groupPath := strings.ToLower(strings.ReplaceAll(group, " ", "-"))
			groupPath = strings.ReplaceAll(groupPath, "/", "/")
			
			// Extract project name from URL
			parts := strings.Split(projectURL, "/")
			if len(parts) > 0 {
				projectName := parts[len(parts)-1]
				if strings.HasSuffix(projectName, ".git") {
					projectName = projectName[:len(projectName)-4]
				}
				return filepath.Join(baseDir, groupPath, projectName)
			}
		}
		return ""
	}
	
	return filepath.Join(baseDir, dirPath)
}

// EnsureDirectoryStructure creates all necessary parent directories for a project
func EnsureDirectoryStructure(projectPath string, logger *Logger) error {
	parentDir := filepath.Dir(projectPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory structure %s: %w", parentDir, err)
	}
	
	logger.Debug("Created directory structure: %s", parentDir)
	return nil
}

// isGitRepository checks if a directory is a git repository
func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}
	return false
}

// IsEmptyRepository checks if a git repository has no commits
func IsEmptyRepository(path string) bool {
	if !IsGitRepository(path) {
		return false
	}

	// Try to get the current commit hash
	// If this fails with exit code 128, the repository is empty (no commits)
	cmd := exec.Command("git", "-C", path, "rev-parse", "HEAD")
	err := cmd.Run()

	// If command fails, repository is empty (no commits yet)
	return err != nil
}

// CloneRepository clones a repository to the specified local path
func CloneRepository(repoURL, localPath string, logger *Logger) OperationResult {
	start := time.Now()
	
	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(localPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return OperationResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create directory %s: %v", parentDir, err),
			IsClone: true,
			Duration: time.Since(start).String(),
		}
	}

	logger.Cloning("%s -> %s", repoURL, localPath)

	// Execute git clone with more options for robustness
	cmd := exec.Command("git", "clone", "--progress", "--verbose", repoURL, localPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return OperationResult{
			Success: false,
			Message: fmt.Sprintf("Failed to clone %s: %v - Output: %s", repoURL, err, string(output)),
			IsClone: true,
			Duration: time.Since(start).String(),
		}
	}

	// Verify the clone was successful by checking if .git directory exists
	if !IsGitRepository(localPath) {
		return OperationResult{
			Success: false,
			Message: fmt.Sprintf("Clone completed but no git repository found at %s", localPath),
			IsClone: true,
			Duration: time.Since(start).String(),
		}
	}

	logger.Success("Cloned: %s", filepath.Base(localPath))
	return OperationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully cloned %s", repoURL),
		IsClone: true,
		Duration: time.Since(start).String(),
	}
}

// PullRepository pulls latest changes from a git repository
func PullRepository(localPath string, logger *Logger) OperationResult {
	start := time.Now()

	// Check if repository is empty (no commits)
	if IsEmptyRepository(localPath) {
		logger.Warning("Repository is empty (no commits): %s", filepath.Base(localPath))
		return OperationResult{
			Success: false,
			Message: "Repository is empty (no commits)",
			IsClone: false,
			IsEmpty: true,
			Duration: time.Since(start).String(),
		}
	}

	logger.Pulling("Getting latest changes: %s", localPath)

	// First, try to fetch to ensure we have the latest refs
	fetchCmd := exec.Command("git", "-C", localPath, "fetch", "--all")
	if err := fetchCmd.Run(); err != nil {
		logger.Warning("Fetch failed for %s: %v", localPath, err)
	}

	// Then pull
	cmd := exec.Command("git", "-C", localPath, "pull", "--ff-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try again without --ff-only in case there are conflicts
		logger.Warning("Fast-forward pull failed, trying regular pull...")
		cmd = exec.Command("git", "-C", localPath, "pull")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return OperationResult{
				Success: false,
				Message: fmt.Sprintf("Failed to pull %s: %v - Output: %s", localPath, err, string(output)),
				IsClone: false,
				IsEmpty: false,
				Duration: time.Since(start).String(),
			}
		}
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Already up to date") || strings.Contains(outputStr, "Already up-to-date") {
		logger.Info("Already up to date: %s", filepath.Base(localPath))
		return OperationResult{
			Success: true,
			Message: "Already up to date",
			IsClone: false,
			IsEmpty: false,
			Duration: time.Since(start).String(),
		}
	} else {
		logger.Updated("Updated: %s", filepath.Base(localPath))
		return OperationResult{
			Success: true,
			Message: fmt.Sprintf("Successfully updated - %s", strings.TrimSpace(outputStr)),
			IsClone: false,
			IsEmpty: false,
			Duration: time.Since(start).String(),
		}
	}
}

// CloneOrUpdateRepository clones a repository or updates if it already exists
func CloneOrUpdateRepository(project ProjectInfo, dryRun bool, logger *Logger) OperationResult {
	if dryRun {
		if _, err := os.Stat(project.LocalPath); err == nil && IsGitRepository(project.LocalPath) {
			logger.DryRun("Would pull: %s", project.LocalPath)
			return OperationResult{Success: true, Project: project, Message: "Would pull", IsClone: false}
		} else {
			logger.DryRun("Would clone: %s -> %s", project.GitURL, project.LocalPath)
			return OperationResult{Success: true, Project: project, Message: "Would clone", IsClone: true}
		}
	}

	// Check if directory exists and is a git repository
	if _, err := os.Stat(project.LocalPath); err == nil {
		if IsGitRepository(project.LocalPath) {
			// It's a git repository, pull latest changes
			result := PullRepository(project.LocalPath, logger)
			result.Project = project
			
			// Update tracker if operation was successful
			if result.Success {
				// Find the tracker in the output directory (parent of all repos)
				outputDir := filepath.Dir(filepath.Dir(project.LocalPath))
				if tracker, err := LoadOrCreateTracker(outputDir, ""); err == nil {
					if commitHash, err := GetCurrentCommitHash(project.LocalPath); err == nil {
						UpdateTrackedProject(tracker, project, "updated", commitHash)
						SaveTracker(tracker)
					}
				}
			}
			
			return result
		} else {
			// Directory exists but is not a git repository
			logger.Warning("Directory exists but is not a git repository: %s", project.LocalPath)
			return OperationResult{
				Success: false,
				Project: project,
				Message: "Directory exists but is not a git repository",
				IsClone: false,
			}
		}
	}

	// Directory doesn't exist, clone the repository
	result := CloneRepository(project.GitURL, project.LocalPath, logger)
	result.Project = project
	
	// Update tracker if clone was successful
	if result.Success {
		if tracker, err := LoadOrCreateTracker(filepath.Dir(project.LocalPath), ""); err == nil {
			if commitHash, err := GetCurrentCommitHash(project.LocalPath); err == nil {
				UpdateTrackedProject(tracker, project, "cloned", commitHash)
				SaveTracker(tracker)
			}
		}
	}
	
	return result
}

// CloneOrUpdateRepositorySilent performs the same operation but with minimal logging
// Used for batch operations to avoid cluttering the progress bar
func CloneOrUpdateRepositorySilent(project ProjectInfo, dryRun bool, logger *Logger) OperationResult {
	if dryRun {
		return OperationResult{
			Success:  true,
			Project:  project,
			Message:  fmt.Sprintf("DRY RUN: Would clone/update %s", project.Name),
			IsClone:  true,
			Duration: "0s",
		}
	}

	// Check if directory exists and is a git repository
	if _, err := os.Stat(project.LocalPath); err == nil {
		if IsGitRepository(project.LocalPath) {
			// It's a git repository, pull latest changes (silently)
			result := PullRepositorySilent(project.LocalPath)
			result.Project = project
			
			// Update tracker if operation was successful
			if result.Success {
				// Find the tracker in the output directory (parent of all repos)
				outputDir := filepath.Dir(filepath.Dir(project.LocalPath))
				if tracker, err := LoadOrCreateTracker(outputDir, ""); err == nil {
					if commitHash, err := GetCurrentCommitHash(project.LocalPath); err == nil {
						UpdateTrackedProject(tracker, project, "updated", commitHash)
						SaveTracker(tracker)
					}
				}
			}
			
			return result
		} else {
			// Directory exists but is not a git repository
			return OperationResult{
				Success: false,
				Project: project,
				Message: "Directory exists but is not a git repository",
				IsClone: false,
			}
		}
	}

	// Directory doesn't exist, clone the repository (silently)
	result := CloneRepositorySilent(project.GitURL, project.LocalPath)
	result.Project = project
	
	// Update tracker if clone was successful
	if result.Success {
		// Find the tracker in the output directory (parent of all repos)
		outputDir := filepath.Dir(filepath.Dir(project.LocalPath))
		if tracker, err := LoadOrCreateTracker(outputDir, ""); err == nil {
			if commitHash, err := GetCurrentCommitHash(project.LocalPath); err == nil {
				UpdateTrackedProject(tracker, project, "cloned", commitHash)
				SaveTracker(tracker)
			}
		}
	}
	
	return result
}

// CloneRepositorySilent clones a repository without logging output
func CloneRepositorySilent(repoURL, localPath string) OperationResult {
	start := time.Now()
	
	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		return OperationResult{
			Success:  false,
			Message:  fmt.Sprintf("Failed to create directory: %v", err),
			IsClone:  true,
			Duration: time.Since(start).String(),
		}
	}

	// Execute git clone silently
	cmd := exec.Command("git", "clone", repoURL, localPath)
	if err := cmd.Run(); err != nil {
		return OperationResult{
			Success:  false,
			Message:  fmt.Sprintf("Clone failed: %v", err),
			IsClone:  true,
			Duration: time.Since(start).String(),
		}
	}

	return OperationResult{
		Success:  true,
		Message:  "Cloned successfully",
		IsClone:  true,
		Duration: time.Since(start).String(),
	}
}

// PullRepositorySilent pulls latest changes without logging output
func PullRepositorySilent(localPath string) OperationResult {
	start := time.Now()

	// Check if repository is empty (no commits)
	if IsEmptyRepository(localPath) {
		return OperationResult{
			Success:  false,
			Message:  "Repository is empty (no commits)",
			IsClone:  false,
			IsEmpty:  true,
			Duration: time.Since(start).String(),
		}
	}

	// Execute git pull silently
	cmd := exec.Command("git", "-C", localPath, "pull")
	if err := cmd.Run(); err != nil {
		return OperationResult{
			Success:  false,
			Message:  fmt.Sprintf("Pull failed: %v", err),
			IsClone:  false,
			IsEmpty:  false,
			Duration: time.Since(start).String(),
		}
	}

	return OperationResult{
		Success:  true,
		Message:  "Updated successfully",
		IsClone:  false,
		IsEmpty:  false,
		Duration: time.Since(start).String(),
	}
}