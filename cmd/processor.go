package cmd

import (
	"fmt"
	"sync"
	"time"

	"olive-clone-assistant-v2/internal"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// processProjectsWithTracking processes projects with the new tracking system
func processProjectsWithTracking(projectsToClone, projectsToPull, projectsUpToDate []internal.ProjectInfo, logger *internal.Logger) internal.Summary {
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

	logger.Header("🚀 Processing Repositories with Smart Tracking")

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
	
	// Show up-to-date projects
	if len(projectsUpToDate) > 0 {
		color.New(color.FgCyan, color.Bold).Printf("✅ Already Up-to-Date (%d):\n", len(projectsUpToDate))
		for _, project := range projectsUpToDate {
			color.New(color.FgCyan).Printf("   ✅ %s (no changes)\n", project.Name)
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

	// Calculate summary including up-to-date projects
	summary := internal.Summary{
		TotalProjects: totalProjects + len(projectsUpToDate),
		SkippedCount:  len(projectsUpToDate),
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