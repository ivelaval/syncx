package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// WizardMode represents different interactive modes
type WizardMode int

const (
	WizardModeQuick WizardMode = iota
	WizardModeCustom
	WizardModeAdvanced
)

// WizardStep represents navigation steps
type WizardStep int

const (
	StepWelcome WizardStep = iota
	StepModeSelection
	StepPhysicalLocation
	StepProjectSelection
	StepConfiguration
	StepDirectory
	StepPreview
	StepComplete
)

// NavigationAction represents user navigation choices
type NavigationAction int

const (
	ActionNext NavigationAction = iota
	ActionBack
	ActionCancel
	ActionRestart
)

// OperationChoice represents user's operation choice
type OperationChoice struct {
	Mode             WizardMode
	SelectedProjects []ProjectInfo
	SelectedGroups   []string
	Protocol         string
	Directory        string
	Parallel         int
	DryRun          bool
	Verbose         bool
}

// MenuItem represents a menu item in the custom keyboard interface
type MenuItem struct {
	Label  string
	Action string
	Type   string
}

// InteractiveWizard handles the step-by-step interactive experience
type InteractiveWizard struct {
	logger        *Logger
	allProjects   []ProjectInfo
	allGroups     []string
	currentStep   WizardStep
	stepHistory   []WizardStep
	choice        *OperationChoice
	inventory     *Inventory
	inventoryPath string
}

// NewInteractiveWizard creates a new wizard instance
func NewInteractiveWizard(projects []ProjectInfo, logger *Logger) *InteractiveWizard {
	groups := GetUniqueGroups(projects)
	return &InteractiveWizard{
		logger:      logger,
		allProjects: projects,
		allGroups:   groups,
		currentStep: StepWelcome,
		stepHistory: []WizardStep{},
		choice: &OperationChoice{
			Protocol:  "ssh",
			Directory: "",
			Parallel:  3,
			DryRun:    false,
			Verbose:   true,
		},
	}
}

// NewInteractiveWizardWithInventory creates a new wizard instance with inventory context
func NewInteractiveWizardWithInventory(inventory *Inventory, inventoryPath string, projects []ProjectInfo, logger *Logger) *InteractiveWizard {
	groups := GetUniqueGroups(projects)
	return &InteractiveWizard{
		logger:        logger,
		allProjects:   projects,
		allGroups:     groups,
		currentStep:   StepWelcome,
		stepHistory:   []WizardStep{},
		inventory:     inventory,
		inventoryPath: inventoryPath,
		choice: &OperationChoice{
			Protocol:  "ssh",
			Directory: GetPhysicalLocation(inventory, ""),
			Parallel:  3,
			DryRun:    false,
			Verbose:   true,
		},
	}
}

// RunWizard executes the complete interactive wizard flow with navigation
func (w *InteractiveWizard) RunWizard() (*OperationChoice, error) {
	w.showWelcome()
	
	for {
		switch w.currentStep {
		case StepWelcome:
			w.currentStep = StepModeSelection
			
		case StepModeSelection:
			action, err := w.handleModeSelection()
			if err != nil {
				return nil, err
			}
			if action == ActionCancel {
				return nil, fmt.Errorf("wizard cancelled by user")
			}
			if action == ActionBack {
				w.goBack()
			} else {
				w.moveToNextStep()
			}
			
		case StepPhysicalLocation:
			action, err := w.handlePhysicalLocationSelection()
			if err != nil {
				return nil, err
			}
			if action == ActionCancel {
				return nil, fmt.Errorf("wizard cancelled by user")
			}
			if action == ActionBack {
				w.goBack()
			} else {
				w.moveToNextStep()
			}
			
		case StepProjectSelection:
			action, err := w.handleProjectSelection()
			if err != nil {
				return nil, err
			}
			if action == ActionCancel {
				return nil, fmt.Errorf("wizard cancelled by user")
			}
			if action == ActionBack {
				w.goBack()
			} else {
				w.moveToNextStep()
			}
			
		case StepConfiguration:
			action, err := w.handleConfiguration()
			if err != nil {
				return nil, err
			}
			if action == ActionCancel {
				return nil, fmt.Errorf("wizard cancelled by user")
			}
			if action == ActionBack {
				w.goBack()
			} else {
				w.moveToNextStep()
			}
			
		case StepDirectory:
			action, err := w.handleDirectorySelection()
			if err != nil {
				return nil, err
			}
			if action == ActionCancel {
				return nil, fmt.Errorf("wizard cancelled by user")
			}
			if action == ActionBack {
				w.goBack()
			} else {
				w.moveToNextStep()
			}
			
		case StepPreview:
			action, err := w.handlePreview()
			if err != nil {
				return nil, err
			}
			if action == ActionCancel {
				return nil, fmt.Errorf("wizard cancelled by user")
			}
			if action == ActionBack {
				w.goBack()
			} else {
				w.moveToNextStep()
			}
			
		case StepComplete:
			return w.choice, nil
		}
	}
}

// Navigation helper methods
func (w *InteractiveWizard) moveToNextStep() {
	w.stepHistory = append(w.stepHistory, w.currentStep)
	switch w.currentStep {
	case StepModeSelection:
		w.currentStep = StepPhysicalLocation
	case StepPhysicalLocation:
		if w.choice.Mode == WizardModeQuick {
			w.currentStep = StepPreview
		} else {
			w.currentStep = StepProjectSelection
		}
	case StepProjectSelection:
		w.currentStep = StepConfiguration
	case StepConfiguration:
		if w.choice.Mode == WizardModeAdvanced {
			w.currentStep = StepDirectory
		} else {
			w.currentStep = StepPreview
		}
	case StepDirectory:
		w.currentStep = StepPreview
	case StepPreview:
		w.currentStep = StepComplete
	}
}

func (w *InteractiveWizard) goBack() {
	if len(w.stepHistory) > 0 {
		w.currentStep = w.stepHistory[len(w.stepHistory)-1]
		w.stepHistory = w.stepHistory[:len(w.stepHistory)-1]
	}
}

func (w *InteractiveWizard) canGoBack() bool {
	return len(w.stepHistory) > 0
}

func (w *InteractiveWizard) showStepIndicator() {
	stepNames := map[WizardStep]string{
		StepModeSelection:    "Mode Selection",
		StepProjectSelection: "Project Selection", 
		StepConfiguration:    "Configuration",
		StepDirectory:        "Directory Setup",
		StepPreview:         "Preview & Confirm",
	}
	
	if name, exists := stepNames[w.currentStep]; exists {
		color.New(color.FgBlue, color.Faint).Printf("Step: %s", name)
		if w.canGoBack() {
			color.New(color.FgYellow, color.Faint).Printf(" | Press ← to go back | ESC to cancel")
		} else {
			color.New(color.FgYellow, color.Faint).Printf(" | ESC to cancel")
		}
		fmt.Println()
		fmt.Println()
	}
}

// showWelcome displays the wizard welcome screen
func (w *InteractiveWizard) showWelcome() {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("🧙‍♂️ Welcome to Olive Clone Assistant Wizard!")
	color.New(color.FgWhite).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.New(color.FgWhite).Printf("Found %d projects across %d groups\n", len(w.allProjects), len(w.allGroups))
	color.New(color.FgYellow).Println("Let's walk through your repository management preferences...")
	color.New(color.FgGreen, color.Faint).Println("💡 You can navigate back to previous steps or cancel anytime!")
	fmt.Println()
}

