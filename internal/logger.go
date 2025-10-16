package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Logger handles colored output based on verbose mode
type Logger struct {
	verbose bool
}

// NewLogger creates a new logger instance
func NewLogger(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

// Info logs info messages (only in verbose mode)
func (l *Logger) Info(format string, args ...interface{}) {
	if l.verbose {
		color.New(color.FgCyan).Printf("ℹ️  "+format+"\n", args...)
	}
}

// Success logs success messages
func (l *Logger) Success(format string, args ...interface{}) {
	color.New(color.FgGreen, color.Bold).Printf("✅ "+format+"\n", args...)
}

// Warning logs warning messages
func (l *Logger) Warning(format string, args ...interface{}) {
	color.New(color.FgYellow, color.Bold).Printf("⚠️  "+format+"\n", args...)
}

// Error logs error messages
func (l *Logger) Error(format string, args ...interface{}) {
	color.New(color.FgRed, color.Bold).Printf("❌ "+format+"\n", args...)
}

// Debug logs debug messages (only in verbose mode)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		color.New(color.FgWhite, color.Faint).Printf("🐛 "+format+"\n", args...)
	}
}

// Skip logs skip messages (only in verbose mode)
func (l *Logger) Skip(format string, args ...interface{}) {
	if l.verbose {
		color.New(color.FgMagenta).Printf("⏭️  "+format+"\n", args...)
	}
}

// Group logs group messages (only in verbose mode)
func (l *Logger) Group(format string, args ...interface{}) {
	if l.verbose {
		color.New(color.FgBlue, color.Bold).Printf("📁 "+format+"\n", args...)
	}
}

// Cloning logs cloning messages
func (l *Logger) Cloning(format string, args ...interface{}) {
	if l.verbose {
		color.New(color.FgGreen).Printf("📥 "+format+"\n", args...)
	}
}

// DryRun logs dry run messages
func (l *Logger) DryRun(format string, args ...interface{}) {
	color.New(color.FgCyan, color.Italic).Printf("🔍 [DRY RUN] "+format+"\n", args...)
}

// Pulling logs pulling messages (only in verbose mode)
func (l *Logger) Pulling(format string, args ...interface{}) {
	if l.verbose {
		color.New(color.FgBlue).Printf("📤 "+format+"\n", args...)
	}
}

// Updated logs updated messages
func (l *Logger) Updated(format string, args ...interface{}) {
	color.New(color.FgGreen).Printf("🔄 "+format+"\n", args...)
}

// Scan logs scan messages (only in verbose mode)
func (l *Logger) Scan(format string, args ...interface{}) {
	if l.verbose {
		color.New(color.FgMagenta).Printf("🔍 "+format+"\n", args...)
	}
}

// Header prints a colored header
func (l *Logger) Header(text string) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold, color.Underline).Println(text)
	color.New(color.FgCyan).Println(strings.Repeat("═", len(text)))
	fmt.Println()
}

// Separator prints a colored separator
func (l *Logger) Separator() {
	color.New(color.FgBlue).Println("─────────────────────────────────────────────────")
}

// Banner prints the application banner
func (l *Logger) Banner() {
	banner := `
🫒 ╔═══════════════════════════════════════╗
   ║        Olive Clone Assistant          ║
   ║     Modern Repository Management      ║
   ╚═══════════════════════════════════════╝`
	
	color.New(color.FgGreen, color.Bold).Println(banner)
	fmt.Println()
}

// Summary prints operation summary with colors
func (l *Logger) Summary(summary Summary) {
	l.Header("📊 Operation Summary")

	color.New(color.FgWhite, color.Bold).Printf("Total Projects: %d\n", summary.TotalProjects)
	color.New(color.FgGreen, color.Bold).Printf("✅ Successful: %d\n", summary.SuccessCount)

	if summary.FailureCount > 0 {
		color.New(color.FgRed, color.Bold).Printf("❌ Failed: %d\n", summary.FailureCount)
	}

	if summary.ClonedCount > 0 {
		color.New(color.FgBlue, color.Bold).Printf("📥 Cloned: %d\n", summary.ClonedCount)
	}

	if summary.UpdatedCount > 0 {
		color.New(color.FgGreen, color.Bold).Printf("🔄 Updated: %d\n", summary.UpdatedCount)
	}

	if summary.SkippedCount > 0 {
		color.New(color.FgYellow, color.Bold).Printf("⏭️  Skipped: %d\n", summary.SkippedCount)
	}

	if summary.EmptyCount > 0 {
		color.New(color.FgYellow).Printf("📭 Empty: %d\n", summary.EmptyCount)
	}

	if summary.TotalDuration != "" {
		color.New(color.FgCyan, color.Bold).Printf("⏱️  Duration: %s\n", summary.TotalDuration)
	}

	if len(summary.EmptyProjects) > 0 {
		fmt.Println()
		color.New(color.FgYellow).Println("Empty Projects (no commits):")
		for _, project := range summary.EmptyProjects {
			color.New(color.FgYellow).Printf("  • %s (%s)\n", project.Name, project.Group)
		}
	}

	if len(summary.FailedProjects) > 0 {
		fmt.Println()
		color.New(color.FgRed, color.Bold).Println("Failed Projects:")
		for _, project := range summary.FailedProjects {
			color.New(color.FgRed).Printf("  • %s (%s)\n", project.Name, project.Group)
		}
	}
}

// Timestamp returns the current timestamp for logging
func (l *Logger) Timestamp() string {
	return time.Now().Format("15:04:05")
}