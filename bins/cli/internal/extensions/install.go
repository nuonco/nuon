package extensions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

const allowedOrg = "nuonco"

// githubRelease represents a GitHub release from the releases API.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

// githubAsset represents a release asset.
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Install installs an extension from a GitHub repository.
func (m *Manager) Install(repo string) (*InstalledExtension, error) {
	repo, name, err := normalizeRepo(repo)
	if err != nil {
		return nil, err
	}

	// Check if already installed
	extDir := filepath.Join(m.dir, "nuon-ext-"+name)
	if _, err := os.Stat(extDir); err == nil {
		return nil, fmt.Errorf("extension %q is already installed (use `nuon ext upgrade %s` to update)", name, name)
	}

	// Fetch and validate manifest
	manifest, err := FetchManifest(repo, "")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch extension manifest: %w", err)
	}

	if err := ValidateManifest(manifest, repo); err != nil {
		return nil, fmt.Errorf("invalid extension manifest: %w", err)
	}

	if err := CheckCLIVersion(manifest); err != nil {
		return nil, err
	}

	// Get latest release
	release, err := getLatestRelease(repo)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest release: %w", err)
	}

	// Find the right binary for this platform
	binaryName := extensionBinaryName(name)
	assetName := fmt.Sprintf("nuon-ext-%s-%s-%s", name, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no binary found for %s/%s in release %s (expected asset: %s)", runtime.GOOS, runtime.GOARCH, release.TagName, assetName)
	}

	// Create extension directory
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		return nil, fmt.Errorf("unable to create extension directory: %w", err)
	}

	// Download binary
	binaryPath := filepath.Join(extDir, binaryName)
	if err := downloadFile(downloadURL, binaryPath); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to download extension binary: %w", err)
	}

	// Make binary executable
	if err := os.Chmod(binaryPath, 0o755); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to make binary executable: %w", err)
	}

	// Write cached nuon-ext.toml
	tomlData, err := fetchRawManifest(repo, "")
	if err == nil {
		os.WriteFile(filepath.Join(extDir, "nuon-ext.toml"), tomlData, 0o644)
	}

	// Write manifest.json
	now := time.Now().UTC().Format(time.RFC3339)
	installed := &InstalledExtension{
		Name:          name,
		Description:   manifest.Extension.Description,
		Repo:          repo,
		Version:       release.TagName,
		Tag:           release.TagName,
		InstalledAt:   now,
		UpdatedAt:     now,
		Binary:        binaryName,
		Platform:      runtime.GOOS + "/" + runtime.GOARCH,
		MinCLIVersion: manifest.Extension.MinCLIVersion,
		RequiresToken: manifest.Extension.Auth.RequiresToken,
		RequiresOrg:   manifest.Extension.Auth.RequiresOrg,
	}

	if err := writeManifestJSON(extDir, installed); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to write manifest: %w", err)
	}

	return installed, nil
}

// normalizeRepo parses and validates the repo input.
// Accepts: "deploy-checker", "nuon-ext-deploy-checker", "nuonco/nuon-ext-deploy-checker"
func normalizeRepo(input string) (repo, name string, err error) {
	input = strings.TrimSpace(input)

	if strings.Contains(input, "/") {
		// Full repo format: org/repo
		parts := strings.SplitN(input, "/", 2)
		org := parts[0]
		repoName := parts[1]

		if org != allowedOrg {
			return "", "", fmt.Errorf("extensions must be from the %s GitHub organization (got: %s)", allowedOrg, org)
		}

		if !strings.HasPrefix(repoName, "nuon-ext-") {
			return "", "", fmt.Errorf("extension repository must use nuon-ext- prefix (got: %s)", repoName)
		}

		name = strings.TrimPrefix(repoName, "nuon-ext-")
		return input, name, nil
	}

	// Shorthand: either "nuon-ext-deploy-checker" or "deploy-checker"
	name = strings.TrimPrefix(input, "nuon-ext-")
	repo = allowedOrg + "/nuon-ext-" + name
	return repo, name, nil
}

// getLatestRelease fetches the latest release from a GitHub repository.
func getLatestRelease(repo string) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

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

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found for %s", repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching releases for %s", resp.StatusCode, repo)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadFile downloads a URL to a local file path.
func downloadFile(url, destPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "nuon-cli/"+version.Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d downloading %s", resp.StatusCode, url)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// fetchRawManifest fetches the raw nuon-ext.toml content from a GitHub repo.
func fetchRawManifest(repo, ref string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/nuon-ext.toml", repo, "HEAD")
	if ref != "" {
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/nuon-ext.toml", repo, ref)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "nuon-cli/"+version.Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to fetch raw manifest: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// writeManifestJSON writes the InstalledExtension to manifest.json in the extension directory.
func writeManifestJSON(extDir string, ext *InstalledExtension) error {
	data, err := json.MarshalIndent(ext, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(extDir, "manifest.json"), data, 0o644)
}

// extensionBinaryName returns the binary name for an extension.
func extensionBinaryName(name string) string {
	binName := "nuon-ext-" + name
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	return binName
}