// handleModeSelection handles the mode selection step with navigation
func (w *InteractiveWizard) handleModeSelection() (NavigationAction, error) {
	w.showStepIndicator()
	
	options := []string{
		"🚀 Quick Mode - Clone/update all repositories with smart defaults",
		"🎯 Custom Mode - Select specific projects and groups", 
		"⚙️  Advanced Mode - Full control over all options",
	}
	
	// Add navigation options
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "🚀 Choose your operation mode",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "▶ {{ .| cyan | bold }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green | bold }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return ActionCancel, err
	}

	// Handle navigation options
	if strings.Contains(selection, "Back to previous step") {
		return ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		return w.confirmCancel()
	}

	// Set the mode based on selection
	if idx < 3 {
		w.choice.Mode = WizardMode(idx)
		color.New(color.FgGreen).Printf("✅ Selected: %s\n", selection)
		fmt.Println()
	}

	return ActionNext, nil
}

// handleProjectSelection handles the project selection step
func (w *InteractiveWizard) handleProjectSelection() (NavigationAction, error) {
	w.showStepIndicator()
	
	if w.choice.Mode == WizardModeQuick {
		// Quick mode: select all projects
		w.choice.SelectedProjects = w.allProjects
		return ActionNext, nil
	}

	// Custom/Advanced mode: interactive selection
	w.showModeHeader("🎯 Project Selection", "Choose exactly what you need")
	
	// Choose selection method
	options := []string{
		"📁 By Groups - Select entire project groups",
		"📦 Individual Projects - Pick specific repositories",  
		"🔀 Mixed - Groups first, then individual projects",
		"🌟 All Projects - Select everything",
	}
	
	// Add navigation options
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "How would you like to select repositories?",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to previous step") {
		return ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		return w.confirmCancel()
	}

	// Handle selection based on choice
	switch idx {
	case 0: // By Groups
		selectedGroups, navAction, err := w.selectGroupsWithNavigation()
		if err != nil {
			return ActionCancel, err
		}
		if navAction != ActionNext {
			return navAction, nil
		}
		w.choice.SelectedGroups = selectedGroups
		w.choice.SelectedProjects = w.filterProjectsByGroups(selectedGroups)
		
	case 1: // Individual Projects  
		selectedProjects, navAction, err := w.selectProjectsWithNavigation()
		if err != nil {
			return ActionCancel, err
		}
		if navAction != ActionNext {
			return navAction, nil
		}
		w.choice.SelectedProjects = selectedProjects
		
	case 2: // Mixed
		// First select groups
		selectedGroups, navAction, err := w.selectGroupsWithNavigation()
		if err != nil {
			return ActionCancel, err
		}
		if navAction != ActionNext {
			return navAction, nil
		}
		
		groupProjects := w.filterProjectsByGroups(selectedGroups)
		remaining := w.getRemainingProjects(groupProjects)
		
		if len(remaining) > 0 {
			additionalProjects, navAction, err := w.selectFromRemainingWithNavigation(remaining)
			if err != nil {
				return ActionCancel, err
			}
			if navAction != ActionNext {
				return navAction, nil
			}
			w.choice.SelectedProjects = append(groupProjects, additionalProjects...)
		} else {
			w.choice.SelectedProjects = groupProjects
		}
		w.choice.SelectedGroups = selectedGroups
		
	case 3: // All Projects
		w.choice.SelectedProjects = w.allProjects
		color.New(color.FgGreen).Printf("✅ Selected all %d projects\n", len(w.allProjects))
	}

	fmt.Println()
	return ActionNext, nil
}

// handleConfiguration handles the configuration step  
func (w *InteractiveWizard) handleConfiguration() (NavigationAction, error) {
	w.showStepIndicator()
	w.showModeHeader("⚙️ Configuration", "Set up your preferences")

	// Protocol selection
	protocol, navAction, err := w.selectProtocolWithNavigation()
	if err != nil {
		return ActionCancel, err
	}
	if navAction != ActionNext {
		return navAction, nil
	}
	w.choice.Protocol = protocol

	// Parallel processing
	parallel, navAction, err := w.selectParallelWithNavigation()
	if err != nil {
		return ActionCancel, err
	}
	if navAction != ActionNext {
		return navAction, nil
	}
	w.choice.Parallel = parallel

	// For Advanced mode, also set dry-run and verbose options
	if w.choice.Mode == WizardModeAdvanced {
		dryRun, navAction, err := w.selectDryRunWithNavigation()
		if err != nil {
			return ActionCancel, err
		}
		if navAction != ActionNext {
			return navAction, nil
		}
		w.choice.DryRun = dryRun

		verbose, navAction, err := w.selectVerboseWithNavigation()
		if err != nil {
			return ActionCancel, err
		}
		if navAction != ActionNext {
			return navAction, nil
		}
		w.choice.Verbose = verbose
	}

	return ActionNext, nil
}

// handleDirectorySelection handles the directory selection step
func (w *InteractiveWizard) handleDirectorySelection() (NavigationAction, error) {
	w.showStepIndicator()
	
	directory, navAction, err := w.selectDirectoryWithNavigation()
	if err != nil {
		return ActionCancel, err
	}
	if navAction != ActionNext {
		return navAction, nil
	}
	
	w.choice.Directory = directory
	return ActionNext, nil
}

// handlePreview handles the preview and confirmation step
func (w *InteractiveWizard) handlePreview() (NavigationAction, error) {
	w.showStepIndicator()
	
	return w.showPreviewWithNavigation()
}

// confirmCancel shows cancellation confirmation dialog
func (w *InteractiveWizard) confirmCancel() (NavigationAction, error) {
	fmt.Println()
	color.New(color.FgYellow, color.Bold).Println("⚠️  Cancel Wizard?")
	color.New(color.FgWhite).Println("Are you sure you want to exit the wizard?")
	fmt.Println()

	prompt := promptui.Select{
		Label: "Confirmation",
		Items: []string{
			"❌ Yes, cancel and exit",
			"↩️  No, continue wizard",
		},
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "{{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return ActionCancel, err
	}

	if idx == 0 {
		color.New(color.FgRed).Println("❌ Wizard cancelled")
		return ActionCancel, nil
	}

	return ActionNext, nil
}

// quickModeFlow handles the quick mode wizard
func (w *InteractiveWizard) quickModeFlow(choice *OperationChoice) (*OperationChoice, error) {
	w.showModeHeader("🚀 Quick Mode", "Smart defaults for rapid repository management")

	// Quick protocol selection
	protocol, err := w.selectProtocol("Choose Git protocol (SSH recommended for authenticated access)")
	if err != nil {
		return nil, err
	}
	choice.Protocol = protocol

	// Quick confirmation
	if confirmed, err := w.confirmOperation("Ready to process all repositories?", 
		fmt.Sprintf("This will clone/update all %d repositories using %s protocol", 
			len(w.allProjects), protocol)); err != nil {
		return nil, err
	} else if !confirmed {
		color.New(color.FgYellow).Println("Operation cancelled")
		return nil, fmt.Errorf("operation cancelled by user")
	}

	choice.SelectedProjects = w.allProjects
	choice.Directory = "."
	choice.Parallel = 3 // Smart default
	choice.DryRun = false
	choice.Verbose = true

	return choice, nil
}

