package extensions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

// extensionTopic is the GitHub topic that official extensions must have to appear in browse results.
const extensionTopic = "nuon-extensions"

// githubRepo represents a GitHub repository from the search API.
type githubRepo struct {
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
}

// githubSearchResult represents the GitHub repo search API response.
type githubSearchResult struct {
	Items []githubRepo `json:"items"`
}

// Browse lists official extensions from the nuonco GitHub org.
// Only repositories with the "nuon-cli-extension" topic are shown.
func (m *Manager) Browse(_ string) ([]AvailableExtension, error) {
	repos, err := searchExtensionRepos(defaultOrg)
	if err != nil {
		return nil, fmt.Errorf("unable to search for extensions: %w", err)
	}

	// Get list of installed extensions to mark installed ones
	installed, _ := m.List()
	installedMap := make(map[string]bool)
	for _, ext := range installed {
		installedMap[ext.Name] = true
	}

	var available []AvailableExtension
	for _, repo := range repos {
		if !strings.HasPrefix(repo.Name, "nuon-ext-") {
			continue
		}
		if !hasTopic(repo.Topics, extensionTopic) {
			continue
		}

		name := strings.TrimPrefix(repo.Name, "nuon-ext-")

		ext := AvailableExtension{
			Name:        name,
			Description: repo.Description,
			Repo:        repo.FullName,
			Installed:   installedMap[name],
		}

		// Try to get latest release tag
		release, err := getLatestRelease(repo.FullName)
		if err == nil {
			ext.LatestTag = release.TagName
		}

		available = append(available, ext)
	}

	return available, nil
}

// hasTopic checks if a topic list contains a specific topic.
func hasTopic(topics []string, target string) bool {
	for _, t := range topics {
		if t == target {
			return true
		}
	}
	return false
}

// searchExtensionRepos searches for nuon-ext-* repos in the given org that have the extension topic.
func searchExtensionRepos(org string) ([]githubRepo, error) {
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=nuon-ext-+in:name+org:%s+topic:%s&per_page=100", org, extensionTopic)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nuon-cli/"+version.Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from GitHub search", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result githubSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Items, nil
}
