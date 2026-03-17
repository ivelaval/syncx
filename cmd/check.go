package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"olive-clone-assistant-v2/internal"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	checkParallel int
	checkGroup    string
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "🔍 Check for local uncommitted changes in repositories",
	Long: color.New(color.FgBlue, color.Bold).Sprint(`
🔍 Check for Local Changes
===========================

Scan all repositories in your inventory for uncommitted changes.
This command will:

• 📁 Identify repositories with modified files
• 📝 Show repositories with uncommitted changes
• 🔄 Display repositories with staged changes
• 📊 Provide a summary of repository states

Perfect for checking what needs to be committed before syncing.
`),
	Run: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)

	checkCmd.Flags().IntVarP(&checkParallel, "parallel", "p", 10, "Number of parallel check operations (1-20)")
	checkCmd.Flags().StringVarP(&checkGroup, "group", "g", "", "Check only repositories from specific group")
}

func runCheck(cmd *cobra.Command, args []string) {
	logger := internal.NewLogger(verbose)
	startTime := time.Now()

	// Show banner
	logger.Banner()

	// Load inventory with spinner
	spinnerLoad := logger.StartSpinner(fmt.Sprintf("Loading inventory from %s", file))
	inventory, err := internal.LoadInventory(file)
	if err != nil {
		logger.StopSpinnerError(spinnerLoad, fmt.Sprintf("Failed to load inventory: %v", err))
		return
	}
	logger.StopSpinnerSuccess(spinnerLoad, "Inventory loaded successfully")

	// Show physical location info
	if inventory.PhysicalLocation != "" {
		logger.Info("📍 Physical Location: %s", inventory.PhysicalLocation)
		// Use physical location as default directory if not specified via flags
		if directory == "" && inventory.PhysicalLocation != "" {
			directory = inventory.PhysicalLocation
		}
	}

	// Collect all projects
	allProjects := internal.CollectAllProjects(*inventory)
	if len(allProjects) == 0 {
		logger.Warning("No projects found in inventory")
		return
	}

	logger.Success("Loaded %d projects from inventory", len(allProjects))

	// Filter by group if specified
	if checkGroup != "" {
		filteredProjects := internal.FilterProjectsByGroup(allProjects, checkGroup)
		if len(filteredProjects) == 0 {
			logger.Warning("No projects found for group: %s", checkGroup)
			return
		}
		allProjects = filteredProjects
		logger.Info("Filtered to %d projects in group: %s", len(allProjects), checkGroup)
	}

	// Ensure output directory exists and is valid
	absDir, err := internal.EnsureOutputDirectory(directory, logger)
	if err != nil {
		logger.Error("Output directory setup failed: %v", err)
		return
	}

	// Display configuration
	logger.Header("⚙️  Check Configuration")
	color.New(color.FgCyan).Printf("   Physical Location: %s\n", absDir)
	color.New(color.FgCyan).Printf("   Projects to scan: %d\n", len(allProjects))
	color.New(color.FgCyan).Printf("   Parallel operations: %d\n", checkParallel)
	fmt.Println()

	// Try to load tracker first (silently)
	tracker, err := internal.LoadOrCreateTracker(absDir, file)

	var existingProjects []internal.ProjectInfo

	if err == nil {
		// Tracker exists, try to use it
		spinnerTracker := logger.StartSpinner("Looking for tracked repositories...")
		trackedCount := 0

		for _, project := range allProjects {
			// Find project in tracker by URL
			for _, trackedProject := range tracker.Projects {
				if trackedProject.URL == project.URL {
					// Use the tracked local path (actual location)
					project.LocalPath = trackedProject.LocalPath

					// Verify it still exists
					if _, err := os.Stat(project.LocalPath); err == nil {
						if internal.IsGitRepository(project.LocalPath) {
							existingProjects = append(existingProjects, project)
							trackedCount++
						}
					}
					break
				}
			}
		}

		if trackedCount > 0 {
			logger.StopSpinnerSuccess(spinnerTracker, fmt.Sprintf("Found %d tracked repositories", trackedCount))
		} else {
			logger.StopSpinnerWarning(spinnerTracker, "No tracked repositories found")
		}
	}

	// If tracker doesn't exist or found no repositories, scan the directory recursively
	if err != nil || len(existingProjects) == 0 {
		// Discover all git repositories recursively
		spinnerScan := logger.StartSpinner("Scanning directory for git repositories...")
		repositories := internal.DiscoverGitRepositories(absDir, 10) // Use depth 10 like scan command
		logger.StopSpinnerSuccess(spinnerScan, fmt.Sprintf("Found %d git repositories", len(repositories)))

		if len(repositories) == 0 {
			logger.Warning("No git repositories found in %s", absDir)
			logger.Info("💡 Use 'clone' command to download repositories first")
			return
		}

		// Convert discovered repositories to ProjectInfo
		existingProjects = make([]internal.ProjectInfo, 0, len(repositories))
		for _, repoPath := range repositories {
			projectInfo := internal.ProjectInfo{
				Name:      filepath.Base(repoPath),
				LocalPath: repoPath,
			}
			existingProjects = append(existingProjects, projectInfo)
		}
	}

	// Process check operations
	checkResults := processCheckOperations(existingProjects, logger)
	displayCheckResults(checkResults, logger, time.Since(startTime).String())
}

