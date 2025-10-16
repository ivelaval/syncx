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
	interactive     bool
	parallel        int
	groupFilter     string
	showGroups      bool
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

	cloneCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Enable interactive mode for project selection")
	cloneCmd.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel operations (1-10)")
	cloneCmd.Flags().StringVarP(&groupFilter, "group", "g", "", "Filter projects by group name")
	cloneCmd.Flags().BoolVar(&showGroups, "show-groups", false, "Show available groups and exit")
	
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

	// Load inventory
	logger.Header("📋 Loading Project Inventory")
	inventory, err := internal.LoadInventory(file)
	if err != nil {
		logger.Error("Failed to load inventory: %v", err)
		return
	}

	// Validate and show inventory statistics
	internal.ValidateAndShowInventoryStats(inventory, logger)

	// Collect all projects
	allProjects := internal.CollectAllProjects(*inventory)
	if len(allProjects) == 0 {
		logger.Warning("No projects found in inventory")
		return
	}

	logger.Success("✅ Successfully loaded %d projects from %s", len(allProjects), file)

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

	// Interactive mode - Use the new wizard system
	if interactive {
		wizard := internal.NewInteractiveWizard(allProjects, logger)
		choice, err := wizard.RunWizard()
		if err != nil {
			logger.Error("Interactive wizard failed: %v", err)
			return
		}
		
		// Apply wizard choices
		allProjects = choice.SelectedProjects
		protocol = choice.Protocol
		if choice.Directory != "" {
			directory = choice.Directory
			// Re-validate the new directory from wizard
			absDir, err = internal.EnsureOutputDirectory(directory, logger)
			if err != nil {
				logger.Error("Wizard output directory setup failed: %v", err)
				return
			}
		}
		parallel = choice.Parallel
		dryRun = choice.DryRun
		verbose = choice.Verbose
		
		// Update logger with new verbosity
		logger = internal.NewLogger(verbose)
	}

	// Display configuration
	logger.Header("⚙️  Configuration")
	color.New(color.FgCyan).Printf("   Protocol: %s\n", protocol)
	color.New(color.FgCyan).Printf("   Output Directory: %s\n", absDir)
	color.New(color.FgCyan).Printf("   Projects: %d\n", len(allProjects))
	color.New(color.FgCyan).Printf("   Parallel: %d\n", parallel)
	fmt.Println()

	// Use new smart tracking system
	projectsToClone, projectsToPull, projectsUpToDate, err := internal.ScanAndClassifyProjectsWithTracking(allProjects, absDir, protocol, file, logger)
	if err != nil {
		logger.Error("Smart tracking failed, falling back to basic scan: %v", err)
		// Fallback to old system
		projectsToClone, projectsToPull = internal.ScanAndClassifyProjects(allProjects, absDir, protocol, logger)
		projectsUpToDate = []internal.ProjectInfo{}
	}

	// Process both clone and pull (smart tracking handles the logic)
	logger.Info("🎯 Smart tracking: Processing (%d clone + %d pull) repositories", len(projectsToClone), len(projectsToPull))

	totalProjects := len(projectsToClone) + len(projectsToPull)
	if totalProjects == 0 {
		if len(projectsUpToDate) > 0 {
			logger.Success("🎉 All %d projects are already up to date!", len(projectsUpToDate))
		} else {
			logger.Success("All projects are already up to date!")
		}
		return
	}

	// Process projects
	summary := processProjectsWithTracking(projectsToClone, projectsToPull, projectsUpToDate, logger)
	summary.TotalDuration = time.Since(startTime).String()

	// Show summary
	logger.Summary(summary)

	if dryRun {
		color.New(color.FgYellow, color.Bold).Println("\n💡 This was a dry run. Run without --dry-run to execute operations.")
	}
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

	logger.Header("🚀 Processing Repositories")

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