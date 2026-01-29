package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
)

const (
	RawContentURL = "https://raw.githubusercontent.com/nuonco/example-app-configs/main"
	GitHubAPIURL  = "https://api.github.com/repos/nuonco/example-app-configs/contents"

	defaultTimeout = 30 * time.Second
)

// AllowedTemplates defines the templates that can be used
var AllowedTemplates = map[string]bool{
	"eks-simple": true,
	"aws-lambda": true,
	"httpbin":    true,
	"coder":      true,
	"mattermost": true,
}

// Fetcher fetches template configs from GitHub
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new GitHub fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// GitHubContent represents a file or directory in GitHub
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url,omitempty"`
}

// FetchTemplateConfig fetches and parses a template from GitHub into an AppConfig
func (f *Fetcher) FetchTemplateConfig(ctx context.Context, template string) (*config.AppConfig, error) {
	if !AllowedTemplates[template] {
		allowedList := make([]string, 0, len(AllowedTemplates))
		for t := range AllowedTemplates {
			allowedList = append(allowedList, t)
		}
		return nil, fmt.Errorf("template must be one of: %s", strings.Join(allowedList, ", "))
	}

	// Create a temp directory to store the fetched files
	tempDir, err := os.MkdirTemp("", "nuon-template-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Fetch all files from the template directory
	if err := f.fetchDirectory(ctx, template, tempDir); err != nil {
		return nil, fmt.Errorf("failed to fetch template from GitHub: %w", err)
	}

	// Parse the directory using the existing parse package
	cfg, err := parse.ParseDir(ctx, parse.ParseConfig{
		Dirname:       tempDir,
		V:             validator.New(),
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse template config: %w", err)
	}

	return cfg, nil
}

// fetchDirectory recursively fetches all files from a GitHub directory
func (f *Fetcher) fetchDirectory(ctx context.Context, remotePath, localDir string) error {
	// Get directory listing from GitHub API
	apiURL := fmt.Sprintf("%s/%s", GitHubAPIURL, remotePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch directory listing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var contents []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return fmt.Errorf("failed to decode directory listing: %w", err)
	}

	for _, item := range contents {
		localPath := filepath.Join(localDir, item.Name)

		if item.Type == "dir" {
			// Create local directory and recurse
			if err := os.MkdirAll(localPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", localPath, err)
			}
			if err := f.fetchDirectory(ctx, item.Path, localPath); err != nil {
				return err
			}
		} else if item.Type == "file" {
			// Fetch and save file
			if err := f.fetchFile(ctx, item.DownloadURL, localPath); err != nil {
				return fmt.Errorf("failed to fetch file %s: %w", item.Name, err)
			}
		}
	}

	return nil
}

// fetchFile fetches a single file from GitHub and saves it locally
func (f *Fetcher) fetchFile(ctx context.Context, url, localPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch file, status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	if err := os.WriteFile(localPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
