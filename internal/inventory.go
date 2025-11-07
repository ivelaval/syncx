package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// LoadInventory loads and parses the inventory JSON file
func LoadInventory(filename string) (*Inventory, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	var inventory Inventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", filename, err)
	}

	return &inventory, nil
}

// CollectAllProjects recursively collects all projects from inventory structure
func CollectAllProjects(inventory Inventory) []ProjectInfo {
	var allProjects []ProjectInfo
	projectsFound := make(map[string]bool) // Track duplicates

	var collectFromGroups func(groups []Group, parentGroup string)
	collectFromGroups = func(groups []Group, parentGroup string) {
		for _, group := range groups {
			if group.Skip {
				continue
			}

			groupName := group.Name
			if parentGroup != "" {
				groupName = parentGroup + "/" + groupName
			}

			// Add projects from this group
			for _, project := range group.Projects {
				if project.URL != "" && project.Name != "" {
					// Create unique key to avoid duplicates
					projectKey := fmt.Sprintf("%s|%s", project.Name, project.URL)
					if !projectsFound[projectKey] {
						allProjects = append(allProjects, ProjectInfo{
							Name:  project.Name,
							URL:   project.URL,
							Group: groupName,
						})
						projectsFound[projectKey] = true
					}
				}
			}

			// Recursively process subgroups
			collectFromGroups(group.Groups, groupName)
		}
	}

	// Determine which structure to use (new format with Root or legacy format)
	var groups []Group
	var projects []Project

	if inventory.Root != nil {
		// New format: data is inside the "root" property
		groups = inventory.Root.Groups
		projects = inventory.Root.Projects
	} else {
		// Legacy format: data is at the top level
		groups = inventory.Groups
		projects = inventory.Projects
	}

	// Collect from main groups
	collectFromGroups(groups, "")

	// Collect standalone projects
	for _, project := range projects {
		if project.URL != "" && project.Name != "" {
			projectKey := fmt.Sprintf("%s|%s", project.Name, project.URL)
			if !projectsFound[projectKey] {
				allProjects = append(allProjects, ProjectInfo{
					Name:  project.Name,
					URL:   project.URL,
					Group: "Standalone",
				})
				projectsFound[projectKey] = true
			}
		}
	}

	return allProjects
}

// ScanAndClassifyProjectsWithTracking scans using the new tracking system (backward compatible)
func ScanAndClassifyProjectsWithTracking(allProjects []ProjectInfo, baseDir, protocol, inventoryFile string, logger *Logger) ([]ProjectInfo, []ProjectInfo, []ProjectInfo, error) {
	return ScanAndClassifyProjectsWithTrackingSkipCheck(allProjects, baseDir, protocol, inventoryFile, false, logger)
}

