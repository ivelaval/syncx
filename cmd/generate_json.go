package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"olive-clone-assistant-v2/internal"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	generateToken    string
	generateOut      string
	generateHost     string
	generateGroup    string
	generateFilename string
)

// gitlabProject represents a GitLab project from the API
type gitlabProject struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	DefaultBranch     string `json:"default_branch"`
	Description       string `json:"description"`
}

// gitlabGroup represents a GitLab group from the API
type gitlabGroup struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullPath string `json:"full_path"`
	WebURL   string `json:"web_url"`
}

var generateJsonCmd = &cobra.Command{
	Use:   "generate-json",
	Short: "🔧 Generate inventory JSON from a GitLab group",
	Long: color.New(color.FgMagenta, color.Bold).Sprint(`
🔧 Generate JSON Command
========================

Connects to the GitLab API and generates a projects-inventory.json
file compatible with syncx, reflecting the group/project hierarchy.

Examples:
  syncx generate-json --token glpat-xxxx --group my-org/my-group --out ~/
  syncx generate-json --token glpat-xxxx --group my-org/my-group --out ~/ --host gitlab.mycompany.com
`),
	Run: runGenerateJson,
}

func init() {
	generateJsonCmd.Flags().StringVar(&generateToken, "token", "", "GitLab Personal Access Token (required)")
	generateJsonCmd.Flags().StringVar(&generateGroup, "group", "", "GitLab group path, e.g. myorg/mygroup (required)")
	generateJsonCmd.Flags().StringVar(&generateOut, "out", ".", "Output directory for the generated JSON file")
	generateJsonCmd.Flags().StringVar(&generateHost, "host", "gitlab.com", "GitLab host (default: gitlab.com)")
	generateJsonCmd.Flags().StringVar(&generateFilename, "filename", "projects-inventory.json", "Output filename")
	generateJsonCmd.MarkFlagRequired("token")
	generateJsonCmd.MarkFlagRequired("group")

	rootCmd.AddCommand(generateJsonCmd)
}