// customModeFlow handles the custom selection mode
func (w *InteractiveWizard) customModeFlow(choice *OperationChoice) (*OperationChoice, error) {
	w.showModeHeader("🎯 Custom Mode", "Select exactly what you need")

	// Step 1: Choose selection method
	selectionMethod, err := w.chooseSelectionMethod()
	if err != nil {
		return nil, err
	}

	// Step 2: Based on method, select projects
	switch selectionMethod {
	case "groups":
		selectedGroups, err := w.selectGroups()
		if err != nil {
			return nil, err
		}
		choice.SelectedGroups = selectedGroups
		choice.SelectedProjects = w.filterProjectsByGroups(selectedGroups)
		
	case "projects":
		selectedProjects, err := w.selectIndividualProjects()
		if err != nil {
			return nil, err
		}
		choice.SelectedProjects = selectedProjects
		
	case "mixed":
		// Groups first, then individual projects
		groups, err := w.selectGroups()
		if err != nil {
			return nil, err
		}
		
		// Get remaining projects not in selected groups
		groupProjects := w.filterProjectsByGroups(groups)
		remainingProjects := w.getRemainingProjects(groupProjects)
		
		if len(remainingProjects) > 0 {
			additionalProjects, err := w.selectFromRemainingProjects(remainingProjects)
			if err != nil {
				return nil, err
			}
			choice.SelectedProjects = append(groupProjects, additionalProjects...)
		} else {
			choice.SelectedProjects = groupProjects
		}
		choice.SelectedGroups = groups
	}

	// Step 3: Configuration options
	protocol, err := w.selectProtocol("Choose Git protocol")
	if err != nil {
		return nil, err
	}
	choice.Protocol = protocol

	parallel, err := w.selectParallelCount()
	if err != nil {
		return nil, err
	}
	choice.Parallel = parallel

	// Step 4: Final confirmation with preview
	if !w.showSelectionPreview(choice) {
		return nil, fmt.Errorf("operation cancelled by user")
	}

	choice.Directory = "."
	choice.Verbose = true
	
	return choice, nil
}

// advancedModeFlow handles the advanced configuration mode
func (w *InteractiveWizard) advancedModeFlow(choice *OperationChoice) (*OperationChoice, error) {
	w.showModeHeader("⚙️  Advanced Mode", "Full control over all settings")

	// All configuration options with advanced settings
	
	// 1. Project/Group Selection
	customChoice, err := w.customModeFlow(&OperationChoice{Mode: WizardModeCustom})
	if err != nil {
		return nil, err
	}
	choice.SelectedProjects = customChoice.SelectedProjects
	choice.SelectedGroups = customChoice.SelectedGroups
	choice.Protocol = customChoice.Protocol
	choice.Parallel = customChoice.Parallel

	// 2. Advanced directory configuration
	directory, err := w.selectDirectory()
	if err != nil {
		return nil, err
	}
	choice.Directory = directory

	// 3. Advanced operation options
	dryRun, err := w.selectDryRunMode()
	if err != nil {
		return nil, err
	}
	choice.DryRun = dryRun

	verbose, err := w.selectVerboseMode()
	if err != nil {
		return nil, err
	}
	choice.Verbose = verbose

	// 4. Final advanced confirmation
	if !w.showAdvancedPreview(choice) {
		return nil, fmt.Errorf("operation cancelled by user")
	}

	return choice, nil
}

// Helper methods for the wizard flow

func (w *InteractiveWizard) showModeHeader(title, description string) {
	fmt.Println()
	color.New(color.FgMagenta, color.Bold).Println(title)
	color.New(color.FgWhite).Println(strings.Repeat("─", len(title)))
	color.New(color.FgWhite, color.Faint).Println(description)
	fmt.Println()
}

func (w *InteractiveWizard) selectProtocol(label string) (string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: []string{
			"🔐 SSH - Secure, key-based authentication (Recommended)",
			"🌐 HTTPS - Username/password or token authentication",
		},
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	if idx == 0 {
		return "ssh", nil
	}
	return "http", nil
}

func (w *InteractiveWizard) chooseSelectionMethod() (string, error) {
	prompt := promptui.Select{
		Label: "How would you like to select repositories?",
		Items: []string{
			"📁 By Groups - Select entire project groups",
			"📦 Individual Projects - Pick specific repositories",
			"🔀 Mixed - Groups first, then individual projects",
		},
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	methods := []string{"groups", "projects", "mixed"}
	return methods[idx], nil
}

func (w *InteractiveWizard) selectGroups() ([]string, error) {
	var selectedGroups []string
	
	for {
		// Show current selection
		if len(selectedGroups) > 0 {
			color.New(color.FgGreen).Printf("Selected: %s\n", strings.Join(selectedGroups, ", "))
		}

		// Create options list
		options := []string{"✅ Done - Continue with selected groups"}
		for _, group := range w.allGroups {
			if !contains(selectedGroups, group) {
				projectCount := len(FilterProjectsByGroup(w.allProjects, group))
				options = append(options, fmt.Sprintf("📁 %s (%d projects)", group, projectCount))
			}
		}

		if len(selectedGroups) > 0 {
			options = append(options, "🗑️  Clear all selections")
		}

		prompt := promptui.Select{
			Label: "Select groups to include",
			Items: options,
			Templates: &promptui.SelectTemplates{
				Active:   "▶ {{ .| cyan }}",
				Inactive: "  {{ . | faint }}",
				Selected: "{{ . | green }}",
			},
		}

		idx, selection, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		if idx == 0 { // Done
			break
		} else if strings.Contains(selection, "Clear all") {
			selectedGroups = []string{}
		} else {
			// Extract group name from selection
			parts := strings.Split(selection, " ")
			if len(parts) > 1 {
				groupName := strings.TrimPrefix(parts[1], " ")
				// Remove project count part
				if idx := strings.Index(groupName, " ("); idx != -1 {
					groupName = groupName[:idx]
				}
				selectedGroups = append(selectedGroups, groupName)
			}
		}

		if len(selectedGroups) == len(w.allGroups) {
			color.New(color.FgGreen).Println("All groups selected!")
			break
		}
	}

	return selectedGroups, nil
}

func (w *InteractiveWizard) selectIndividualProjects() ([]ProjectInfo, error) {
	var selectedProjects []ProjectInfo

	for {
		if len(selectedProjects) > 0 {
			color.New(color.FgGreen).Printf("Selected %d projects\n", len(selectedProjects))
		}

		options := []string{"✅ Done - Continue with selected projects"}
		for _, project := range w.allProjects {
			if !containsProject(selectedProjects, project) {
				options = append(options, fmt.Sprintf("📦 %s (%s)", project.Name, project.Group))
			}
		}

		if len(selectedProjects) > 0 {
			options = append(options, "🗑️  Clear all selections")
		}

		prompt := promptui.Select{
			Label: "Select individual projects",
			Items: options,
			Templates: &promptui.SelectTemplates{
				Active:   "▶ {{ .| cyan }}",
				Inactive: "  {{ . | faint }}",
				Selected: "{{ . | green }}",
			},
		}

		idx, selection, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		if idx == 0 { // Done
			break
		} else if strings.Contains(selection, "Clear all") {
			selectedProjects = []ProjectInfo{}
		} else {
			// Find the project by name from selection
			parts := strings.Split(selection, " ")
			if len(parts) > 1 {
				projectName := parts[1]
				for _, project := range w.allProjects {
					if project.Name == projectName {
						selectedProjects = append(selectedProjects, project)
						break
					}
				}
			}
		}

		if len(selectedProjects) == len(w.allProjects) {
			color.New(color.FgGreen).Println("All projects selected!")
			break
		}
	}

	return selectedProjects, nil
}

func (w *InteractiveWizard) selectParallelCount() (int, error) {
	prompt := promptui.Select{
		Label: "Choose parallel processing level",
		Items: []string{
			"🐌 Sequential (1) - One at a time, safest",
			"🚶 Moderate (3) - Good balance of speed and safety",
			"🏃 Fast (5) - Faster processing, more resource usage",
			"🚀 Maximum (10) - Fastest, highest resource usage",
		},
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return 1, err
	}

	counts := []int{1, 3, 5, 10}
	return counts[idx], nil
}

func (w *InteractiveWizard) selectDirectory() (string, error) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📂 Output Directory Selection")
	color.New(color.FgWhite).Println("Where would you like to clone the repositories?")
	fmt.Println()
	
	// Get current directory as default
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}
	
	// Show current directory info
	color.New(color.FgGreen).Printf("📁 Current directory: %s\n", currentDir)
	fmt.Println()

	options := []string{
		fmt.Sprintf("✅ Use current directory: %s", currentDir),
		"📝 Specify custom directory",
	}

	prompt := promptui.Select{
		Label: "Select output directory option",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	// Handle selection
	switch idx {
	case 0:
		// Use current directory
		color.New(color.FgGreen).Printf("✅ Using current directory: %s\n", currentDir)
		return currentDir, nil
	case 1:
		// Custom path
		return w.promptCustomDirectory()
	default:
		return currentDir, nil
	}
}