// ScanAndClassifyProjectsWithTrackingSkipCheck scans using the new tracking system with skip check option
func ScanAndClassifyProjectsWithTrackingSkipCheck(allProjects []ProjectInfo, baseDir, protocol, inventoryFile string, skipCheck bool, logger *Logger) ([]ProjectInfo, []ProjectInfo, []ProjectInfo, error) {
	// Spinner is handled by caller

	// Load or create tracker
	tracker, err := LoadOrCreateTracker(baseDir, inventoryFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize tracker: %w", err)
	}

	// Compare with inventory to find differences
	diff, err := CompareWithInventory(tracker, allProjects, inventoryFile, baseDir, protocol, logger)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to compare with inventory: %w", err)
	}

	var projectsToClone []ProjectInfo
	var projectsToUpdate []ProjectInfo
	var projectsUpToDate []ProjectInfo

	// Process new projects (need to be cloned)
	for _, project := range diff.NewProjects {
		// Ensure directory structure exists using improved function
		if err := EnsureDirectoryStructure(project.LocalPath, logger); err != nil {
			logger.Warning("Failed to create directory structure for %s: %v", project.Name, err)
			continue
		}
		
		logger.Info("📁 Ensured directory structure exists for: %s -> %s", project.Name, project.LocalPath)

		if _, err := os.Stat(project.LocalPath); err == nil {
			if IsGitRepository(project.LocalPath) {
				logger.Info("🔍 Found existing repository for new project: %s", project.Name)
				// Check if it needs updates
				hasChanges, currentHash, checkErr := CheckGitChanges(project.LocalPath, logger)
				if checkErr != nil {
					// Check if error is due to empty repository
					if strings.Contains(checkErr.Error(), "empty") {
						logger.Warning("Repository %s is empty (no commits), marking for pull attempt", project.Name)
						projectsToUpdate = append(projectsToUpdate, project)
					} else {
						logger.Warning("Failed to check git changes for %s: %v", project.Name, checkErr)
						projectsToUpdate = append(projectsToUpdate, project)
					}
				} else if hasChanges {
					logger.Info("🔄 Repository %s has remote changes", project.Name)
					projectsToUpdate = append(projectsToUpdate, project)
				} else {
					logger.Info("✅ Repository %s is up to date", project.Name)
					projectsUpToDate = append(projectsUpToDate, project)
					// Update tracker with current info
					UpdateTrackedProject(tracker, project, "up-to-date", currentHash)
				}
			} else {
				logger.Warning("Directory exists but is not a git repository: %s", project.LocalPath)
				// Create alternative path
				altPath := project.LocalPath + "_clone"
				project.LocalPath = altPath
				projectsToClone = append(projectsToClone, project)
			}
		} else {
			logger.Info("📥 New project to clone: %s (Group: %s)", project.Name, project.Group)
			projectsToClone = append(projectsToClone, project)
		}
	}

	// Process modified projects
	for _, project := range diff.ModifiedProjects {
		logger.Info("🔄 Modified project detected: %s", project.Name)
		projectsToUpdate = append(projectsToUpdate, project)
	}

	// Process unchanged projects (check for git changes)
	if skipCheck {
		logger.Info("⚡ Skip-check enabled: Marking all unchanged projects as up-to-date (no remote verification)")
		for _, project := range diff.UnchangedProjects {
			if _, err := os.Stat(project.LocalPath); err == nil && IsGitRepository(project.LocalPath) {
				logger.Debug("✅ Repository %s assumed up to date (skip-check)", project.Name)
				projectsUpToDate = append(projectsUpToDate, project)
			} else {
				logger.Warning("Tracked project not found or not a git repo: %s", project.LocalPath)
				projectsToClone = append(projectsToClone, project)
			}
		}
	} else {
		logger.Info("🔍 Checking unchanged projects for remote updates...")
		for _, project := range diff.UnchangedProjects {
			if _, err := os.Stat(project.LocalPath); err == nil && IsGitRepository(project.LocalPath) {
				hasChanges, currentHash, checkErr := CheckGitChanges(project.LocalPath, logger)
				if checkErr != nil {
					// Check if error is due to empty repository
					if strings.Contains(checkErr.Error(), "empty") {
						logger.Warning("Repository %s is empty (no commits), marking for pull attempt", project.Name)
						projectsToUpdate = append(projectsToUpdate, project)
					} else {
						logger.Warning("Failed to check git changes for %s: %v", project.Name, checkErr)
						projectsToUpdate = append(projectsToUpdate, project)
					}
				} else if hasChanges {
					logger.Info("🔄 Repository %s has remote changes", project.Name)
					projectsToUpdate = append(projectsToUpdate, project)
				} else {
					logger.Debug("✅ Repository %s is up to date", project.Name)
					projectsUpToDate = append(projectsUpToDate, project)
					// Update tracker timestamp
					UpdateTrackedProject(tracker, project, "up-to-date", currentHash)
				}
			} else {
				logger.Warning("Tracked project not found or not a git repo: %s", project.LocalPath)
				projectsToClone = append(projectsToClone, project)
			}
		}
	}

	// Handle removed projects
	for _, project := range diff.RemovedProjects {
		logger.Info("➖ Project removed from inventory: %s", project.Name)
		RemoveTrackedProject(tracker, project)
	}

	// Save updated tracker
	if err := SaveTracker(tracker); err != nil {
		logger.Warning("Failed to save tracker: %v", err)
	}

	logger.Info("📊 Smart Analysis Results:")
	logger.Info("   📥 To clone: %d repositories", len(projectsToClone))
	logger.Info("   🔄 To update: %d repositories", len(projectsToUpdate))
	logger.Info("   ✅ Up to date: %d repositories", len(projectsUpToDate))
	logger.Info("   📈 Total: %d repositories", len(projectsToClone)+len(projectsToUpdate)+len(projectsUpToDate))

	return projectsToClone, projectsToUpdate, projectsUpToDate, nil
}

