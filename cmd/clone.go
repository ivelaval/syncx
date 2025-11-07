package cmd

import (
	"fmt"
	"sync"
	"time"

	"olive-clone-assistant-v2/internal"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	parallel        int
	groupFilter     string
	showGroups      bool
	checkRemote     bool
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "🚀 Clone and update repositories from inventory",
	Long: color.New(color.FgBlue, color.Bold).Sprint(`
🚀 Clone Command
================

Clone new repositories and update existing ones based on your inventory file.
This command intelligently scans your directory structure and:

• 📥 Clones missing repositories
• 🔄 Updates existing repositories  
• 📊 Shows detailed progress with beautiful visualizations
• ⚡ Supports parallel operations for faster execution
• 🎯 Allows filtering by groups for targeted operations
`),
	Run: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)

	cloneCmd.Flags().IntVarP(&parallel, "parallel", "p", 10, "Number of parallel operations (1-20)")
	cloneCmd.Flags().StringVarP(&groupFilter, "group", "g", "", "Filter projects by group name")
	cloneCmd.Flags().BoolVar(&showGroups, "show-groups", false, "Show available groups and exit")
	cloneCmd.Flags().BoolVar(&checkRemote, "check-remote", false, "Check remote for updates on existing repos (slower)")
}

func runClone(cmd *cobra.Command, args []string) {
	logger := internal.NewLogger(verbose)
	startTime := time.Now()

	// Show banner
	logger.Banner()

	if dryRun {
		color.New(color.FgYellow, color.Bold).Println("🔍 DRY RUN MODE - No actual operations will be performed")
		fmt.Println()
	}

	// Load inventory with spinner
	spinnerLoad := logger.StartSpinner(fmt.Sprintf("Loading inventory from %s", file))
	inventory, err := internal.LoadInventory(file)
	if err != nil {
		logger.StopSpinnerError(spinnerLoad, fmt.Sprintf("Failed to load inventory: %v", err))
		return
	}
	logger.StopSpinnerSuccess(spinnerLoad, "Inventory loaded successfully")

	// Collect all projects
	allProjects := internal.CollectAllProjects(*inventory)
	if len(allProjects) == 0 {
		logger.Warning("No projects found in inventory")
		return
	}

	color.New(color.FgGreen).Printf("✅ All projects look valid (%d projects)\n", len(allProjects))
	fmt.Println()

	// Show groups if requested
	if showGroups {
		showAvailableGroups(allProjects, logger)
		return
	}

	// Filter by group if specified
	if groupFilter != "" {
		filteredProjects := internal.FilterProjectsByGroup(allProjects, groupFilter)
		if len(filteredProjects) == 0 {
			logger.Warning("No projects found for group: %s", groupFilter)
			showAvailableGroups(allProjects, logger)
			return
		}
		allProjects = filteredProjects
		logger.Info("Filtered to %d projects in group: %s", len(allProjects), groupFilter)
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
	logger.Header("⚙️  Configuration")
	fmt.Println()
	color.New(color.FgWhite).Printf("   Protocol: %s\n", protocol)
	color.New(color.FgWhite).Printf("   Output Directory: %s\n", absDir)
	color.New(color.FgWhite).Printf("   Projects: %d\n", len(allProjects))
	color.New(color.FgWhite).Printf("   Parallel: %d\n", parallel)
	fmt.Println()

	// Use new smart tracking system with spinner
	var spinnerMessage string
	if checkRemote {
		spinnerMessage = "Analyzing project differences (checking remote for updates)..."
	} else {
		spinnerMessage = "Analyzing project differences (fast mode - local check only)..."
	}
	spinnerAnalysis := logger.StartSpinner(spinnerMessage)
	skipCheck := !checkRemote  // Invert logic: if NOT checking remote, then skip check
	projectsToClone, projectsToPull, projectsUpToDate, err := internal.ScanAndClassifyProjectsWithTrackingSkipCheck(allProjects, absDir, protocol, file, skipCheck, logger)
	if err != nil {
		logger.StopSpinnerWarning(spinnerAnalysis, "Smart tracking failed, using basic scan")
		// Fallback to old system
		projectsToClone, projectsToPull = internal.ScanAndClassifyProjects(allProjects, absDir, protocol, logger)
		projectsUpToDate = []internal.ProjectInfo{}
	} else {
		existingCount := len(projectsToPull) + len(projectsUpToDate)
		logger.StopSpinnerSuccess(spinnerAnalysis, fmt.Sprintf("Analysis complete: %d new, %d existing", len(projectsToClone), existingCount))
	}
	fmt.Println()

	// Clone mode: ONLY clone new projects (no pull)
	if len(projectsToClone) == 0 {
		existingCount := len(projectsToPull) + len(projectsUpToDate)
		logger.Success("✅ No new projects to clone. All %d projects already exist!", existingCount)
		logger.Info("💡 Use 'syncx pull' to update existing repositories")
		return
	}

	existingCount := len(projectsToPull) + len(projectsUpToDate)
	if existingCount > 0 {
		logger.Info("📦 Cloning %d new projects (%d already exist)", len(projectsToClone), existingCount)
	} else {
		logger.Info("📦 Cloning %d new projects", len(projectsToClone))
	}

	// Process ONLY new projects (clone only, no pull)
	summary := processCloneOnly(projectsToClone, logger)
	summary.TotalDuration = time.Since(startTime).String()

	// Show summary
	logger.Summary(summary)

	if dryRun {
		color.New(color.FgYellow, color.Bold).Println("\n💡 This was a dry run. Run without --dry-run to execute operations.")
	}
}


func processCloneOnly(projectsToClone []internal.ProjectInfo, logger *internal.Logger) internal.Summary {
	totalProjects := len(projectsToClone)

	var results []internal.OperationResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create clean progress bar with proper single-line rendering
	bar := progressbar.NewOptions(totalProjects,
		progressbar.OptionSetDescription("📥 Cloning new repositories"),
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
		progressbar.OptionUseANSICodes(true), // Force ANSI codes for proper single-line updates
	)

	// Create semaphore for parallel processing
	semaphore := make(chan struct{}, parallel)

	// Process function for cloning only
	processProject := func(project internal.ProjectInfo) {
		defer wg.Done()
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		var result internal.OperationResult
		if dryRun {
			result = internal.OperationResult{
				Success:  true,
				Project:  project,
				Message:  fmt.Sprintf("DRY RUN: Would clone %s", project.Name),
				IsClone:  true,
				Duration: "0s",
			}
		} else {
			result = internal.CloneRepositorySilent(project.GitURL, project.LocalPath)
			result.Project = project
		}

		mutex.Lock()
		results = append(results, result)
		bar.Add(1)
		mutex.Unlock()
	}

	// Start clone operations
	for _, project := range projectsToClone {
		wg.Add(1)
		go processProject(project)
	}

	// Wait for all operations to complete
	wg.Wait()
	bar.Finish()
	fmt.Println()

	// Show detailed results after progress bar completes
	logger.Header("📊 Clone Results")

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
		color.New(color.FgGreen, color.Bold).Printf("✅ Successfully Cloned (%d):\n", len(successfulOps))
		for _, result := range successfulOps {
			color.New(color.FgGreen).Printf("   %s (%s)\n", result.Project.Name, result.Duration)
		}
		fmt.Println()
	}

	// Show empty repositories
	if len(emptyOps) > 0 {
		color.New(color.FgYellow).Printf("⚠️  Empty Repositories (%d):\n", len(emptyOps))
		for _, result := range emptyOps {
			color.New(color.FgYellow).Printf("   📭 %s (no commits)\n", result.Project.Name)
		}
		fmt.Println()
	}

	// Show failed operations details
	if len(failedOps) > 0 {
		color.New(color.FgRed, color.Bold).Printf("❌ Failed Clones (%d):\n", len(failedOps))
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
			summary.ClonedCount++
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

func processProjects(projectsToClone, projectsToPull []internal.ProjectInfo, logger *internal.Logger) internal.Summary {
	totalProjects := len(projectsToClone) + len(projectsToPull)
	
	var results []internal.OperationResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create clean progress bar with better formatting
	bar := progressbar.NewOptions(totalProjects,
		progressbar.OptionSetDescription("🚀 Processing repositories"),
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
	semaphore := make(chan struct{}, parallel)
	
	// Create a silent logger for batch operations
	silentLogger := internal.NewLogger(false) // Non-verbose mode for clean progress

	// Process function
	processProject := func(project internal.ProjectInfo) {
		defer wg.Done()
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		// Use silent logger during batch processing to avoid cluttering output
		result := internal.CloneOrUpdateRepositorySilent(project, dryRun, silentLogger)
		
		mutex.Lock()
		results = append(results, result)
		// Update progress bar with current project info
		bar.Describe(fmt.Sprintf("🚀 Processing: %s", project.Name))
		bar.Add(1)
		mutex.Unlock()
	}

	// Start clone operations
	for _, project := range projectsToClone {
		wg.Add(1)
		go processProject(project)
	}

	// Start pull operations
	for _, project := range projectsToPull {
		wg.Add(1)
		go processProject(project)
	}

	// Wait for all operations to complete
	wg.Wait()
	bar.Finish()
	fmt.Println()
	fmt.Println()

	// Show detailed results after progress bar completes
	fmt.Println()
	logger.Header("📊 Operation Results")

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
		color.New(color.FgGreen, color.Bold).Printf("✅ Successful Operations (%d):\n", len(successfulOps))
		for _, result := range successfulOps {
			action := "Updated"
			if result.IsClone {
				action = "Cloned"
			}
			color.New(color.FgGreen).Printf("   %s %s (%s)\n", action, result.Project.Name, result.Duration)
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
		color.New(color.FgRed, color.Bold).Printf("❌ Failed Operations (%d):\n", len(failedOps))
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
			if result.IsClone {
				summary.ClonedCount++
			} else {
				summary.UpdatedCount++
			}
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


func showAvailableGroups(projects []internal.ProjectInfo, logger *internal.Logger) {
	groups := internal.GetUniqueGroups(projects)
	
	logger.Header("📁 Available Groups")
	
	for _, group := range groups {
		projectCount := len(internal.FilterProjectsByGroup(projects, group))
		color.New(color.FgCyan).Printf("  • %s (%d projects)\n", group, projectCount)
	}
	
	fmt.Println()
	color.New(color.FgYellow).Println("💡 Use --group <group-name> to filter by a specific group")
}