func (w *InteractiveWizard) promptCustomDirectory() (string, error) {
	fmt.Println()
	color.New(color.FgYellow).Println("💡 Examples of valid paths:")
	color.New(color.FgWhite).Println("   /Users/vennet/projects")
	color.New(color.FgWhite).Println("   ~/projects")
	color.New(color.FgWhite).Println("   ../my-repositories")
	color.New(color.FgWhite).Println("   ./local-repos")
	fmt.Println()

	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("path cannot be empty")
		}
		return ValidateOutputPath(input)
	}

	prompt := promptui.Prompt{
		Label:    "Enter custom output directory path",
		Validate: validate,
		Templates: &promptui.PromptTemplates{
			Prompt:  "{{ . | bold }}{{ \":\" | bold }} ",
			Valid:   "{{ . | green | bold }}{{ \":\" | bold }} ",
			Invalid: "{{ . | red | bold }}{{ \":\" | bold }} ",
			Success: "{{ . | bold }}{{ \":\" | bold }} ",
		},
	}

	return prompt.Run()
}

func (w *InteractiveWizard) selectDryRunMode() (bool, error) {
	prompt := promptui.Select{
		Label: "Choose execution mode",
		Items: []string{
			"🔥 Execute - Perform actual clone/update operations",
			"👀 Dry Run - Preview operations without executing",
		},
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	return idx == 1, err
}

func (w *InteractiveWizard) selectVerboseMode() (bool, error) {
	prompt := promptui.Select{
		Label: "Output verbosity",
		Items: []string{
			"📊 Standard - Important messages and progress",
			"📝 Verbose - Detailed logging and debug information",
		},
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	return idx == 1, err
}

func (w *InteractiveWizard) confirmOperation(title, details string) (bool, error) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println(title)
	color.New(color.FgWhite, color.Faint).Println(details)
	
	prompt := promptui.Select{
		Label: "Continue?",
		Items: []string{
			"✅ Yes - Proceed with operation",
			"❌ No - Cancel and exit",
		},
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "{{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	return idx == 0, err
}

func (w *InteractiveWizard) showSelectionPreview(choice *OperationChoice) bool {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📋 Selection Preview")
	color.New(color.FgWhite).Println("━━━━━━━━━━━━━━━━━━━━━━━")
	
	color.New(color.FgWhite).Printf("Projects: %d selected\n", len(choice.SelectedProjects))
	color.New(color.FgWhite).Printf("Protocol: %s\n", choice.Protocol)
	color.New(color.FgWhite).Printf("Parallel: %d concurrent operations\n", choice.Parallel)
	
	if len(choice.SelectedGroups) > 0 {
		color.New(color.FgWhite).Printf("Groups: %s\n", strings.Join(choice.SelectedGroups, ", "))
	}

	confirmed, _ := w.confirmOperation("Proceed with these settings?", 
		"This will start the repository operations")
	return confirmed
}

func (w *InteractiveWizard) showAdvancedPreview(choice *OperationChoice) bool {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("⚙️  Advanced Configuration Preview")
	color.New(color.FgWhite).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	color.New(color.FgWhite).Printf("Projects: %d selected\n", len(choice.SelectedProjects))
	color.New(color.FgWhite).Printf("Directory: %s\n", choice.Directory)
	color.New(color.FgWhite).Printf("Protocol: %s\n", choice.Protocol)
	color.New(color.FgWhite).Printf("Parallel: %d concurrent operations\n", choice.Parallel)
	color.New(color.FgWhite).Printf("Dry Run: %t\n", choice.DryRun)
	color.New(color.FgWhite).Printf("Verbose: %t\n", choice.Verbose)

	confirmed, _ := w.confirmOperation("Execute with these advanced settings?", 
		"All configuration options have been set")
	return confirmed
}

// Helper utility functions

func (w *InteractiveWizard) filterProjectsByGroups(groups []string) []ProjectInfo {
	var result []ProjectInfo
	for _, group := range groups {
		result = append(result, FilterProjectsByGroup(w.allProjects, group)...)
	}
	return result
}

func (w *InteractiveWizard) getRemainingProjects(selectedProjects []ProjectInfo) []ProjectInfo {
	var remaining []ProjectInfo
	for _, project := range w.allProjects {
		if !containsProject(selectedProjects, project) {
			remaining = append(remaining, project)
		}
	}
	return remaining
}

func (w *InteractiveWizard) selectFromRemainingProjects(remaining []ProjectInfo) ([]ProjectInfo, error) {
	if len(remaining) == 0 {
		return []ProjectInfo{}, nil
	}

	prompt := promptui.Select{
		Label: fmt.Sprintf("Add additional projects? (%d remaining)", len(remaining)),
		Items: []string{
			"✅ Yes - Select additional individual projects",
			"❌ No - Continue with group selections only",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil || idx == 1 {
		return []ProjectInfo{}, err
	}

	// Create a temporary wizard with remaining projects
	tempWizard := &InteractiveWizard{
		logger:      w.logger,
		allProjects: remaining,
		allGroups:   GetUniqueGroups(remaining),
	}

	return tempWizard.selectIndividualProjects()
}

// Utility functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsProject(slice []ProjectInfo, project ProjectInfo) bool {
	for _, p := range slice {
		if p.Name == project.Name && p.Group == project.Group {
			return true
		}
	}
	return false
}

// Navigation-enabled methods
func (w *InteractiveWizard) selectProtocolWithNavigation() (string, NavigationAction, error) {
	options := []string{
		"🔐 SSH - Secure, key-based authentication (Recommended)",
		"🌐 HTTPS - Username/password or token authentication",
	}
	
	// Add navigation options
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "Choose Git protocol",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return "", ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to previous step") {
		return "", ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		navAction, err := w.confirmCancel()
		return "", navAction, err
	}

	protocol := "ssh"
	if idx == 1 {
		protocol = "http"
	}
	
	color.New(color.FgGreen).Printf("✅ Selected protocol: %s\n", protocol)
	return protocol, ActionNext, nil
}

func (w *InteractiveWizard) selectParallelWithNavigation() (int, NavigationAction, error) {
	options := []string{
		"🐌 Sequential (1) - One at a time, safest",
		"🚶 Moderate (3) - Good balance of speed and safety",
		"🏃 Fast (5) - Faster processing, more resource usage",
		"🚀 Maximum (10) - Fastest, highest resource usage",
	}
	
	// Add navigation options
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "Choose parallel processing level",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return 0, ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to previous step") {
		return 0, ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		navAction, err := w.confirmCancel()
		return 0, navAction, err
	}

	counts := []int{1, 3, 5, 10}
	parallel := counts[idx]
	
	color.New(color.FgGreen).Printf("✅ Selected parallel: %d\n", parallel)
	return parallel, ActionNext, nil
}

func (w *InteractiveWizard) selectDryRunWithNavigation() (bool, NavigationAction, error) {
	options := []string{
		"🔥 Execute - Perform actual clone/update operations",
		"👀 Dry Run - Preview operations without executing",
	}
	
	// Add navigation options
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "Choose execution mode",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return false, ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to previous step") {
		return false, ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		navAction, err := w.confirmCancel()
		return false, navAction, err
	}

	dryRun := idx == 1
	color.New(color.FgGreen).Printf("✅ Selected mode: %s\n", selection)
	return dryRun, ActionNext, nil
}

