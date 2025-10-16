package cmd

import (
	"fmt"
	"sort"
	"strings"

	"olive-clone-assistant-v2/internal"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	listGroups bool
	compact    bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "📋 List projects from inventory",
	Long: color.New(color.FgBlue, color.Bold).Sprint(`
📋 List Command  
================

Display projects and groups from your inventory file in a beautiful format.
Perfect for exploring your repository structure and understanding what's available.

• 📁 Show all groups and their projects
• 🔍 Filter and search capabilities
• 📊 Summary statistics
• 🎨 Colorized output for easy reading
`),
	Run: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listGroups, "groups-only", false, "Show only groups (no individual projects)")
	listCmd.Flags().BoolVar(&compact, "compact", false, "Use compact display format")
}

func runList(cmd *cobra.Command, args []string) {
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

	// Group projects by group
	groupMap := make(map[string][]internal.ProjectInfo)
	for _, project := range allProjects {
		groupMap[project.Group] = append(groupMap[project.Group], project)
	}

	// Sort groups
	var groups []string
	for group := range groupMap {
		groups = append(groups, group)
	}
	sort.Strings(groups)

	// Display summary
	logger.Header(fmt.Sprintf("📊 Inventory Summary (%s)", file))
	color.New(color.FgCyan, color.Bold).Printf("Total Projects: %d\n", len(allProjects))
	color.New(color.FgCyan, color.Bold).Printf("Total Groups: %d\n", len(groups))
	fmt.Println()

	if listGroups {
		// Show groups only
		showGroupsOnly(groups, groupMap, logger)
	} else {
		// Show detailed view
		showDetailedView(groups, groupMap, compact, logger)
	}
}

func showGroupsOnly(groups []string, groupMap map[string][]internal.ProjectInfo, logger *internal.Logger) {
	logger.Header("📁 Groups")
	
	for _, group := range groups {
		projects := groupMap[group]
		color.New(color.FgBlue, color.Bold).Printf("📁 %s", group)
		color.New(color.FgWhite).Printf(" (%d projects)\n", len(projects))
	}
}

func showDetailedView(groups []string, groupMap map[string][]internal.ProjectInfo, compact bool, logger *internal.Logger) {
	for i, group := range groups {
		projects := groupMap[group]
		
		// Group header
		color.New(color.FgBlue, color.Bold).Printf("📁 %s", group)
		color.New(color.FgWhite).Printf(" (%d projects)", len(projects))
		fmt.Println()
		
		if compact {
			showCompactProjects(projects)
		} else {
			showDetailedProjects(projects)
		}

		// Add separator between groups (except last one)
		if i < len(groups)-1 {
			fmt.Println()
			color.New(color.FgBlue).Println(strings.Repeat("─", 60))
			fmt.Println()
		}
	}
}

func showCompactProjects(projects []internal.ProjectInfo) {
	for _, project := range projects {
		color.New(color.FgGreen).Printf("  • %s", project.Name)
		if verbose {
			color.New(color.FgWhite, color.Faint).Printf(" (%s)", project.URL)
		}
		fmt.Println()
	}
}

func showDetailedProjects(projects []internal.ProjectInfo) {
	for _, project := range projects {
		color.New(color.FgGreen, color.Bold).Printf("  📦 %s\n", project.Name)
		color.New(color.FgWhite, color.Faint).Printf("      🔗 %s\n", project.URL)
		
		if verbose {
			// Show additional details in verbose mode
			gitURL := internal.FormatGitURL(project.URL, protocol)
			color.New(color.FgCyan, color.Faint).Printf("      🌐 Git URL: %s\n", gitURL)
			
			dirPath := internal.ExtractDirectoryPath(project.URL)
			if dirPath != "" {
				color.New(color.FgYellow, color.Faint).Printf("      📂 Path: %s\n", dirPath)
			}
		}
		fmt.Println()
	}
}