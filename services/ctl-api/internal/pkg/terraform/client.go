package terraform

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client interface {
	GetLatestVersion() (string, error)
}

// This client ONLY fetches the latest version from github releases
// It doesn't actually execute terraform commands
type TerraformClient struct {
	cache      Cache
	httpClient *http.Client
}

func New() Client {
	client := TerraformClient{
		cache: Cache{
			ttl:   60 * time.Minute,
			store: make(map[string]CacheEntry),
		},
		httpClient: http.DefaultClient,
	}
	return &client
}

func (client *TerraformClient) GetLatestVersion() (string, error) {
	if version, ok := client.cache.Get("version"); ok {
		return version, nil
	}

	version, err := client.FetchVersion()
	if err != nil {
		return "", err
	}

	client.cache.Set("version", version)
	return version, nil
}

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func (client *TerraformClient) FetchVersion() (string, error) {
	url := "https://api.github.com/repos/hashicorp/terraform/releases/latest"
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode JSON: %w", err)
	}

	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}