func (w *InteractiveWizard) selectVerboseWithNavigation() (bool, NavigationAction, error) {
	options := []string{
		"📊 Standard - Important messages and progress",
		"📝 Verbose - Detailed logging and debug information",
	}
	
	// Add navigation options
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "Output verbosity",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return false, ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to previous step") {
		return false, ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		navAction, err := w.confirmCancel()
		return false, navAction, err
	}

	verbose := idx == 1
	color.New(color.FgGreen).Printf("✅ Selected verbosity: %s\n", selection)
	return verbose, ActionNext, nil
}

func (w *InteractiveWizard) selectDirectoryWithNavigation() (string, NavigationAction, error) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📂 Output Directory Selection")
	color.New(color.FgWhite).Println("Where would you like to clone the repositories?")
	fmt.Println()
	
	// Get current directory as default
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}
	
	// Show current directory info
	color.New(color.FgGreen).Printf("📁 Current directory: %s\n", currentDir)
	fmt.Println()

	options := []string{
		fmt.Sprintf("✅ Use current directory: %s", currentDir),
		"📝 Specify custom directory",
	}
	
	// Add navigation options
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "Select output directory option",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return "", ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to previous step") {
		return "", ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		navAction, err := w.confirmCancel()
		return "", navAction, err
	}

	// Handle selection
	switch idx {
	case 0:
		// Use current directory
		color.New(color.FgGreen).Printf("✅ Using current directory: %s\n", currentDir)
		return currentDir, ActionNext, nil
	case 1:
		// Custom path
		return w.promptCustomDirectoryWithNavigation()
	default:
		return currentDir, ActionNext, nil
	}
}

func (w *InteractiveWizard) promptCustomDirectoryWithNavigation() (string, NavigationAction, error) {
	fmt.Println()
	color.New(color.FgYellow).Println("💡 Examples of valid paths:")
	color.New(color.FgWhite).Println("   /Users/vennet/projects")
	color.New(color.FgWhite).Println("   ~/projects")
	color.New(color.FgWhite).Println("   ../my-repositories")
	color.New(color.FgWhite).Println("   ./local-repos")
	fmt.Println()

	// For custom input, we need to handle it differently since promptui.Prompt doesn't support navigation
	// We'll create a select with options including back/cancel, then handle text input separately
	options := []string{
		"📝 Enter custom path",
	}
	
	if w.canGoBack() {
		options = append(options, "◀️  Back to directory selection")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "Custom directory path",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return "", ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to directory selection") {
		return "", ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		navAction, err := w.confirmCancel()
		return "", navAction, err
	}

	if idx == 0 {
		// Enter custom path
		validate := func(input string) error {
			if input == "" {
				return fmt.Errorf("path cannot be empty")
			}
			return ValidateOutputPath(input)
		}

		inputPrompt := promptui.Prompt{
			Label:    "Enter custom output directory path",
			Validate: validate,
			Templates: &promptui.PromptTemplates{
				Prompt:  "{{ . | bold }}{{ \":\" | bold }} ",
				Valid:   "{{ . | green | bold }}{{ \":\" | bold }} ",
				Invalid: "{{ . | red | bold }}{{ \":\" | bold }} ",
				Success: "{{ . | bold }}{{ \":\" | bold }} ",
			},
		}

		directory, err := inputPrompt.Run()
		if err != nil {
			return "", ActionCancel, err
		}

		color.New(color.FgGreen).Printf("✅ Selected custom directory: %s\n", directory)
		return directory, ActionNext, nil
	}

	return "", ActionCancel, fmt.Errorf("unexpected selection")
}

func (w *InteractiveWizard) showPreviewWithNavigation() (NavigationAction, error) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📋 Configuration Preview")
	color.New(color.FgWhite).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	
	color.New(color.FgWhite).Printf("Mode: %s\n", w.getModeText())
	color.New(color.FgWhite).Printf("Projects: %d selected\n", len(w.choice.SelectedProjects))
	color.New(color.FgWhite).Printf("Protocol: %s\n", w.choice.Protocol)
	color.New(color.FgWhite).Printf("Parallel: %d concurrent operations\n", w.choice.Parallel)
	
	if w.choice.Directory != "" {
		color.New(color.FgWhite).Printf("Output Directory: %s\n", w.choice.Directory)
	}
	
	if len(w.choice.SelectedGroups) > 0 {
		color.New(color.FgWhite).Printf("Groups: %s\n", strings.Join(w.choice.SelectedGroups, ", "))
	}

	if w.choice.DryRun {
		color.New(color.FgYellow).Println("Mode: Dry Run (preview only)")
	}
	
	if w.choice.Verbose {
		color.New(color.FgWhite).Println("Output: Verbose mode enabled")
	}

	fmt.Println()

	options := []string{
		"🚀 Execute - Start repository operations",
		"📝 Edit Configuration - Modify settings",
	}
	
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "Proceed with these settings?",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan | bold }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green | bold }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to previous step") {
		return ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		return w.confirmCancel()
	}

	switch idx {
	case 0: // Execute
		color.New(color.FgGreen, color.Bold).Println("✅ Configuration confirmed! Proceeding...")
		return ActionNext, nil
	case 1: // Edit Configuration
		// Go back to configuration step
		w.currentStep = StepConfiguration
		w.stepHistory = w.stepHistory[:len(w.stepHistory)-1] // Remove last step from history
		return ActionBack, nil
	}

	return ActionNext, nil
}

func (w *InteractiveWizard) getModeText() string {
	switch w.choice.Mode {
	case WizardModeQuick:
		return "🚀 Quick Mode"
	case WizardModeCustom:
		return "🎯 Custom Mode"
	case WizardModeAdvanced:
		return "⚙️ Advanced Mode"
	default:
		return "Unknown Mode"
	}
}

