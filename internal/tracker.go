package internal

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const TrackingFileName = ".olive-clone-tracker.json"

// truncateHash safely truncates a hash string for display
func truncateHash(hash string) string {
	if len(hash) < 8 {
		if hash == "" {
			return "none"
		}
		return hash
	}
	return hash[:8] + "..."
}

// LoadOrCreateTracker loads existing tracker or creates a new one
func LoadOrCreateTracker(outputDir, inventoryFile string) (*ProjectTracker, error) {
	trackerPath := filepath.Join(outputDir, TrackingFileName)
	
	// Try to load existing tracker
	if _, err := os.Stat(trackerPath); err == nil {
		data, err := ioutil.ReadFile(trackerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read tracker file: %w", err)
		}
		
		var tracker ProjectTracker
		if err := json.Unmarshal(data, &tracker); err != nil {
			return nil, fmt.Errorf("failed to parse tracker file: %w", err)
		}
		
		return &tracker, nil
	}
	
	// Create new tracker
	tracker := &ProjectTracker{
		LastSync:        time.Now().Format(time.RFC3339),
		OutputDirectory: outputDir,
		InventoryFile:   inventoryFile,
		Projects:        []TrackedProject{},
	}
	
	return tracker, nil
}

// SaveTracker saves the tracker to disk
func SaveTracker(tracker *ProjectTracker) error {
	trackerPath := filepath.Join(tracker.OutputDirectory, TrackingFileName)
	
	// Update last sync time
	tracker.LastSync = time.Now().Format(time.RFC3339)
	
	// Marshal to JSON with proper formatting
	data, err := json.MarshalIndent(tracker, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tracker: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(trackerPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tracker file: %w", err)
	}
	
	return nil
}

// CalculateInventoryHash calculates MD5 hash of inventory file for change detection
func CalculateInventoryHash(inventoryPath string) (string, error) {
	data, err := ioutil.ReadFile(inventoryPath)
	if err != nil {
		return "", fmt.Errorf("failed to read inventory file: %w", err)
	}
	
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash), nil
}

// CompareWithInventory compares current inventory with tracked state
func CompareWithInventory(tracker *ProjectTracker, currentProjects []ProjectInfo, inventoryPath string, outputDir, protocol string, logger *Logger) (*ProjectDiff, error) {
	logger.Header("🔍 Analyzing Project Differences")
	
	// Calculate current inventory hash
	currentHash, err := CalculateInventoryHash(inventoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate inventory hash: %w", err)
	}
	
	logger.Info("📊 Current inventory hash: %s", truncateHash(currentHash))
	logger.Info("📊 Tracked inventory hash: %s", truncateHash(tracker.InventoryHash))
	
	diff := &ProjectDiff{
		NewProjects:       []ProjectInfo{},
		RemovedProjects:   []ProjectInfo{},
		ModifiedProjects:  []ProjectInfo{},
		UnchangedProjects: []ProjectInfo{},
	}
	
	// Create maps for quick lookup
	trackedMap := make(map[string]TrackedProject)
	for _, tracked := range tracker.Projects {
		key := fmt.Sprintf("%s|%s", tracked.Name, tracked.URL)
		trackedMap[key] = tracked
	}
	
	currentMap := make(map[string]ProjectInfo)
	for _, current := range currentProjects {
		key := fmt.Sprintf("%s|%s", current.Name, current.URL)
		
		// Populate ProjectInfo with local path and git URL using improved logic
		current.GitURL = FormatGitURL(current.URL, protocol)
		current.LocalPath = CreateProjectLocalPath(outputDir, current.URL, current.Group)
		
		// If we couldn't determine the path, skip this project
		if current.LocalPath == "" {
			logger.Warning("Could not determine local path for project %s (URL: %s)", current.Name, current.URL)
			continue
		}
		
		currentMap[key] = current
	}
	
	// Find new projects (in current but not in tracked)
	for key, current := range currentMap {
		if _, exists := trackedMap[key]; !exists {
			diff.NewProjects = append(diff.NewProjects, current)
		}
	}
	
	// Find removed projects (in tracked but not in current)
	for key, tracked := range trackedMap {
		if _, exists := currentMap[key]; !exists {
			// Convert tracked back to ProjectInfo for consistency
			project := ProjectInfo{
				Name:      tracked.Name,
				URL:       tracked.URL,
				Group:     tracked.Group,
				LocalPath: tracked.LocalPath,
				GitURL:    tracked.GitURL,
			}
			diff.RemovedProjects = append(diff.RemovedProjects, project)
		}
	}
	
	// Find unchanged projects (exist in both)
	for key, current := range currentMap {
		if tracked, exists := trackedMap[key]; exists {
			// Check if the project has been moved or URL changed
			expectedLocalPath := current.LocalPath
			if tracked.LocalPath != expectedLocalPath {
				// Project path changed, treat as modified
				diff.ModifiedProjects = append(diff.ModifiedProjects, current)
			} else {
				diff.UnchangedProjects = append(diff.UnchangedProjects, current)
			}
		}
	}
	
	// Log summary
	logger.Info("📊 Analysis Results:")
	logger.Info("   ➕ New projects: %d", len(diff.NewProjects))
	logger.Info("   ➖ Removed projects: %d", len(diff.RemovedProjects))
	logger.Info("   🔄 Modified projects: %d", len(diff.ModifiedProjects))
	logger.Info("   ✅ Unchanged projects: %d", len(diff.UnchangedProjects))
	
	// Update tracker hash
	tracker.InventoryHash = currentHash
	
	return diff, nil
}

