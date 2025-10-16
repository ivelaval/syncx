package cmd

import (
	"fmt"

	"olive-clone-assistant-v2/internal"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// wizardCmd represents the wizard command
var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "🧙‍♂️ Interactive wizard for guided repository management",
	Long: color.New(color.FgMagenta, color.Bold).Sprint(`
🧙‍♂️ Interactive Wizard
====================

Step-by-step guided experience for repository management.
Perfect for first-time users or when you want full control over every option.

The wizard offers three modes:
• 🚀 Quick Mode - Smart defaults for rapid setup
• 🎯 Custom Mode - Select specific projects and groups
• ⚙️  Advanced Mode - Full control over all settings

Inspired by modern CLI tools like gitcook, this wizard makes complex 
repository operations simple and intuitive.
`),
	Run: runWizard,
}

func init() {
	rootCmd.AddCommand(wizardCmd)
}

func runWizard(cmd *cobra.Command, args []string) {
	logger := internal.NewLogger(verbose)

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

	// Run the wizard with inventory context
	wizard := internal.NewInteractiveWizardWithInventory(inventory, file, allProjects, logger)
	choice, err := wizard.RunWizard()
	if err != nil {
		logger.Error("Wizard failed: %v", err)
		return
	}

	// Show final summary
	logger.Header("🎯 Wizard Complete!")
	color.New(color.FgGreen, color.Bold).Println("✅ Configuration completed successfully")
	
	color.New(color.FgWhite).Printf("Selected %d projects\n", len(choice.SelectedProjects))
	color.New(color.FgWhite).Printf("Protocol: %s\n", choice.Protocol)
	color.New(color.FgWhite).Printf("Directory: %s\n", choice.Directory)
	color.New(color.FgWhite).Printf("Parallel: %d\n", choice.Parallel)
	
	if choice.DryRun {
		color.New(color.FgYellow).Println("Mode: Dry Run (preview only)")
	} else {
		color.New(color.FgGreen).Println("Mode: Execute operations")
	}

	// Execute the actual operation
	logger.Header("🚀 Executing Operations")
	
	// Set global variables from wizard choices
	protocol = choice.Protocol
	directory = choice.Directory
	parallel = choice.Parallel
	dryRun = choice.DryRun
	verbose = choice.Verbose
	
	// Update logger with chosen verbosity
	logger = internal.NewLogger(verbose)
	
	// Call the main clone logic with the selected projects
	executeCloneOperation(choice.SelectedProjects, logger)
}

func executeCloneOperation(selectedProjects []internal.ProjectInfo, logger *internal.Logger) {
	if dryRun {
		color.New(color.FgYellow, color.Bold).Println("🔍 DRY RUN MODE - No actual operations will be performed")
		fmt.Println()
	}

	// Ensure output directory exists
	absDir, err := internal.EnsureOutputDirectory(directory, logger)
	if err != nil {
		logger.Error("Output directory setup failed: %v", err)
		return
	}

	// Use new smart tracking system
	projectsToClone, projectsToPull, projectsUpToDate, err := internal.ScanAndClassifyProjectsWithTracking(selectedProjects, absDir, protocol, file, logger)
	if err != nil {
		logger.Error("Smart tracking failed, falling back to basic scan: %v", err)
		// Fallback to old system
		projectsToClone, projectsToPull = internal.ScanAndClassifyProjects(selectedProjects, absDir, protocol, logger)
		projectsUpToDate = []internal.ProjectInfo{}
	}

	totalProjects := len(projectsToClone) + len(projectsToPull)
	if totalProjects == 0 {
		if len(projectsUpToDate) > 0 {
			logger.Success("🎉 All %d selected projects are already up to date!", len(projectsUpToDate))
		} else {
			logger.Success("All projects are already up to date!")
		}
		return
	}

	// Process projects
	summary := processProjectsWithTracking(projectsToClone, projectsToPull, projectsUpToDate, logger)

	// Show summary
	logger.Summary(summary)

	if dryRun {
		color.New(color.FgYellow, color.Bold).Println("\n💡 This was a dry run. Run the wizard again and choose 'Execute' to perform operations.")
	}
}