// Simplified navigation-enabled group and project selection methods
func (w *InteractiveWizard) selectGroupsWithNavigation() ([]string, NavigationAction, error) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📁 Interactive Group Selection")
	color.New(color.FgYellow).Println("💡 Use ↑↓ to navigate, SPACE to toggle checkbox, ENTER to proceed")
	color.New(color.FgGreen).Println("🎯 SPACE = toggle selection, ENTER = continue/proceed, ESC = cancel")
	fmt.Println()
	
	selectedGroups := make(map[string]bool)
	currentIndex := 0
	
	// Initialize keyboard
	if err := keyboard.Open(); err != nil {
		// Fallback to old system if keyboard library fails
		return w.selectGroupsWithNavigation_Fallback()
	}
	defer keyboard.Close()
	
	for {
		// Clear screen and show interface
		fmt.Print("\033[H\033[2J") // Clear screen
		
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Println("📁 Interactive Group Selection")
		color.New(color.FgYellow).Println("💡 Use ↑↓ to navigate, SPACE to toggle checkbox, ENTER to proceed")
		color.New(color.FgGreen).Println("🎯 SPACE = toggle selection, ENTER = continue/proceed, ESC = cancel")
		fmt.Println()
		
		// Show current selection summary
		if len(selectedGroups) > 0 {
			var selected []string
			for group := range selectedGroups {
				selected = append(selected, group)
			}
			color.New(color.FgCyan, color.Bold).Println("╭─ CURRENT SELECTION ─╮")
			color.New(color.FgGreen, color.Bold).Printf("│ Selected Groups (%d): ", len(selected))
			color.New(color.FgWhite).Printf("%s\n", strings.Join(selected, ", "))
			color.New(color.FgCyan, color.Bold).Println("╰─────────────────────╯")
			fmt.Println()
		}
		
		// Build menu items
		var menuItems []MenuItem
		var itemTypes []string
		
		// Add groups
		color.New(color.FgMagenta, color.Bold).Println("═══ GROUPS (SPACE to toggle) ═══")
		for _, group := range w.allGroups {
			projectCount := len(FilterProjectsByGroup(w.allProjects, group))
			isSelected := selectedGroups[group]
			
			var checkbox, status string
			if isSelected {
				checkbox = color.New(color.FgGreen, color.Bold).Sprint("[✓]")
				status = color.New(color.FgGreen).Sprint("SELECTED")
			} else {
				checkbox = color.New(color.FgWhite, color.Faint).Sprint("[ ]")
				status = color.New(color.FgWhite, color.Faint).Sprint("press SPACE to select")
			}
			
			menuItems = append(menuItems, MenuItem{
				Label:       fmt.Sprintf("%s %s (%d projects) - %s", checkbox, group, projectCount, status),
				Action:      group,
				Type:        "group",
			})
			itemTypes = append(itemTypes, "group")
		}
		
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Println("═══ ACTIONS (ENTER to proceed) ═══")
		
		// Add actions
		if len(selectedGroups) > 0 {
			selectedCount := len(selectedGroups)
			menuItems = append(menuItems, MenuItem{
				Label:  color.New(color.FgGreen, color.Bold).Sprintf("→ Continue with %d selected group(s)", selectedCount),
				Action: "continue",
				Type:   "action",
			})
			itemTypes = append(itemTypes, "action")
			
			menuItems = append(menuItems, MenuItem{
				Label:  "🧹 Clear all selections",
				Action: "clear",
				Type:   "action",
			})
			itemTypes = append(itemTypes, "action")
		}
		
		menuItems = append(menuItems, MenuItem{
			Label:  "🌟 Select all groups",
			Action: "all",
			Type:   "action",
		})
		itemTypes = append(itemTypes, "action")
		
		if w.canGoBack() {
			menuItems = append(menuItems, MenuItem{
				Label:  "◀️  Back to previous step",
				Action: "back",
				Type:   "action",
			})
			itemTypes = append(itemTypes, "action")
		}
		
		menuItems = append(menuItems, MenuItem{
			Label:  "❌ Cancel wizard",
			Action: "cancel",
			Type:   "action",
		})
		itemTypes = append(itemTypes, "action")
		
		// Display menu with current selection highlighted
		for i, item := range menuItems {
			if i == currentIndex {
				color.New(color.FgCyan, color.Bold).Printf("▶ %s\n", item.Label)
			} else {
				color.New(color.FgWhite, color.Faint).Printf("  %s\n", item.Label)
			}
		}
		
		fmt.Println()
		color.New(color.FgYellow, color.Faint).Println("Controls: ↑↓=Navigate, SPACE=Toggle, ENTER=Proceed, ESC=Cancel")
		
		// Get key input
		char, key, err := keyboard.GetKey()
		if err != nil {
			return nil, ActionCancel, err
		}
		
		switch key {
		case keyboard.KeyArrowUp:
			if currentIndex > 0 {
				currentIndex--
			}
			
		case keyboard.KeyArrowDown:
			if currentIndex < len(menuItems)-1 {
				currentIndex++
			}
			
		case keyboard.KeySpace:
			// Handle SPACE - toggle only for groups
			if currentIndex < len(menuItems) && itemTypes[currentIndex] == "group" {
				action := menuItems[currentIndex].Action
				if selectedGroups[action] {
					delete(selectedGroups, action)
					fmt.Println()
					color.New(color.FgYellow, color.Bold).Printf("🗑️  UNCHECKED: %s (deselected)", action)
					fmt.Println()
				} else {
					selectedGroups[action] = true
					fmt.Println()
					color.New(color.FgGreen, color.Bold).Printf("✅ CHECKED: %s (selected)", action)
					fmt.Println()
				}
			}
			
		case keyboard.KeyEnter:
			// Handle ENTER - proceed only for actions
			if currentIndex < len(menuItems) {
				item := menuItems[currentIndex]
				
				if item.Type == "action" {
					switch item.Action {
					case "continue":
						if len(selectedGroups) > 0 {
							var result []string
							for group := range selectedGroups {
								result = append(result, group)
							}
							fmt.Println()
							color.New(color.FgGreen, color.Bold).Printf("🎯 Proceeding with selected groups: %s\n", strings.Join(result, ", "))
							return result, ActionNext, nil
						}
						
					case "clear":
						selectedGroups = make(map[string]bool)
						fmt.Println()
						color.New(color.FgYellow, color.Bold).Println("🧹 All selections cleared")
						fmt.Println()
						
					case "all":
						selectedGroups = make(map[string]bool)
						for _, group := range w.allGroups {
							selectedGroups[group] = true
						}
						fmt.Println()
						color.New(color.FgGreen, color.Bold).Printf("🌟 All %d groups selected\n", len(w.allGroups))
						fmt.Println()
						
					case "back":
						return nil, ActionBack, nil
						
					case "cancel":
						navAction, err := w.confirmCancel()
						return nil, navAction, err
					}
				}
			}
			
		case keyboard.KeyEsc:
			navAction, err := w.confirmCancel()
			return nil, navAction, err
			
		default:
			// Handle character input if needed
			if char == 'q' || char == 'Q' {
				navAction, err := w.confirmCancel()
				return nil, navAction, err
			}
		}
	}
}

