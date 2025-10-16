package internal

// Project represents a single project with name and URL
type Project struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ProjectInfo represents extended project information
type ProjectInfo struct {
	Name      string
	URL       string
	GitURL    string
	LocalPath string
	Group     string
}

// Group represents a group that can contain projects and/or subgroups
type Group struct {
	Name     string    `json:"name"`
	Skip     bool      `json:"skip,omitempty"`
	Projects []Project `json:"projects,omitempty"`
	Groups   []Group   `json:"groups,omitempty"`
}

// Inventory represents the root structure of the JSON file
type Inventory struct {
	PhysicalLocation string    `json:"phisical-location,omitempty"`
	Groups           []Group   `json:"groups"`
	Projects         []Project `json:"projects"`
}

// OperationResult represents the result of a clone/pull operation
type OperationResult struct {
	Success   bool
	Project   ProjectInfo
	Message   string
	IsClone   bool
	IsEmpty   bool // True if repository exists but has no commits
	Duration  string
}

// Summary represents the final operation summary
type Summary struct {
	TotalProjects    int
	SuccessCount     int
	FailureCount     int
	ClonedCount      int
	UpdatedCount     int
	SkippedCount     int
	EmptyCount       int // Count of empty repositories (no commits)
	TotalDuration    string
	FailedProjects   []ProjectInfo
	EmptyProjects    []ProjectInfo // Projects that are empty (no commits)
}

// TrackedProject represents a project that has been cloned with tracking info
type TrackedProject struct {
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	Group         string    `json:"group"`
	LocalPath     string    `json:"local_path"`
	GitURL        string    `json:"git_url"`
	LastCloned    string    `json:"last_cloned"`
	LastUpdated   string    `json:"last_updated"`
	LastCommitHash string   `json:"last_commit_hash"`
	Status        string    `json:"status"` // "cloned", "updated", "error"
}

// ProjectTracker represents the tracking file structure
type ProjectTracker struct {
	LastSync        string           `json:"last_sync"`
	OutputDirectory string           `json:"output_directory"`
	InventoryFile   string           `json:"inventory_file"`
	InventoryHash   string           `json:"inventory_hash"`
	Projects        []TrackedProject `json:"projects"`
}

// ProjectDiff represents differences between inventory and tracked state
type ProjectDiff struct {
	NewProjects     []ProjectInfo `json:"new_projects"`
	RemovedProjects []ProjectInfo `json:"removed_projects"`
	ModifiedProjects []ProjectInfo `json:"modified_projects"`
	UnchangedProjects []ProjectInfo `json:"unchanged_projects"`
}