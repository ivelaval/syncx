package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"olive-clone-assistant-v2/internal"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "📊 Check status of repositories",
	Long: color.New(color.FgBlue, color.Bold).Sprint(`
📊 Status Command
=================

Check the current status of all repositories in your inventory.
This command provides detailed information about:

• 📁 Which repositories are cloned vs missing
• 🔄 Which repositories have uncommitted changes
• 📈 Which repositories are ahead/behind remote
• 🎯 Overall health of your repository collection
`),
	Run: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

type RepoStatus struct {
	Project     internal.ProjectInfo
	Exists      bool
	IsGitRepo   bool
	IsClean     bool
	Branch      string
	Ahead       int
	Behind      int
	Uncommitted int
	Error       string
}

func runStatus(cmd *cobra.Command, args []string) {
	logger := internal.NewLogger(verbose)

	// Show banner
	logger.Banner()

	// Load inventory
	logger.Header("📋 Loading Project Inventory")
	inventory, err := internal.LoadInventory(file)
	if err != nil {
		logger.Error("Failed to load inventory: %v", err)
		return
	}

	// Collect all projects
	allProjects := internal.CollectAllProjects(*inventory)
	if len(allProjects) == 0 {
		logger.Warning("No projects found in inventory")
		return
	}

	// Filter by group if specified
	if groupFilter != "" {
		filteredProjects := internal.FilterProjectsByGroup(allProjects, groupFilter)
		if len(filteredProjects) == 0 {
			logger.Warning("No projects found for group: %s", groupFilter)
			return
		}
		allProjects = filteredProjects
	}

	// Get absolute path for directory (don't create it for status check)
	absDir, err := filepath.Abs(directory)
	if err != nil {
		logger.Error("Failed to get absolute path for %s: %v", directory, err)
		return
	}

	logger.Header("🔍 Checking Repository Status")
	color.New(color.FgCyan).Printf("Output directory: %s\n", absDir)
	color.New(color.FgCyan).Printf("Checking %d repositories...\n", len(allProjects))
	fmt.Println()

	// Prepare projects with full paths
	var projectsWithPaths []internal.ProjectInfo
	for _, project := range allProjects {
		gitURL := internal.FormatGitURL(project.URL, protocol)
		dirPath := internal.ExtractDirectoryPath(project.URL)
		localPath := filepath.Join(absDir, dirPath)

		projectInfo := internal.ProjectInfo{
			Name:      project.Name,
			URL:       project.URL,
			GitURL:    gitURL,
			LocalPath: localPath,
			Group:     project.Group,
		}
		projectsWithPaths = append(projectsWithPaths, projectInfo)
	}

	// Check status of all repositories
	statuses := checkAllRepositories(projectsWithPaths, logger)

	// Display results
	displayStatusResults(statuses, logger)
}