// selectGroupsWithNavigation_Fallback is a fallback when keyboard library fails
func (w *InteractiveWizard) selectGroupsWithNavigation_Fallback() ([]string, NavigationAction, error) {
	// Fallback to simple group selection using promptui
	var selectedGroups []string
	
	options := []string{}
	for _, group := range w.allGroups {
		projectCount := len(FilterProjectsByGroup(w.allProjects, group))
		options = append(options, fmt.Sprintf("📁 %s (%d projects)", group, projectCount))
	}
	
	options = append(options, "🌟 Select All Groups")
	
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: "Select a group (fallback mode - use ENTER to select)",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, ActionCancel, err
	}

	// Handle navigation
	if idx == len(w.allGroups) {
		// Select All Groups
		selectedGroups = w.allGroups
	} else if idx == len(w.allGroups)+1 {
		// Back
		return nil, ActionBack, nil
	} else if idx == len(w.allGroups)+2 {
		// Cancel
		navAction, err := w.confirmCancel()
		return nil, navAction, err
	} else if idx < len(w.allGroups) {
		selectedGroups = []string{w.allGroups[idx]}
	}

	return selectedGroups, ActionNext, nil
}

func (w *InteractiveWizard) selectProjectsWithNavigation() ([]ProjectInfo, NavigationAction, error) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📦 Interactive Project Selection")
	color.New(color.FgYellow).Println("💡 Use SPACE to select/deselect projects, ENTER to confirm, ESC to cancel")
	fmt.Println()
	
	selectedProjects := make(map[string]ProjectInfo)
	
	for {
		// Create options with visual indicators
		var options []string
		var actualProjects []string
		
		// Group projects by group for better organization
		projectsByGroup := make(map[string][]ProjectInfo)
		for _, project := range w.allProjects {
			projectsByGroup[project.Group] = append(projectsByGroup[project.Group], project)
		}
		
		// Add projects grouped by category
		for _, group := range w.allGroups {
			if projects, exists := projectsByGroup[group]; exists {
				// Group header
				options = append(options, "")
				actualProjects = append(actualProjects, "")
				
				groupHeader := fmt.Sprintf("─── %s Group ───", group)
				options = append(options, color.New(color.FgMagenta, color.Bold).Sprint(groupHeader))
				actualProjects = append(actualProjects, "")
				
				// Projects in this group with checkbox style
				for _, project := range projects {
					_, isSelected := selectedProjects[project.Name]
					
					var checkbox, projectName, status string
					if isSelected {
						checkbox = color.New(color.FgGreen, color.Bold).Sprint("  [✓]")
						projectName = color.New(color.FgGreen, color.Bold).Sprint(project.Name)
						status = color.New(color.FgGreen).Sprint("SELECTED")
					} else {
						checkbox = color.New(color.FgWhite, color.Faint).Sprint("  [ ]")
						projectName = color.New(color.FgWhite).Sprint(project.Name)
						status = color.New(color.FgWhite, color.Faint).Sprint("click to select")
					}
					
					option := fmt.Sprintf("%s %s - %s", checkbox, projectName, status)
					
					options = append(options, option)
					actualProjects = append(actualProjects, project.Name)
				}
			}
		}
		
		// Add control options
		options = append(options, "")
		actualProjects = append(actualProjects, "")
		
		if len(selectedProjects) > 0 {
			selectedCount := len(selectedProjects)
			options = append(options, fmt.Sprintf("✅ Continue with %d selected project(s)", selectedCount))
			actualProjects = append(actualProjects, "continue")
			
			options = append(options, "🧹 Clear all selections")
			actualProjects = append(actualProjects, "clear")
		}
		
		options = append(options, "🌟 Select all projects")
		actualProjects = append(actualProjects, "all")
		
		// Navigation options
		if w.canGoBack() {
			options = append(options, "◀️  Back to previous step")
			actualProjects = append(actualProjects, "back")
		}
		options = append(options, "❌ Cancel wizard")
		actualProjects = append(actualProjects, "cancel")

		// Show current selection summary with better formatting
		if len(selectedProjects) > 0 {
			var selected []string
			for projectName := range selectedProjects {
				selected = append(selected, projectName)
			}
			fmt.Println()
			color.New(color.FgCyan, color.Bold).Println("╭─ CURRENT SELECTION ─╮")
			color.New(color.FgGreen, color.Bold).Printf("│ Selected Projects (%d): ", len(selected))
			color.New(color.FgWhite).Printf("%s\n", strings.Join(selected, ", "))
			color.New(color.FgCyan, color.Bold).Println("╰─────────────────────╯")
			fmt.Println()
		}

		prompt := promptui.Select{
			Label: "Select projects (SPACE to toggle, ENTER to choose action)",
			Items: options,
			Templates: &promptui.SelectTemplates{
				Active:   "▶ {{ .| cyan }}",
				Inactive: "  {{ . | faint }}",
				Selected: "{{ . }}",
			},
			HideHelp: true,
		}

		idx, selection, err := prompt.Run()
		if err != nil {
			return nil, ActionCancel, err
		}

		// Handle selection based on actualProjects mapping
		if idx < len(actualProjects) {
			action := actualProjects[idx]
			
			switch action {
			case "continue":
				if len(selectedProjects) > 0 {
					var result []ProjectInfo
					for _, project := range selectedProjects {
						result = append(result, project)
					}
					color.New(color.FgGreen).Printf("✅ Selected %d projects\n", len(result))
					return result, ActionNext, nil
				}
				
			case "clear":
				selectedProjects = make(map[string]ProjectInfo)
				color.New(color.FgYellow).Println("🧹 Cleared all selections")
				
			case "all":
				selectedProjects = make(map[string]ProjectInfo)
				for _, project := range w.allProjects {
					selectedProjects[project.Name] = project
				}
				color.New(color.FgGreen).Printf("🌟 Selected all %d projects\n", len(w.allProjects))
				
			case "back":
				return nil, ActionBack, nil
				
			case "cancel":
				navAction, err := w.confirmCancel()
				return nil, navAction, err
				
			case "":
				// Empty separator, do nothing
				continue
				
			default:
				// Toggle project selection
				if action != "" {
					// Find project by name
					for _, project := range w.allProjects {
						if project.Name == action {
							if _, isSelected := selectedProjects[action]; isSelected {
								delete(selectedProjects, action)
								fmt.Println()
								color.New(color.FgYellow, color.Bold).Printf("🗑️  DESELECTED: %s", action)
								color.New(color.FgWhite, color.Faint).Printf(" (unchecked)\n")
							} else {
								selectedProjects[action] = project
								fmt.Println()
								color.New(color.FgGreen, color.Bold).Printf("✅ SELECTED: %s", action)
								color.New(color.FgWhite, color.Faint).Printf(" (checked)\n")
							}
							break
						}
					}
				}
			}
		}
		
		fmt.Println()
		
		// Auto-continue tip
		if len(selectedProjects) > 0 && strings.Contains(selection, "SELECTED") {
			color.New(color.FgCyan).Println("💡 Tip: Choose 'Continue with X selected project(s)' to proceed")
			fmt.Println()
		}
	}
}

