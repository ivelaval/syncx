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
	pullParallel int
	pullGroup    string
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "🔄 Pull updates for existing repositories only",
	Long: color.New(color.FgBlue, color.Bold).Sprint(`
🔄 Pull Only Mode
=================

Update existing repositories without cloning new ones.
This command will:

• 🔍 Scan for existing repositories in your physical location
• 🔄 Pull latest changes for found repositories  
• ⚡ Skip repositories that don't exist locally
• 📊 Show detailed progress and results

Perfect for updating your existing workspace without adding new repos.
`),
	Run: runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)

	pullCmd.Flags().IntVarP(&pullParallel, "parallel", "p", 3, "Number of parallel pull operations (1-10)")
	pullCmd.Flags().StringVarP(&pullGroup, "group", "g", "", "Pull only repositories from specific group")
}

func runPull(cmd *cobra.Command, args []string) {
	logger := internal.NewLogger(verbose)
	startTime := time.Now()

	// Show banner
	logger.Banner()

	if dryRun {
		color.New(color.FgYellow, color.Bold).Println("🔍 DRY RUN MODE - No actual operations will be performed")
		fmt.Println()
	}

	// Load inventory
	logger.Header("📋 Loading Project Inventory")
	inventory, err := internal.LoadInventory(file)
	if err != nil {
		logger.Error("Failed to load inventory: %v", err)
		return
	}

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
	if pullGroup != "" {
		filteredProjects := internal.FilterProjectsByGroup(allProjects, pullGroup)
		if len(filteredProjects) == 0 {
			logger.Warning("No projects found for group: %s", pullGroup)
			showAvailableGroups(allProjects, logger)
			return
		}
		allProjects = filteredProjects
		logger.Info("Filtered to %d projects in group: %s", len(allProjects), pullGroup)
	}

	// Ensure output directory exists and is valid
	absDir, err := internal.EnsureOutputDirectory(directory, logger)
	if err != nil {
		logger.Error("Output directory setup failed: %v", err)
		return
	}

	// Show output directory info if verbose
	if verbose {
		internal.ShowOutputDirectoryInfo(absDir, logger)
	}

	// Display configuration
	logger.Header("⚙️  Pull Configuration")
	color.New(color.FgCyan).Printf("   Physical Location: %s\n", absDir)
	color.New(color.FgCyan).Printf("   Projects to scan: %d\n", len(allProjects))
	color.New(color.FgCyan).Printf("   Parallel operations: %d\n", pullParallel)
	color.New(color.FgYellow).Printf("   Mode: Pull Only (existing repos only)\n")
	fmt.Println()

	// Scan for existing repositories only
	logger.Header("🔍 Scanning for Existing Repositories")
	var existingProjects []internal.ProjectInfo

	for _, project := range allProjects {
		if _, err := os.Stat(project.LocalPath); err == nil {
			if internal.IsGitRepository(project.LocalPath) {
				existingProjects = append(existingProjects, project)
				logger.Info("✓ Found: %s", project.Name)
			} else {
				logger.Warning("⚠️  Directory exists but not a git repo: %s", project.Name)
			}
		}
	}

	if len(existingProjects) == 0 {
		logger.Warning("No existing repositories found in %s", absDir)
		logger.Info("💡 Use 'clone' command to download repositories first")
		return
	}

	logger.Success("Found %d existing repositories to update", len(existingProjects))
	fmt.Println()

	// Process only existing projects
	summary := processPullOperations(existingProjects, logger)
	summary.TotalDuration = time.Since(startTime).String()

	// Show summary
	logger.Summary(summary)

	if dryRun {
		color.New(color.FgYellow, color.Bold).Println("\n💡 This was a dry run. Run without --dry-run to execute operations.")
	}
}

func processPullOperations(projects []internal.ProjectInfo, logger *internal.Logger) internal.Summary {
	totalProjects := len(projects)
	
	var results []internal.OperationResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create clean progress bar
	bar := progressbar.NewOptions(totalProjects,
		progressbar.OptionSetDescription("🔄 Pulling updates"),
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
	)

	// Create semaphore for parallel processing
	semaphore := make(chan struct{}, pullParallel)

	// Process function
	processProject := func(project internal.ProjectInfo) {
		defer wg.Done()
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		// For pull-only mode, we only do pull operations
		var result internal.OperationResult
		if dryRun {
			result = internal.OperationResult{
				Success:  true,
				Project:  project,
				Message:  fmt.Sprintf("DRY RUN: Would pull updates for %s", project.Name),
				IsClone:  false,
				Duration: "0s",
			}
		} else {
			result = internal.PullRepositorySilent(project.LocalPath)
			result.Project = project
		}
		
		mutex.Lock()
		results = append(results, result)
		bar.Describe(fmt.Sprintf("🔄 Updating: %s", project.Name))
		bar.Add(1)
		mutex.Unlock()
	}

	logger.Header("🔄 Pulling Updates")

	// Start pull operations
	for _, project := range projects {
		wg.Add(1)
		go processProject(project)
	}

	// Wait for all operations to complete
	wg.Wait()
	bar.Finish()
	fmt.Println()
	fmt.Println()

	// Show detailed results after progress bar completes
	logger.Header("📊 Pull Results")

	var successfulOps, failedOps, emptyOps []internal.OperationResult
	for _, result := range results {
		if result.Success {
			successfulOps = append(successfulOps, result)
		} else if result.IsEmpty {
			emptyOps = append(emptyOps, result)
		} else {
			failedOps = append(failedOps, result)
		}
	}

	// Show successful operations summary
	if len(successfulOps) > 0 {
		color.New(color.FgGreen, color.Bold).Printf("✅ Successfully Updated (%d):\n", len(successfulOps))
		for _, result := range successfulOps {
			color.New(color.FgGreen).Printf("   %s (%s)\n", result.Project.Name, result.Duration)
		}
		fmt.Println()
	}

	// Show empty repositories (less prominent color)
	if len(emptyOps) > 0 {
		color.New(color.FgYellow).Printf("⚠️  Empty Repositories (%d):\n", len(emptyOps))
		for _, result := range emptyOps {
			color.New(color.FgYellow).Printf("   📭 %s (no commits)\n", result.Project.Name)
		}
		fmt.Println()
	}

	// Show failed operations details
	if len(failedOps) > 0 {
		color.New(color.FgRed, color.Bold).Printf("❌ Failed Updates (%d):\n", len(failedOps))
		for _, result := range failedOps {
			color.New(color.FgRed).Printf("   %s: %s\n", result.Project.Name, result.Message)
		}
		fmt.Println()
	}

	// Calculate summary
	summary := internal.Summary{
		TotalProjects: totalProjects,
	}

	var failedProjects []internal.ProjectInfo
	var emptyProjects []internal.ProjectInfo

	for _, result := range results {
		if result.Success {
			summary.SuccessCount++
			summary.UpdatedCount++ // All operations in pull mode are updates
		} else if result.IsEmpty {
			summary.EmptyCount++
			emptyProjects = append(emptyProjects, result.Project)
		} else {
			summary.FailureCount++
			failedProjects = append(failedProjects, result.Project)
		}
	}

	summary.FailedProjects = failedProjects
	summary.EmptyProjects = emptyProjects
	return summary
}