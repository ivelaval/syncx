package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// Version information (injected at build time)
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"

	// Global flags
	cfgFile   string
	verbose   bool
	dryRun    bool
	protocol  string
	directory string
	file      string
	outputDir string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "syncx",
	Short: "⚡ A powerful repository synchronization assistant",
	Long: color.New(color.FgCyan, color.Bold).Sprint(`
⚡ SyncX - Repository Synchronization Assistant
════════════════════════════════════════════════

A modern, interactive repository synchronization tool for managing multiple Git projects.
Features include:

• 🎨 Beautiful colored output and progress bars
• 🔄 Smart clone/update detection
• 📊 Intelligent repository tracking
• ⚡ Fast parallel operations
• 🔧 Flexible configuration options
• 📈 Detailed progress tracking

Use 'syncx [command] --help' for more information about a command.`),
	Version: fmt.Sprintf("%s (built: %s, commit: %s)", Version, BuildTime, GitCommit),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Setup global configuration
		setupGlobals(cmd)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&protocol, "protocol", "ssh", "Protocol to use for cloning (ssh or http)")
	rootCmd.PersistentFlags().StringVar(&file, "file", "projects-inventory.json", "Path to projects inventory JSON file")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "Output directory for cloning repositories (default: ../repositories)")
	rootCmd.PersistentFlags().StringVar(&directory, "directory", "", "DEPRECATED: Use --output instead. Base directory for cloning")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.olive-clone.yaml)")

	// Mark directory as deprecated
	rootCmd.PersistentFlags().MarkDeprecated("directory", "use --output or -o instead")

	// Add version template with colors
	rootCmd.SetVersionTemplate(color.New(color.FgCyan, color.Bold).Sprintf("⚡ SyncX v{{.Version}}\n"))
}

func initConfig() {
	// Configuration initialization will be implemented later
}

func setupGlobals(cmd *cobra.Command) {
	// Commands that don't require inventory file
	commandsWithoutInventory := map[string]bool{
		"scan":    true,
		"version": true,
		"help":    true,
	}

	// Validate protocol
	if protocol != "ssh" && protocol != "http" {
		color.New(color.FgRed, color.Bold).Printf("❌ Invalid protocol: %s. Must be 'ssh' or 'http'\n", protocol)
		os.Exit(1)
	}

	// Handle output directory logic
	setupOutputDirectory()

	// Check if inventory file exists (skip for commands that don't need it)
	cmdName := cmd.Name()
	if !commandsWithoutInventory[cmdName] {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			color.New(color.FgRed, color.Bold).Printf("❌ Inventory file not found: %s\n", file)
			color.New(color.FgYellow).Println("💡 Tip: Create a projects-inventory.json file or specify a different file with --file")
			os.Exit(1)
		}
	}
}

// setupOutputDirectory determines the best output directory
func setupOutputDirectory() {
	// If user provided --output, use that
	if outputDir != "" {
		directory = outputDir
		return
	}

	// If user provided deprecated --directory, use that with a warning
	if directory != "" {
		color.New(color.FgYellow).Println("⚠️  --directory is deprecated, use --output or -o instead")
		return
	}

	// Default: Use ../repositories (outside the script folder)
	executableDir, err := os.Executable()
	if err != nil {
		// Fallback to current working directory if we can't determine executable location
		directory = "./repositories"
	} else {
		// Get the directory containing the executable and go one level up
		scriptDir := filepath.Dir(executableDir)
		parentDir := filepath.Dir(scriptDir)
		directory = filepath.Join(parentDir, "repositories")
	}

	// Ensure the directory is absolute for consistency
	absDir, err := filepath.Abs(directory)
	if err != nil {
		color.New(color.FgYellow).Printf("⚠️  Could not resolve absolute path for %s, using as-is\n", directory)
	} else {
		directory = absDir
	}
}

// GetOutputDirectory returns the current output directory (for use by commands)
func GetOutputDirectory() string {
	return directory
}