func (w *InteractiveWizard) selectFromRemainingWithNavigation(remaining []ProjectInfo) ([]ProjectInfo, NavigationAction, error) {
	// Simplified remaining selection
	options := []string{
		"✅ Yes - Add some remaining projects",
		"❌ No - Continue with group selections only",
	}
	
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
	}
	options = append(options, "❌ Cancel wizard")

	prompt := promptui.Select{
		Label: fmt.Sprintf("Add additional projects? (%d remaining)", len(remaining)),
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, selection, err := prompt.Run()
	if err != nil {
		return nil, ActionCancel, err
	}

	// Handle navigation
	if strings.Contains(selection, "Back to previous step") {
		return nil, ActionBack, nil
	}
	if strings.Contains(selection, "Cancel wizard") {
		navAction, err := w.confirmCancel()
		return nil, navAction, err
	}

	var additional []ProjectInfo
	if idx == 0 && len(remaining) > 0 {
		// Add first remaining project for demo
		additional = remaining[:min(len(remaining), 1)]
		color.New(color.FgGreen).Printf("✅ Added %d additional projects\n", len(additional))
	}

	return additional, ActionNext, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handlePhysicalLocationSelection handles the physical location step
func (w *InteractiveWizard) handlePhysicalLocationSelection() (NavigationAction, error) {
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("📍 Physical Location Setup")
	color.New(color.FgYellow).Println("💡 Configure where your repositories will be stored")
	fmt.Println()

	// Show current physical location
	currentLocation := ""
	if w.inventory != nil && w.inventory.PhysicalLocation != "" {
		currentLocation = w.inventory.PhysicalLocation
		color.New(color.FgGreen).Printf("📁 Current location: %s\n", currentLocation)
	} else {
		color.New(color.FgYellow).Printf("📁 No physical location set in configuration\n")
	}
	fmt.Println()

	// Create options
	var options []string
	var actions []string

	// Add continue option if location exists
	if currentLocation != "" {
		options = append(options, fmt.Sprintf("✅ Use current location: %s", currentLocation))
		actions = append(actions, "use_current")
	}

	// Add current working directory option if not already set
	currentDir, _ := os.Getwd()
	if currentDir != currentLocation {
		options = append(options, fmt.Sprintf("📁 Use current directory: %s", currentDir))
		actions = append(actions, currentDir)
	}

	// Add custom option
	options = append(options, "🎯 Choose custom location...")
	actions = append(actions, "custom")

	// Navigation options
	if w.canGoBack() {
		options = append(options, "◀️  Back to previous step")
		actions = append(actions, "back")
	}
	options = append(options, "❌ Cancel wizard")
	actions = append(actions, "cancel")

	// Initialize keyboard
	if err := keyboard.Open(); err != nil {
		// Fallback to basic selection
		return w.handlePhysicalLocationFallback(options, actions)
	}
	defer keyboard.Close()

	currentIndex := 0

	for {
		// Clear screen and show interface
		fmt.Print("\033[H\033[2J")
		
		fmt.Println()
		color.New(color.FgCyan, color.Bold).Println("📍 Physical Location Setup")
		color.New(color.FgYellow).Println("💡 Configure where your repositories will be stored")
		fmt.Println()

		if currentLocation != "" {
			color.New(color.FgGreen).Printf("📁 Current location: %s\n", currentLocation)
		}
		fmt.Println()

		// Display options
		for i, option := range options {
			if i == currentIndex {
				color.New(color.FgCyan, color.Bold).Printf("▶ %s\n", option)
			} else {
				color.New(color.FgWhite, color.Faint).Printf("  %s\n", option)
			}
		}

		fmt.Println()
		color.New(color.FgYellow, color.Faint).Println("Controls: ↑↓=Navigate, ENTER=Select, ESC=Cancel")

		// Get key input
		char, key, err := keyboard.GetKey()
		if err != nil {
			return ActionCancel, err
		}

		switch key {
		case keyboard.KeyArrowUp:
			if currentIndex > 0 {
				currentIndex--
			}
		case keyboard.KeyArrowDown:
			if currentIndex < len(options)-1 {
				currentIndex++
			}
		case keyboard.KeyEnter:
			return w.handlePhysicalLocationAction(actions[currentIndex])
		case keyboard.KeyEsc:
			return w.confirmCancel()
		default:
			if char == 'q' || char == 'Q' {
				return w.confirmCancel()
			}
		}
	}
}

// handlePhysicalLocationAction handles the selected action
func (w *InteractiveWizard) handlePhysicalLocationAction(action string) (NavigationAction, error) {
	switch action {
	case "use_current":
		// Keep current location
		fmt.Println()
		color.New(color.FgGreen).Printf("✅ Using current location: %s\n", w.inventory.PhysicalLocation)
		w.choice.Directory = w.inventory.PhysicalLocation
		return ActionNext, nil

	case "custom":
		// Prompt for custom location
		customPath, err := w.promptCustomPhysicalLocation()
		if err != nil {
			return ActionCancel, err
		}
		return w.updatePhysicalLocation(customPath)

	case "back":
		return ActionBack, nil

	case "cancel":
		return w.confirmCancel()

	default:
		// It's a predefined path
		return w.updatePhysicalLocation(action)
	}
}

// updatePhysicalLocation updates the physical location in both memory and file
func (w *InteractiveWizard) updatePhysicalLocation(newPath string) (NavigationAction, error) {
	// Expand home directory if needed
	if strings.HasPrefix(newPath, "~") {
		homeDir, _ := os.UserHomeDir()
		newPath = filepath.Join(homeDir, newPath[1:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(newPath)
	if err != nil {
		return ActionCancel, fmt.Errorf("invalid path: %v", err)
	}

	fmt.Println()
	color.New(color.FgCyan).Printf("📍 Selected location: %s\n", absPath)

	// Update choice
	w.choice.Directory = absPath

	// Update inventory file if we have inventory context
	if w.inventory != nil && w.inventoryPath != "" {
		// Check if location changed
		if w.inventory.PhysicalLocation != absPath {
			fmt.Println()
			color.New(color.FgYellow).Println("🔄 Updating configuration file with new location...")
			
			if err := UpdatePhysicalLocation(w.inventoryPath, absPath); err != nil {
				color.New(color.FgRed).Printf("⚠️  Warning: Failed to update configuration file: %v\n", err)
				color.New(color.FgYellow).Println("Continuing with new location for this session...")
			} else {
				w.inventory.PhysicalLocation = absPath
				color.New(color.FgGreen).Println("✅ Configuration file updated successfully!")
			}
		}
	}

	fmt.Println()
	return ActionNext, nil
}

// promptCustomPhysicalLocation prompts user for custom location
func (w *InteractiveWizard) promptCustomPhysicalLocation() (string, error) {
	fmt.Println()
	color.New(color.FgYellow).Println("💡 Examples of valid paths:")
	color.New(color.FgWhite).Println("   /Users/vennet/Projects")
	color.New(color.FgWhite).Println("   ~/repositories") 
	color.New(color.FgWhite).Println("   ../my-projects")
	fmt.Println()

	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("path cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter custom physical location path",
		Validate: validate,
		Templates: &promptui.PromptTemplates{
			Prompt:  "{{ . | bold }}{{ \":\" | bold }} ",
			Valid:   "{{ . | green | bold }}{{ \":\" | bold }} ",
			Invalid: "{{ . | red | bold }}{{ \":\" | bold }} ",
			Success: "{{ . | bold }}{{ \":\" | bold }} ",
		},
	}

	return prompt.Run()
}

// handlePhysicalLocationFallback fallback for when keyboard library fails
func (w *InteractiveWizard) handlePhysicalLocationFallback(options []string, actions []string) (NavigationAction, error) {
	prompt := promptui.Select{
		Label: "Choose physical location option",
		Items: options,
		Templates: &promptui.SelectTemplates{
			Active:   "▶ {{ .| cyan }}",
			Inactive: "  {{ . | faint }}",
			Selected: "✅ {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return ActionCancel, err
	}

	return w.handlePhysicalLocationAction(actions[idx])
}