// CheckGitChanges checks if a git repository has remote changes
func CheckGitChanges(localPath string, logger *Logger) (bool, string, error) {
	if !IsGitRepository(localPath) {
		return false, "", fmt.Errorf("not a git repository: %s", localPath)
	}

	// Check if repository is empty (no commits)
	if IsEmptyRepository(localPath) {
		return false, "", fmt.Errorf("repository is empty (no commits)")
	}

	// Get current commit hash
	currentHashCmd := exec.Command("git", "-C", localPath, "rev-parse", "HEAD")
	currentHashOutput, err := currentHashCmd.Output()
	if err != nil {
		return false, "", fmt.Errorf("failed to get current commit hash: %w", err)
	}
	currentHash := strings.TrimSpace(string(currentHashOutput))
	
	// Fast fetch from remote with timeout (only current branch)
	if err := runGitCommandWithTimeout(20*time.Second, "-C", localPath, "fetch", "--quiet"); err != nil {
		logger.Warning("Failed to fetch for %s: %v", localPath, err)
		return false, currentHash, nil // Continue even if fetch fails
	}
	
	// Get remote commit hash (assuming origin/main or origin/master)
	remoteHashCmd := exec.Command("git", "-C", localPath, "rev-parse", "@{upstream}")
	remoteHashOutput, err := remoteHashCmd.Output()
	if err != nil {
		// Try origin/main
		remoteHashCmd = exec.Command("git", "-C", localPath, "rev-parse", "origin/main")
		remoteHashOutput, err = remoteHashCmd.Output()
		if err != nil {
			// Try origin/master
			remoteHashCmd = exec.Command("git", "-C", localPath, "rev-parse", "origin/master")
			remoteHashOutput, err = remoteHashCmd.Output()
			if err != nil {
				logger.Warning("Could not determine remote HEAD for %s", localPath)
				return false, currentHash, nil
			}
		}
	}
	remoteHash := strings.TrimSpace(string(remoteHashOutput))
	
	hasChanges := currentHash != remoteHash
	logger.Debug("Git check for %s: local=%s remote=%s hasChanges=%v", 
		filepath.Base(localPath), currentHash[:8], remoteHash[:8], hasChanges)
	
	return hasChanges, currentHash, nil
}

// UpdateTrackedProject updates or adds a project to the tracker
func UpdateTrackedProject(tracker *ProjectTracker, project ProjectInfo, status string, commitHash string) {
	now := time.Now().Format(time.RFC3339)
	
	// Find existing project
	for i, tracked := range tracker.Projects {
		if tracked.Name == project.Name && tracked.URL == project.URL {
			// Update existing
			tracker.Projects[i].LocalPath = project.LocalPath
			tracker.Projects[i].GitURL = project.GitURL
			tracker.Projects[i].Group = project.Group
			tracker.Projects[i].LastUpdated = now
			tracker.Projects[i].LastCommitHash = commitHash
			tracker.Projects[i].Status = status
			return
		}
	}
	
	// Add new project
	tracked := TrackedProject{
		Name:           project.Name,
		URL:            project.URL,
		Group:          project.Group,
		LocalPath:      project.LocalPath,
		GitURL:         project.GitURL,
		LastCloned:     now,
		LastUpdated:    now,
		LastCommitHash: commitHash,
		Status:         status,
	}
	tracker.Projects = append(tracker.Projects, tracked)
}

// RemoveTrackedProject removes a project from the tracker
func RemoveTrackedProject(tracker *ProjectTracker, project ProjectInfo) {
	newProjects := []TrackedProject{}
	for _, tracked := range tracker.Projects {
		if tracked.Name != project.Name || tracked.URL != project.URL {
			newProjects = append(newProjects, tracked)
		}
	}
	tracker.Projects = newProjects
}

// GetCurrentCommitHash gets the current commit hash of a git repository
func GetCurrentCommitHash(localPath string) (string, error) {
	if !IsGitRepository(localPath) {
		return "", fmt.Errorf("not a git repository: %s", localPath)
	}
	
	cmd := exec.Command("git", "-C", localPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}