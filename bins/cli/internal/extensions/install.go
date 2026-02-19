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

// defaultOrg is used when resolving shorthand extension names (e.g. "deploy-checker").
const defaultOrg = "nuonco"

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

// isLocalPath returns true if the input looks like a filesystem path.
func isLocalPath(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, ".") || strings.HasPrefix(input, "/") || strings.HasPrefix(input, "~")
}

// Install installs an extension from a GitHub repository or a local directory.
func (m *Manager) Install(repo string) (*InstalledExtension, error) {
	if isLocalPath(repo) {
		return m.InstallLocal(repo)
	}

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
// Accepts: "deploy-checker", "nuon-ext-deploy-checker", "nuonco/nuon-ext-deploy-checker", "myorg/nuon-ext-foo"
func normalizeRepo(input string) (repo, name string, err error) {
	input = strings.TrimSpace(input)

	if strings.Contains(input, "/") {
		// Full repo format: org/repo
		parts := strings.SplitN(input, "/", 2)
		repoName := parts[1]

		if !strings.HasPrefix(repoName, "nuon-ext-") {
			return "", "", fmt.Errorf("extension repository must use nuon-ext- prefix (got: %s)", repoName)
		}

		name = strings.TrimPrefix(repoName, "nuon-ext-")
		return input, name, nil
	}

	// Shorthand: either "nuon-ext-deploy-checker" or "deploy-checker"
	name = strings.TrimPrefix(input, "nuon-ext-")
	repo = defaultOrg + "/nuon-ext-" + name
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

// InstallLocal installs an extension from a local directory.
// The directory must contain a nuon-ext.toml and a pre-built binary named nuon-ext-<name>.
// The binary is symlinked (not copied) so rebuilds take effect immediately.
func (m *Manager) InstallLocal(path string) (*InstalledExtension, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Read and validate nuon-ext.toml from the local directory
	tomlPath := filepath.Join(absPath, "nuon-ext.toml")
	tomlData, err := os.ReadFile(tomlPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read nuon-ext.toml: %w (does the directory contain a nuon-ext.toml?)", err)
	}

	manifest, err := ParseManifest(tomlData)
	if err != nil {
		return nil, fmt.Errorf("invalid extension manifest: %w", err)
	}

	if manifest.Extension.Name == "" {
		return nil, fmt.Errorf("extension.name is required in nuon-ext.toml")
	}
	if manifest.Extension.Description == "" {
		return nil, fmt.Errorf("extension.description is required in nuon-ext.toml")
	}

	name := manifest.Extension.Name

	// Verify the directory name matches the convention
	dirName := filepath.Base(absPath)
	expectedDir := "nuon-ext-" + name
	if dirName != expectedDir {
		return nil, fmt.Errorf("directory name %q does not match extension name %q (expected directory %s)", dirName, name, expectedDir)
	}

	if err := CheckCLIVersion(manifest); err != nil {
		return nil, err
	}

	// Check if already installed
	extDir := filepath.Join(m.dir, "nuon-ext-"+name)
	if _, err := os.Stat(extDir); err == nil {
		return nil, fmt.Errorf("extension %q is already installed (use `nuon ext remove %s` first)", name, name)
	}

	// Find the local binary
	binaryName := extensionBinaryName(name)
	srcBinary := filepath.Join(absPath, binaryName)
	if _, err := os.Stat(srcBinary); err != nil {
		return nil, fmt.Errorf("binary not found at %s (build your extension first)", srcBinary)
	}

	// Create extension directory
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		return nil, fmt.Errorf("unable to create extension directory: %w", err)
	}

	// Symlink the binary so rebuilds take effect immediately
	destBinary := filepath.Join(extDir, binaryName)
	if err := os.Symlink(srcBinary, destBinary); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to symlink binary: %w", err)
	}

	// Copy nuon-ext.toml
	if err := os.WriteFile(filepath.Join(extDir, "nuon-ext.toml"), tomlData, 0o644); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to write nuon-ext.toml: %w", err)
	}

	// Write manifest.json
	now := time.Now().UTC().Format(time.RFC3339)
	installed := &InstalledExtension{
		Name:          name,
		Description:   manifest.Extension.Description,
		Repo:          "local:" + absPath,
		Version:       "dev",
		Tag:           "dev",
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
