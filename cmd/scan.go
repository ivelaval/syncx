package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"olive-clone-assistant-v2/internal"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	scanParallel int
	scanMaxDepth int
	scanShowClean bool
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [directory]",
	Short: "🔎 Recursively scan directory for git repositories with changes",
	Long: color.New(color.FgBlue, color.Bold).Sprint(`
🔎 Recursive Repository Scanner
================================

Recursively scan a directory tree for git repositories and check for uncommitted changes.
This command does NOT require an inventory file - it discovers repositories automatically.

Features:
• 🔍 Automatic git repository detection
• 📁 Recursive directory scanning
• 📝 Detects modified, staged, and untracked files
• ⚡ Fast parallel processing
• 🎯 No inventory file needed

Perfect for checking an entire workspace or drive for uncommitted changes.
`),
	Args: cobra.MaximumNArgs(1),
	Run: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().IntVarP(&scanParallel, "parallel", "p", 10, "Number of parallel scan operations (1-20)")
	scanCmd.Flags().IntVarP(&scanMaxDepth, "max-depth", "d", 10, "Maximum directory depth to scan (default: 10)")
	scanCmd.Flags().BoolVarP(&scanShowClean, "show-clean", "c", false, "Show clean repositories in results")
}

func runScan(cmd *cobra.Command, args []string) {
	logger := internal.NewLogger(verbose)
	startTime := time.Now()

	// Show banner
	logger.Banner()

	// Determine scan directory
	scanDir := "."
	if len(args) > 0 {
		scanDir = args[0]
	}

	// Get absolute path
	absDir, err := filepath.Abs(scanDir)
	if err != nil {
		logger.Error("Failed to get absolute path for %s: %v", scanDir, err)
		return
	}

	// Verify directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		logger.Error("Directory does not exist: %s", absDir)
		return
	}

	// Display configuration
	logger.Header("⚙️  Scan Configuration")
	color.New(color.FgCyan).Printf("   Scan directory: %s\n", absDir)
	color.New(color.FgCyan).Printf("   Max depth: %d\n", scanMaxDepth)
	color.New(color.FgCyan).Printf("   Parallel operations: %d\n", scanParallel)
	color.New(color.FgCyan).Printf("   Show clean repos: %v\n", scanShowClean)
	fmt.Println()

	// Discover git repositories
	logger.Header("🔍 Discovering Git Repositories")
	spinner := logger.StartSpinner("Scanning directory tree...")

	repositories := discoverGitRepositories(absDir, scanMaxDepth)

	logger.StopSpinnerSuccess(spinner, fmt.Sprintf("Found %d git repositories", len(repositories)))

	if len(repositories) == 0 {
		logger.Warning("No git repositories found in %s", absDir)
		return
	}

	// Check repositories for changes
	logger.Header("📝 Checking for Uncommitted Changes")
	results := scanRepositoriesForChanges(repositories, logger)

	// Display results
	displayScanResults(results, logger, time.Since(startTime).String())
}

// discoverGitRepositories recursively finds all git repositories in a directory
func discoverGitRepositories(rootDir string, maxDepth int) []string {
	var repositories []string
	var mu sync.Mutex

	var walkDir func(path string, depth int)
	walkDir = func(path string, depth int) {
		// Stop if we've reached max depth
		if depth > maxDepth {
			return
		}

		// Read directory contents
		entries, err := os.ReadDir(path)
		if err != nil {
			return // Skip directories we can't read
		}

		// Check if this directory is a git repository
		gitPath := filepath.Join(path, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			mu.Lock()
			repositories = append(repositories, path)
			mu.Unlock()
			return // Don't scan inside git repositories
		}

		// Recursively scan subdirectories
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			// Skip common non-repository directories
			name := entry.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" ||
			   name == ".venv" || name == "venv" || name == "__pycache__" ||
			   name == ".next" || name == ".nuxt" || name == "dist" || name == "build" {
				continue
			}

			subPath := filepath.Join(path, name)
			walkDir(subPath, depth+1)
		}
	}

	walkDir(rootDir, 0)
	return repositories
}

// ScanResult represents the result of scanning a repository
type ScanResult struct {
	Path             string
	Name             string
	HasChanges       bool
	ModifiedFiles    int
	StagedFiles      int
	UntrackedFiles   int
	Branch           string
	Error            string
}

