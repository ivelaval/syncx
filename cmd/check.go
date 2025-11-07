package cmd

import (
	"fmt"
	"os"
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

	// Load tracker to find actually cloned repositories
	spinnerScan := logger.StartSpinner("Scanning for existing repositories using tracker...")
	tracker, err := internal.LoadOrCreateTracker(absDir, file)
	if err != nil {
		logger.StopSpinnerWarning(spinnerScan, "Tracker not found, using inventory paths")

		// Fallback: scan using inventory paths
		var existingProjects []internal.ProjectInfo
		for _, project := range allProjects {
			if _, err := os.Stat(project.LocalPath); err == nil {
				if internal.IsGitRepository(project.LocalPath) {
					existingProjects = append(existingProjects, project)
				}
			}
		}

		if len(existingProjects) == 0 {
			logger.Warning("No existing repositories found in %s", absDir)
			logger.Info("💡 Use 'clone' command to download repositories first")
			return
		}

		// Process fallback
		checkResults := processCheckOperations(existingProjects, logger)
		displayCheckResults(checkResults, logger, time.Since(startTime).String())
		return
	}

	// Use tracker to find existing projects
	var existingProjects []internal.ProjectInfo
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

	logger.StopSpinnerSuccess(spinnerScan, fmt.Sprintf("Found %d tracked repositories", trackedCount))

	if len(existingProjects) == 0 {
		logger.Warning("No existing repositories found in %s", absDir)
		logger.Info("💡 Use 'clone' command to download repositories first")
		return
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
	Branch           string
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

	return result
}

func displayCheckResults(results []CheckResult, logger *internal.Logger, duration string) {
	// Group results
	var cleanRepos []CheckResult
	var modifiedRepos []CheckResult
	var stagedRepos []CheckResult
	var untrackedRepos []CheckResult
	var errorRepos []CheckResult

	for _, result := range results {
		if result.Error != "" {
			errorRepos = append(errorRepos, result)
		} else if !result.HasChanges {
			cleanRepos = append(cleanRepos, result)
		} else {
			if result.ModifiedFiles > 0 {
				modifiedRepos = append(modifiedRepos, result)
			}
			if result.StagedFiles > 0 {
				stagedRepos = append(stagedRepos, result)
			}
			if result.UntrackedFiles > 0 {
				untrackedRepos = append(untrackedRepos, result)
			}
		}
	}

	// Display summary
	logger.Header("📊 Check Results Summary")

	total := len(results)
	color.New(color.FgWhite, color.Bold).Printf("Total repositories scanned: %d\n", total)
	color.New(color.FgCyan).Printf("Duration: %s\n", duration)
	fmt.Println()

	if len(cleanRepos) > 0 {
		color.New(color.FgGreen, color.Bold).Printf("✅ Clean (no changes): %d\n", len(cleanRepos))
	}

	if len(modifiedRepos) > 0 {
		color.New(color.FgYellow, color.Bold).Printf("📝 Modified files: %d\n", len(modifiedRepos))
	}

	if len(stagedRepos) > 0 {
		color.New(color.FgBlue, color.Bold).Printf("📦 Staged changes: %d\n", len(stagedRepos))
	}

	if len(untrackedRepos) > 0 {
		color.New(color.FgMagenta, color.Bold).Printf("❓ Untracked files: %d\n", len(untrackedRepos))
	}

	if len(errorRepos) > 0 {
		color.New(color.FgRed, color.Bold).Printf("❌ Errors: %d\n", len(errorRepos))
	}

	// Show detailed information for repositories with changes
	if len(modifiedRepos) > 0 {
		fmt.Println()
		logger.Header("📝 Repositories with Modified Files")
		for _, result := range modifiedRepos {
			color.New(color.FgYellow).Printf("  • %s (%s)\n", result.Project.Name, result.Branch)
			color.New(color.FgWhite).Printf("    Modified: %d", result.ModifiedFiles)
			if result.StagedFiles > 0 {
				color.New(color.FgBlue).Printf(" | Staged: %d", result.StagedFiles)
			}
			if result.UntrackedFiles > 0 {
				color.New(color.FgMagenta).Printf(" | Untracked: %d", result.UntrackedFiles)
			}
			fmt.Println()
			if verbose {
				color.New(color.FgWhite, color.Faint).Printf("    Path: %s\n", result.Project.LocalPath)
			}
		}
	}

	if len(stagedRepos) > 0 && len(modifiedRepos) == 0 {
		fmt.Println()
		logger.Header("📦 Repositories with Staged Changes")
		for _, result := range stagedRepos {
			if result.ModifiedFiles == 0 { // Only show if not already shown in modified section
				color.New(color.FgBlue).Printf("  • %s (%s)\n", result.Project.Name, result.Branch)
				color.New(color.FgWhite).Printf("    Staged: %d", result.StagedFiles)
				if result.UntrackedFiles > 0 {
					color.New(color.FgMagenta).Printf(" | Untracked: %d", result.UntrackedFiles)
				}
				fmt.Println()
			}
		}
	}

	if len(errorRepos) > 0 {
		fmt.Println()
		logger.Header("❌ Repositories with Errors")
		for _, result := range errorRepos {
			color.New(color.FgRed).Printf("  • %s: %s\n", result.Project.Name, result.Error)
		}
	}

	if len(cleanRepos) > 0 && verbose {
		fmt.Println()
		logger.Header("✅ Clean Repositories")
		for _, result := range cleanRepos {
			color.New(color.FgGreen).Printf("  • %s (%s)\n", result.Project.Name, result.Branch)
		}
	}

	// Show helpful tips
	if len(modifiedRepos) > 0 || len(stagedRepos) > 0 {
		fmt.Println()
		logger.Info("💡 Tip: Review and commit your changes before running sync operations")
	}
}