func runGenerateJson(cmd *cobra.Command, args []string) {
	bold := color.New(color.FgCyan, color.Bold)
	success := color.New(color.FgGreen, color.Bold)
	errColor := color.New(color.FgRed, color.Bold)

	bold.Printf("\n🔧 Generating inventory from GitLab group: %s\n\n", generateGroup)

	// Resolve group ID
	groupID, err := resolveGroupID(generateHost, generateToken, generateGroup)
	if err != nil {
		errColor.Printf("❌ Failed to resolve group '%s': %v\n", generateGroup, err)
		os.Exit(1)
	}

	color.New(color.FgYellow).Printf("   Group ID: %d\n", groupID)

	// Build inventory tree
	inventory, totalGroups, totalProjects, err := buildInventory(generateHost, generateToken, groupID)
	if err != nil {
		errColor.Printf("❌ Failed to build inventory: %v\n", err)
		os.Exit(1)
	}

	// Write JSON file
	outPath, err := resolveOutputPath(generateOut, generateFilename)
	if err != nil {
		errColor.Printf("❌ Failed to resolve output path: %v\n", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		errColor.Printf("❌ Failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		errColor.Printf("❌ Failed to write file: %v\n", err)
		os.Exit(1)
	}

	success.Printf("\n✅ Inventory generated successfully!\n")
	fmt.Printf("   📁 File:     %s\n", outPath)
	fmt.Printf("   📦 Groups:   %d\n", totalGroups)
	fmt.Printf("   🗂️  Projects: %d\n\n", totalProjects)
}

func resolveGroupID(host, token, groupPath string) (int, error) {
	encoded := ""
	for i, c := range groupPath {
		if c == '/' {
			if i > 0 {
				encoded += "%2F"
			}
		} else {
			encoded += string(c)
		}
	}
	// URL-encode the full path
	urlEncoded := urlEncodePath(groupPath)
	url := fmt.Sprintf("https://%s/api/v4/groups/%s", host, urlEncoded)

	var group gitlabGroup
	if err := gitlabGet(url, token, &group); err != nil {
		return 0, err
	}
	return group.ID, nil
}

func buildInventory(host, token string, groupID int) (internal.Inventory, int, int, error) {
	totalGroups := 0
	totalProjects := 0

	rootGroup, g, p, err := fetchGroupTree(host, token, groupID)
	if err != nil {
		return internal.Inventory{}, 0, 0, err
	}
	totalGroups += g
	totalProjects += p

	inventory := internal.Inventory{
		Groups:   rootGroup.Groups,
		Projects: rootGroup.Projects,
	}

	return inventory, totalGroups, totalProjects, nil
}

func fetchGroupTree(host, token string, groupID int) (internal.Group, int, int, error) {
	totalGroups := 0
	totalProjects := 0

	node := internal.Group{}

	// Fetch direct projects (not recursive)
	projects, err := gitlabGetAll[gitlabProject](
		fmt.Sprintf("https://%s/api/v4/groups/%d/projects?include_subgroups=false&per_page=100", host, groupID),
		token,
	)
	if err != nil {
		return node, 0, 0, fmt.Errorf("fetching projects for group %d: %w", groupID, err)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	for _, p := range projects {
		node.Projects = append(node.Projects, internal.Project{
			Name:          p.Name,
			URL:           p.SSHURLToRepo,
			HTTPUrl:       p.HTTPURLToRepo,
			DefaultBranch: p.DefaultBranch,
			Description:   p.Description,
		})
		totalProjects++
	}

	// Fetch subgroups
	subgroups, err := gitlabGetAll[gitlabGroup](
		fmt.Sprintf("https://%s/api/v4/groups/%d/subgroups?per_page=100", host, groupID),
		token,
	)
	if err != nil {
		return node, 0, 0, fmt.Errorf("fetching subgroups for group %d: %w", groupID, err)
	}

	sort.Slice(subgroups, func(i, j int) bool {
		return subgroups[i].Name < subgroups[j].Name
	})

	for _, sg := range subgroups {
		child, g, p, err := fetchGroupTree(host, token, sg.ID)
		if err != nil {
			return node, 0, 0, err
		}
		child.Name = sg.Name
		totalGroups += g + 1
		totalProjects += p
		node.Groups = append(node.Groups, child)
	}

	return node, totalGroups, totalProjects, nil
}

// gitlabGetAll fetches all pages from a paginated GitLab API endpoint
func gitlabGetAll[T any](baseURL, token string) ([]T, error) {
	var all []T
	page := 1

	for {
		sep := "?"
		if len(baseURL) > 0 {
			for _, c := range baseURL {
				if c == '?' {
					sep = "&"
					break
				}
			}
		}
		url := fmt.Sprintf("%s%spage=%d", baseURL, sep, page)

		var items []T
		if err := gitlabGet(url, token, &items); err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}
		all = append(all, items...)
		page++
	}

	return all, nil
}

// gitlabGet performs a GET request to the GitLab API and decodes the response
func gitlabGet(url, token string, dest any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("not found (404): %s", url)
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("unauthorized (%d): check your token", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

// urlEncodePath encodes a group path for use in GitLab API URLs
func urlEncodePath(path string) string {
	result := ""
	for _, c := range path {
		switch {
		case c == '/':
			result += "%2F"
		case c == ' ':
			result += "%20"
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.':
			result += string(c)
		default:
			result += "%" + strconv.FormatInt(int64(c), 16)
		}
	}
	return result
}

// resolveOutputPath resolves and creates the output directory if needed
func resolveOutputPath(outDir, filename string) (string, error) {
	// Expand ~ to home directory
	if len(outDir) > 0 && outDir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		outDir = filepath.Join(home, outDir[1:])
	}

	absDir, err := filepath.Abs(outDir)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(absDir, filename), nil
}