// scanRepositoriesForChanges checks all repositories for uncommitted changes
func scanRepositoriesForChanges(repositories []string, logger *internal.Logger) []ScanResult {
	var results []ScanResult
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Create progress bar
	bar := progressbar.NewOptions(len(repositories),
		progressbar.OptionSetDescription("🔎 Scanning repositories"),
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
	semaphore := make(chan struct{}, scanParallel)

	// Process function
	processRepo := func(repoPath string) {
		defer wg.Done()
		semaphore <- struct{}{}
		defer func() { <-semaphore }()

		result := scanRepository(repoPath)

		mutex.Lock()
		results = append(results, result)
		bar.Add(1)
		mutex.Unlock()
	}

	// Start scanning
	for _, repoPath := range repositories {
		wg.Add(1)
		go processRepo(repoPath)
	}

	// Wait for all operations to complete
	wg.Wait()
	bar.Finish()
	fmt.Println()

	return results
}

// scanRepository checks a single repository for changes
func scanRepository(repoPath string) ScanResult {
	result := ScanResult{
		Path: repoPath,
		Name: filepath.Base(repoPath),
	}

	// Get current branch
	branch, err := internal.GetGitBranch(repoPath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get branch: %v", err)
		return result
	}
	result.Branch = branch

	// Check for changes
	modified, staged, untracked, err := internal.CheckRepositoryChanges(repoPath)
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

// displayScanResults shows the scan results in a formatted way
func displayScanResults(results []ScanResult, logger *internal.Logger, duration string) {
	// Group results
	var cleanRepos []ScanResult
	var modifiedRepos []ScanResult
	var stagedRepos []ScanResult
	var untrackedRepos []ScanResult
	var errorRepos []ScanResult

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
	logger.Header("📊 Scan Results Summary")

	total := len(results)
	color.New(color.FgWhite, color.Bold).Printf("Total repositories scanned: %d\n", total)
	color.New(color.FgCyan).Printf("Scan duration: %s\n", duration)
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
			color.New(color.FgYellow).Printf("  • %s (%s)\n", result.Name, result.Branch)
			color.New(color.FgWhite).Printf("    Modified: %d", result.ModifiedFiles)
			if result.StagedFiles > 0 {
				color.New(color.FgBlue).Printf(" | Staged: %d", result.StagedFiles)
			}
			if result.UntrackedFiles > 0 {
				color.New(color.FgMagenta).Printf(" | Untracked: %d", result.UntrackedFiles)
			}
			fmt.Println()
			if verbose {
				color.New(color.FgWhite, color.Faint).Printf("    Path: %s\n", result.Path)
			}
		}
	}

	if len(stagedRepos) > 0 && len(modifiedRepos) == 0 {
		fmt.Println()
		logger.Header("📦 Repositories with Staged Changes")
		for _, result := range stagedRepos {
			if result.ModifiedFiles == 0 { // Only show if not already shown in modified section
				color.New(color.FgBlue).Printf("  • %s (%s)\n", result.Name, result.Branch)
				color.New(color.FgWhite).Printf("    Staged: %d", result.StagedFiles)
				if result.UntrackedFiles > 0 {
					color.New(color.FgMagenta).Printf(" | Untracked: %d", result.UntrackedFiles)
				}
				fmt.Println()
				if verbose {
					color.New(color.FgWhite, color.Faint).Printf("    Path: %s\n", result.Path)
				}
			}
		}
	}

	if len(errorRepos) > 0 {
		fmt.Println()
		logger.Header("❌ Repositories with Errors")
		for _, result := range errorRepos {
			color.New(color.FgRed).Printf("  • %s: %s\n", result.Name, result.Error)
			if verbose {
				color.New(color.FgWhite, color.Faint).Printf("    Path: %s\n", result.Path)
			}
		}
	}

	if scanShowClean && len(cleanRepos) > 0 {
		fmt.Println()
		logger.Header("✅ Clean Repositories")
		for _, result := range cleanRepos {
			color.New(color.FgGreen).Printf("  • %s (%s)\n", result.Name, result.Branch)
			if verbose {
				color.New(color.FgWhite, color.Faint).Printf("    Path: %s\n", result.Path)
			}
		}
	}

	// Show helpful tips
	if len(modifiedRepos) > 0 || len(stagedRepos) > 0 {
		fmt.Println()
		logger.Info("💡 Tip: Review and commit your changes before syncing")
		logger.Info("💡 Use --verbose or -v to see full paths")
	}
}