// ScanAndClassifyProjects - Legacy function for backward compatibility
func ScanAndClassifyProjects(allProjects []ProjectInfo, baseDir, protocol string, logger *Logger) ([]ProjectInfo, []ProjectInfo) {
	// Use the new tracking system but return in old format
	projectsToClone, projectsToUpdate, _, err := ScanAndClassifyProjectsWithTracking(allProjects, baseDir, protocol, "", logger)
	if err != nil {
		logger.Warning("Failed to use tracking system, falling back to legacy scan: %v", err)
		// Fallback to simple scanning logic here if needed
		return []ProjectInfo{}, []ProjectInfo{}
	}
	
	return projectsToClone, projectsToUpdate
}

// FilterProjectsByGroup allows filtering projects by group pattern
func FilterProjectsByGroup(projects []ProjectInfo, groupPattern string) []ProjectInfo {
	if groupPattern == "" {
		return projects
	}

	var filtered []ProjectInfo
	for _, project := range projects {
		if project.Group == groupPattern {
			filtered = append(filtered, project)
		}
	}
	return filtered
}

// GetUniqueGroups returns a list of unique groups from projects
func GetUniqueGroups(projects []ProjectInfo) []string {
	groupMap := make(map[string]bool)
	var groups []string

	for _, project := range projects {
		if !groupMap[project.Group] {
			groups = append(groups, project.Group)
			groupMap[project.Group] = true
		}
	}

	return groups
}

// UpdatePhysicalLocation updates the physical location in the inventory file
func UpdatePhysicalLocation(inventoryPath, newLocation string) error {
	// Read current inventory
	inventory, err := LoadInventory(inventoryPath)
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	// Update physical location
	inventory.PhysicalLocation = newLocation

	// Write back to file with proper formatting
	return SaveInventory(inventoryPath, inventory)
}

// SaveInventory saves the inventory structure back to JSON file with proper formatting
func SaveInventory(inventoryPath string, inventory *Inventory) error {
	// Convert to JSON with proper indentation
	jsonData, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(inventoryPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write inventory file: %w", err)
	}

	return nil
}

// GetPhysicalLocation returns the physical location from inventory, with fallback
func GetPhysicalLocation(inventory *Inventory, fallbackDir string) string {
	if inventory.PhysicalLocation != "" {
		return inventory.PhysicalLocation
	}
	return fallbackDir
}

// ValidateAndShowInventoryStats validates and shows detailed statistics about the inventory
func ValidateAndShowInventoryStats(inventory *Inventory, logger *Logger) {
	logger.Header("📋 Inventory Analysis")
	
	allProjects := CollectAllProjects(*inventory)
	groups := GetUniqueGroups(allProjects)
	
	logger.Info("📊 Inventory Statistics:")
	logger.Info("   📁 Total Groups: %d", len(groups))
	logger.Info("   📦 Total Projects: %d", len(allProjects))
	
	if inventory.PhysicalLocation != "" {
		logger.Info("   📍 Physical Location: %s", inventory.PhysicalLocation)
	} else {
		logger.Warning("   ⚠️  No physical location set")
	}
	
	// Show groups breakdown
	logger.Info("📁 Groups breakdown:")
	for _, group := range groups {
		projectCount := len(FilterProjectsByGroup(allProjects, group))
		logger.Info("   📁 %s: %d projects", group, projectCount)
	}
	
	// Validate for potential issues
	logger.Info("🔍 Validation:")
	emptyProjects := 0
	invalidURLs := 0
	
	for _, project := range allProjects {
		if project.Name == "" {
			emptyProjects++
		}
		if project.URL == "" || (!strings.Contains(project.URL, "gitlab.com") && !strings.Contains(project.URL, "github.com")) {
			invalidURLs++
		}
	}
	
	if emptyProjects > 0 {
		logger.Warning("   ⚠️  Found %d projects with empty names", emptyProjects)
	}
	if invalidURLs > 0 {
		logger.Warning("   ⚠️  Found %d projects with invalid URLs", invalidURLs)
	}
	
	if emptyProjects == 0 && invalidURLs == 0 {
		logger.Success("   ✅ All projects look valid")
	}
	
	fmt.Println()
}