func checkAllRepositories(projects []internal.ProjectInfo, logger *internal.Logger) []RepoStatus {
	var statuses []RepoStatus

	// Create progress bar
	bar := progressbar.NewOptions(len(projects),
		progressbar.OptionSetDescription("Checking status..."),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "🟢",
			SaucerHead:    "🟡",
			SaucerPadding: "⚪",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	for _, project := range projects {
		status := checkRepositoryStatus(project)
		statuses = append(statuses, status)
		bar.Add(1)
	}

	bar.Finish()
	fmt.Println()

	return statuses
}

func checkRepositoryStatus(project internal.ProjectInfo) RepoStatus {
	status := RepoStatus{
		Project: project,
	}

	// Check if directory exists
	if _, err := os.Stat(project.LocalPath); os.IsNotExist(err) {
		return status
	}
	status.Exists = true

	// Check if it's a git repository
	if !internal.IsGitRepository(project.LocalPath) {
		return status
	}
	status.IsGitRepo = true

	// Get current branch
	if branch, err := getGitBranch(project.LocalPath); err == nil {
		status.Branch = branch
	}

	// Check if working directory is clean
	if isClean, uncommitted := isWorkingDirectoryClean(project.LocalPath); isClean {
		status.IsClean = true
	} else {
		status.Uncommitted = uncommitted
	}

	// Check ahead/behind status
	if ahead, behind, err := getAheadBehindCount(project.LocalPath); err == nil {
		status.Ahead = ahead
		status.Behind = behind
	}

	return status
}

func getGitBranch(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func isWorkingDirectoryClean(path string) (bool, int) {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, 0
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return true, 0
	}
	
	return false, len(lines)
}

func getAheadBehindCount(path string) (int, int, error) {
	cmd := exec.Command("git", "-C", path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	
	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected output format")
	}
	
	var ahead, behind int
	fmt.Sscanf(parts[0], "%d", &ahead)
	fmt.Sscanf(parts[1], "%d", &behind)
	
	return ahead, behind, nil
}

func displayStatusResults(statuses []RepoStatus, logger *internal.Logger) {
	// Group by status
	var missing []RepoStatus
	var notGitRepos []RepoStatus
	var clean []RepoStatus
	var dirty []RepoStatus
	var needsPull []RepoStatus
	var needsPush []RepoStatus

	for _, status := range statuses {
		if !status.Exists {
			missing = append(missing, status)
		} else if !status.IsGitRepo {
			notGitRepos = append(notGitRepos, status)
		} else if !status.IsClean {
			dirty = append(dirty, status)
		} else if status.Behind > 0 {
			needsPull = append(needsPull, status)
		} else if status.Ahead > 0 {
			needsPush = append(needsPush, status)
		} else {
			clean = append(clean, status)
		}
	}

	// Display summary
	logger.Header("📊 Status Summary")
	
	total := len(statuses)
	color.New(color.FgWhite, color.Bold).Printf("Total repositories: %d\n", total)
	
	if len(clean) > 0 {
		color.New(color.FgGreen, color.Bold).Printf("✅ Clean: %d\n", len(clean))
	}
	
	if len(missing) > 0 {
		color.New(color.FgRed, color.Bold).Printf("❌ Missing: %d\n", len(missing))
	}
	
	if len(dirty) > 0 {
		color.New(color.FgYellow, color.Bold).Printf("📝 Uncommitted changes: %d\n", len(dirty))
	}
	
	if len(needsPull) > 0 {
		color.New(color.FgBlue, color.Bold).Printf("⬇️  Need pull: %d\n", len(needsPull))
	}
	
	if len(needsPush) > 0 {
		color.New(color.FgMagenta, color.Bold).Printf("⬆️  Need push: %d\n", len(needsPush))
	}
	
	if len(notGitRepos) > 0 {
		color.New(color.FgRed, color.Bold).Printf("⚠️  Not git repos: %d\n", len(notGitRepos))
	}

	// Show detailed information for problematic repos
	if len(missing) > 0 {
		fmt.Println()
		logger.Header("❌ Missing Repositories")
		for _, status := range missing {
			color.New(color.FgRed).Printf("  • %s (%s)\n", status.Project.Name, status.Project.Group)
			if verbose {
				color.New(color.FgWhite, color.Faint).Printf("    Expected at: %s\n", status.Project.LocalPath)
			}
		}
	}

	if len(dirty) > 0 {
		fmt.Println()
		logger.Header("📝 Repositories with Uncommitted Changes")
		for _, status := range dirty {
			color.New(color.FgYellow).Printf("  • %s (%d files) - %s\n", 
				status.Project.Name, status.Uncommitted, status.Branch)
			if verbose {
				color.New(color.FgWhite, color.Faint).Printf("    Path: %s\n", status.Project.LocalPath)
			}
		}
	}

	if len(needsPull) > 0 {
		fmt.Println()
		logger.Header("⬇️  Repositories Behind Remote")
		for _, status := range needsPull {
			color.New(color.FgBlue).Printf("  • %s (%d commits behind) - %s\n", 
				status.Project.Name, status.Behind, status.Branch)
		}
	}

	if len(needsPush) > 0 {
		fmt.Println()
		logger.Header("⬆️  Repositories Ahead of Remote")
		for _, status := range needsPush {
			color.New(color.FgMagenta).Printf("  • %s (%d commits ahead) - %s\n", 
				status.Project.Name, status.Ahead, status.Branch)
		}
	}

	if len(clean) > 0 && verbose {
		fmt.Println()
		logger.Header("✅ Clean Repositories")
		for _, status := range clean {
			color.New(color.FgGreen).Printf("  • %s - %s\n", status.Project.Name, status.Branch)
		}
	}
}