type CheckResult struct {
	Project          internal.ProjectInfo
	HasChanges       bool
	ModifiedFiles    int
	StagedFiles      int
	UntrackedFiles   int
	UnpushedCommits  int
	Branch           string
	ChangeSummary    []string
	Error            string
}

func processCheckOperations(projects []internal.ProjectInfo, logger *internal.Logger) []CheckResult {
	totalProjects := len(projects)

	var results []CheckResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create clean progress bar
	bar := progressbar.NewOptions(totalProjects,
		progressbar.OptionSetDescription("🔍 Checking for changes"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("repos"),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionClearOnFinish(),
	)

	// Create semaphore for parallel processing
	semaphore := make(chan struct{}, checkParallel)

	// Process function
	processProject := func(project internal.ProjectInfo) {
		defer wg.Done()
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		result := checkRepositoryChanges(project)

		mutex.Lock()
		results = append(results, result)
		bar.Add(1)
		mutex.Unlock()
	}

	// Start check operations
	for _, project := range projects {
		wg.Add(1)
		go processProject(project)
	}

	// Wait for all operations to complete
	wg.Wait()
	bar.Finish()
	fmt.Println()

	return results
}

func checkRepositoryChanges(project internal.ProjectInfo) CheckResult {
	result := CheckResult{
		Project: project,
	}

	// Check if it's a valid git repository
	if !internal.IsGitRepository(project.LocalPath) {
		result.Error = "Not a git repository"
		return result
	}

	// Get current branch
	branch, err := internal.GetGitBranch(project.LocalPath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get branch: %v", err)
		return result
	}
	result.Branch = branch

	// Check for changes
	modified, staged, untracked, err := internal.CheckRepositoryChanges(project.LocalPath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to check changes: %v", err)
		return result
	}

	result.ModifiedFiles = modified
	result.StagedFiles = staged
	result.UntrackedFiles = untracked
	result.HasChanges = (modified > 0) || (staged > 0) || (untracked > 0)

	// Check for unpushed commits
	unpushed, err := internal.CheckUnpushedCommits(project.LocalPath)
	if err == nil {
		result.UnpushedCommits = unpushed
	}

	// Get change summary
	summary, err := internal.GetChangeSummary(project.LocalPath)
	if err == nil {
		result.ChangeSummary = summary
	}

	return result
}

func displayCheckResults(results []CheckResult, logger *internal.Logger, duration string) {
	// Count statistics
	total := len(results)
	withUncommittedChanges := 0
	withUnpushedCommits := 0
	var reposWithUncommittedChanges []CheckResult
	var reposWithUnpushedCommits []CheckResult

	for _, result := range results {
		if result.Error != "" {
			continue
		}

		// Count repos with uncommitted changes (modified, staged, or untracked files)
		if result.HasChanges {
			withUncommittedChanges++
			reposWithUncommittedChanges = append(reposWithUncommittedChanges, result)
		}

		// Count repos with unpushed commits
		if result.UnpushedCommits > 0 {
			withUnpushedCommits++
			// Only add to list if not already in uncommitted changes list
			alreadyListed := false
			for _, r := range reposWithUncommittedChanges {
				if r.Project.LocalPath == result.Project.LocalPath {
					alreadyListed = true
					break
				}
			}
			if !alreadyListed {
				reposWithUnpushedCommits = append(reposWithUnpushedCommits, result)
			}
		}
	}

	// Display summary
	logger.Header("📊 Repository Summary")
	fmt.Println()

	color.New(color.FgWhite, color.Bold).Printf("  Total projects: %d\n", total)
	color.New(color.FgYellow, color.Bold).Printf("  With uncommitted changes: %d\n", withUncommittedChanges)
	color.New(color.FgBlue, color.Bold).Printf("  With unpushed commits: %d\n", withUnpushedCommits)
	fmt.Println()

	// Show repositories with uncommitted changes
	if len(reposWithUncommittedChanges) > 0 {
		logger.Header("📝 Repositories with Uncommitted Changes")
		fmt.Println()

		for _, result := range reposWithUncommittedChanges {
			// Show path
			color.New(color.FgYellow, color.Bold).Printf("  %s\n", result.Project.LocalPath)

			// Show summary
			if len(result.ChangeSummary) > 0 {
				color.New(color.FgWhite).Printf("    └─ %s", strings.Join(result.ChangeSummary, ", "))
			}

			// Show unpushed commits if any
			if result.UnpushedCommits > 0 {
				color.New(color.FgCyan).Printf(" | %d unpushed commits", result.UnpushedCommits)
			}

			fmt.Println()
			fmt.Println()
		}
	}

	// Show repositories with unpushed commits (but no uncommitted changes)
	if len(reposWithUnpushedCommits) > 0 {
		logger.Header("📤 Repositories with Unpushed Commits")
		fmt.Println()

		for _, result := range reposWithUnpushedCommits {
			color.New(color.FgBlue, color.Bold).Printf("  %s\n", result.Project.LocalPath)
			color.New(color.FgWhite).Printf("    └─ %d commits pending push\n", result.UnpushedCommits)
			fmt.Println()
		}
	}

	// Show message if everything is clean
	if withUncommittedChanges == 0 && withUnpushedCommits == 0 {
		color.New(color.FgGreen, color.Bold).Printf("\n  ✅ All repositories are clean and synced\n\n")
	}
}
