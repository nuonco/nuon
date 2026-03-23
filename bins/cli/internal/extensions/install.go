package extensions

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
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

// cloneRepo clones a GitHub repository into the given directory.
// If ref is non-empty, it checks out that branch, tag, or commit.
func cloneRepo(repo, destDir, ref string) error {
	url := fmt.Sprintf("https://github.com/%s.git", repo)

	if ref == "" {
		cmd := exec.Command("git", "clone", "--depth", "1", url, destDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Try --branch first (works for branches and tags with shallow clone)
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", ref, url, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// --branch failed (likely a commit SHA), do a full clone + checkout
		os.RemoveAll(destDir)
		cmd = exec.Command("git", "clone", url, destDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		cmd = exec.Command("git", "checkout", ref)
		cmd.Dir = destDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return nil
}

// detectExtType detects the extension type from the contents of a directory.
// It checks for pyproject.toml (python), then nuon-ext-<name> script, then falls back to binary.
func detectExtType(dir, name string) (ExtType, string) {
	// Check for python project
	if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		entrypoint := "nuon-ext-" + name
		// If there's no script entrypoint, it's a pure python project
		if _, err := os.Stat(filepath.Join(dir, entrypoint)); err != nil {
			return ExtTypePython, entrypoint
		}
	}

	// Check for script entrypoint at repo root
	entrypoint := extensionBinaryName(name)
	if _, err := os.Stat(filepath.Join(dir, entrypoint)); err == nil {
		return ExtTypeScript, entrypoint
	}

	return ExtTypeBinary, ""
}

// findReleaseAsset looks for a matching platform binary in a release's assets.
// It checks for bare binaries first, then .tar.gz and .zip archives.
// Returns the download URL and asset name, or empty strings if not found.
func findReleaseAsset(release *githubRelease, name string) (downloadURL, assetName string) {
	baseName := fmt.Sprintf("nuon-ext-%s-%s-%s", name, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		baseName += ".exe"
	}

	// Try exact match first (bare binary)
	for _, asset := range release.Assets {
		if asset.Name == baseName {
			return asset.BrowserDownloadURL, asset.Name
		}
	}

	// Try .tar.gz archive
	for _, asset := range release.Assets {
		if asset.Name == baseName+".tar.gz" {
			return asset.BrowserDownloadURL, asset.Name
		}
	}

	// Try .zip archive
	for _, asset := range release.Assets {
		if asset.Name == baseName+".zip" {
			return asset.BrowserDownloadURL, asset.Name
		}
	}

	return "", baseName
}

// Install installs an extension from a GitHub repository or a local directory.
func (m *Manager) Install(repo string) (*InstalledExtension, error) {
	if isLocalPath(repo) {
		ui.PrintDebug(fmt.Sprintf("detected local path: %s", repo))
		return m.InstallLocal(repo)
	}

	ui.PrintDebug(fmt.Sprintf("installing from GitHub: %s", repo))

	repo, name, ref, err := normalizeRepo(repo)
	if err != nil {
		return nil, err
	}
	ui.PrintDebug(fmt.Sprintf("resolved repo=%s name=%s ref=%s", repo, name, ref))

	// Check if already installed
	extDir := filepath.Join(m.dir, "nuon-ext-"+name)
	if _, err := os.Stat(extDir); err == nil {
		return nil, fmt.Errorf("extension %q is already installed (use `nuon ext upgrade %s` to update)", name, name)
	}

	// Fetch and validate manifest (at the pinned ref if provided)
	ui.PrintDebug(fmt.Sprintf("fetching nuon-ext.toml from %s (ref=%s)", repo, ref))
	manifest, err := FetchManifest(repo, ref)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch extension manifest: %w", err)
	}
	ui.PrintDebug(fmt.Sprintf("manifest: name=%s description=%s", manifest.Extension.Name, manifest.Extension.Description))

	if err := ValidateManifest(manifest, repo); err != nil {
		return nil, fmt.Errorf("invalid extension manifest: %w", err)
	}

	if err := CheckCLIVersion(manifest); err != nil {
		return nil, err
	}

	// When a ref is pinned, try release-based install for that tag first,
	// then fall back to clone (for interpreted extensions or commit SHAs).
	if ref != "" {
		ui.PrintDebug(fmt.Sprintf("ref pinned to %s, trying release first", ref))
		release, err := getReleaseByTag(repo, ref)
		if err == nil && len(release.Assets) > 0 {
			downloadURL, assetName := findReleaseAsset(release, name)
			if downloadURL != "" {
				ui.PrintDebug(fmt.Sprintf("found platform asset %s in release %s", assetName, release.TagName))
				return m.installByRelease(repo, name, extDir, manifest, release, downloadURL)
			}
			ui.PrintDebug(fmt.Sprintf("no matching platform asset in release %s, falling back to clone", release.TagName))
		} else {
			ui.PrintDebug(fmt.Sprintf("no release found for tag %s, falling back to clone", ref))
		}
		return m.installByClone(repo, name, ref, extDir, manifest)
	}

	// Auto-detect: try latest release with platform assets first, fall back to clone
	ui.PrintDebug(fmt.Sprintf("fetching latest release for %s", repo))
	release, err := getLatestRelease(repo)
	if err == nil && len(release.Assets) > 0 {
		downloadURL, assetName := findReleaseAsset(release, name)
		if downloadURL != "" {
			ui.PrintDebug(fmt.Sprintf("found platform asset %s in release %s", assetName, release.TagName))
			return m.installByRelease(repo, name, extDir, manifest, release, downloadURL)
		}
		ui.PrintDebug(fmt.Sprintf("no matching platform asset in release %s, falling back to clone", release.TagName))
	} else {
		ui.PrintDebug("no release with assets found, falling back to clone")
	}

	return m.installByClone(repo, name, "", extDir, manifest)
}

// installByRelease installs a binary extension by downloading a release asset.
func (m *Manager) installByRelease(repo, name, extDir string, manifest *ExtensionManifest, release *githubRelease, downloadURL string) (*InstalledExtension, error) {
	binaryName := extensionBinaryName(name)

	// Create extension directory
	ui.PrintDebug(fmt.Sprintf("creating extension directory: %s", extDir))
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		return nil, fmt.Errorf("unable to create extension directory: %w", err)
	}

	// Download and install binary (handling archives if needed)
	ui.PrintDebug(fmt.Sprintf("downloading binary from %s", downloadURL))
	binaryPath := filepath.Join(extDir, binaryName)
	if err := downloadAndExtractBinary(downloadURL, binaryPath, binaryName); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to download extension binary: %w", err)
	}

	// Write cached nuon-ext.toml
	tomlData, err := fetchRawManifest(repo, "")
	if err == nil {
		os.WriteFile(filepath.Join(extDir, "nuon-ext.toml"), tomlData, 0o644)
	}

	// Write manifest.json
	now := time.Now().UTC().Format(time.RFC3339)
	installed := &InstalledExtension{
		Name:            name,
		Description:     manifest.Extension.Description,
		Repo:            repo,
		Version:         release.TagName,
		Tag:             release.TagName,
		InstalledAt:     now,
		UpdatedAt:       now,
		Binary:          binaryName,
		Type:            ExtTypeBinary,
		Platform:        runtime.GOOS + "/" + runtime.GOARCH,
		MinCLIVersion:   manifest.Extension.MinCLIVersion,
		RequiresToken:   manifest.Extension.Auth.RequiresToken,
		RequiresOrg:     manifest.Extension.Auth.RequiresOrg,
		RequiresApp:     manifest.Extension.Auth.RequiresApp,
		RequiresInstall: manifest.Extension.Auth.RequiresInstall,
	}

	if err := writeManifestJSON(extDir, installed); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to write manifest: %w", err)
	}

	ui.PrintDebug(fmt.Sprintf("installed %s %s to %s", name, release.TagName, extDir))
	return installed, nil
}

// installByClone installs a script or python extension by cloning the repo.
// The type is auto-detected from the repo contents after cloning.
// If ref is non-empty, the clone checks out that specific branch, tag, or commit.
func (m *Manager) installByClone(repo, name, ref, extDir string, manifest *ExtensionManifest) (*InstalledExtension, error) {
	ui.PrintDebug(fmt.Sprintf("cloning %s (ref=%s) into %s", repo, ref, extDir))
	if err := cloneRepo(repo, extDir, ref); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to clone repository: %w", err)
	}

	extType, entrypoint := detectExtType(extDir, name)
	ui.PrintDebug(fmt.Sprintf("detected type=%s entrypoint=%s", extType, entrypoint))

	if extType == ExtTypeScript {
		scriptPath := filepath.Join(extDir, entrypoint)
		if err := os.Chmod(scriptPath, 0o755); err != nil {
			os.RemoveAll(extDir)
			return nil, fmt.Errorf("unable to make script executable: %w", err)
		}
	}

	// Determine version: use pinned ref, or latest release tag, or "latest"
	ver := "latest"
	if ref != "" {
		ver = ref
	} else {
		release, err := getLatestRelease(repo)
		if err == nil {
			ver = release.TagName
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	installed := &InstalledExtension{
		Name:            name,
		Description:     manifest.Extension.Description,
		Repo:            repo,
		Version:         ver,
		Tag:             ver,
		Ref:             ref,
		InstalledAt:     now,
		UpdatedAt:       now,
		Binary:          "",
		Type:            extType,
		Entrypoint:      entrypoint,
		Platform:        runtime.GOOS + "/" + runtime.GOARCH,
		MinCLIVersion:   manifest.Extension.MinCLIVersion,
		RequiresToken:   manifest.Extension.Auth.RequiresToken,
		RequiresOrg:     manifest.Extension.Auth.RequiresOrg,
		RequiresApp:     manifest.Extension.Auth.RequiresApp,
		RequiresInstall: manifest.Extension.Auth.RequiresInstall,
	}

	if err := writeManifestJSON(extDir, installed); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to write manifest: %w", err)
	}

	ui.PrintDebug(fmt.Sprintf("installed %s (%s) %s to %s", name, extType, ver, extDir))
	return installed, nil
}

// normalizeRepo parses and validates the repo input.
// Accepts: "deploy-checker", "nuon-ext-deploy-checker", "nuonco/nuon-ext-deploy-checker", "myorg/nuon-ext-foo"
// An optional @ref suffix pins to a specific branch, tag, or commit (e.g. "nuonco/nuon-ext-demo@main").
func normalizeRepo(input string) (repo, name, ref string, err error) {
	input = strings.TrimSpace(input)

	// Parse @ref suffix before any other processing
	if idx := strings.LastIndex(input, "@"); idx > 0 {
		ref = input[idx+1:]
		input = input[:idx]
		if ref == "" {
			return "", "", "", fmt.Errorf("empty ref after @")
		}
	}

	if strings.Contains(input, "/") {
		// Full repo format: org/repo
		parts := strings.SplitN(input, "/", 2)
		repoName := parts[1]

		if !strings.HasPrefix(repoName, "nuon-ext-") {
			return "", "", "", fmt.Errorf("extension repository must use nuon-ext- prefix (got: %s)", repoName)
		}

		name = strings.TrimPrefix(repoName, "nuon-ext-")
		return input, name, ref, nil
	}

	// Shorthand: either "nuon-ext-deploy-checker" or "deploy-checker"
	name = strings.TrimPrefix(input, "nuon-ext-")
	repo = defaultOrg + "/nuon-ext-" + name
	return repo, name, ref, nil
}

// getReleaseByTag fetches a specific release by tag name from a GitHub repository.
func getReleaseByTag(repo, tag string) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, tag)

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
		return nil, fmt.Errorf("no release found for tag %s in %s", tag, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching release %s for %s", resp.StatusCode, tag, repo)
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

// downloadAndExtractBinary downloads a release asset and extracts the binary if it's an archive.
// Supports bare binaries, .tar.gz archives, and .zip archives.
func downloadAndExtractBinary(url, binaryPath, binaryName string) error {
	if strings.HasSuffix(url, ".tar.gz") {
		return downloadAndExtractTarGz(url, binaryPath, binaryName)
	}

	// Bare binary download
	if err := downloadFile(url, binaryPath); err != nil {
		return err
	}
	return os.Chmod(binaryPath, 0o755)
}

// downloadAndExtractTarGz downloads a .tar.gz archive and extracts the named binary from it.
func downloadAndExtractTarGz(url, binaryPath, binaryName string) error {
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

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to decompress archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("unable to read archive: %w", err)
		}

		// Match the binary by base name (archives may have paths like ./nuon-ext-api)
		if filepath.Base(hdr.Name) == binaryName && hdr.Typeflag == tar.TypeReg {
			out, err := os.Create(binaryPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			return os.Chmod(binaryPath, 0o755)
		}
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
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
	ui.PrintDebug(fmt.Sprintf("resolved local path: %s", absPath))

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Read and validate nuon-ext.toml from the local directory
	tomlPath := filepath.Join(absPath, "nuon-ext.toml")
	ui.PrintDebug(fmt.Sprintf("reading manifest from %s", tomlPath))
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
	ui.PrintDebug(fmt.Sprintf("manifest: name=%s description=%s", name, manifest.Extension.Description))

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

	// Auto-detect extension type from directory contents
	extType, entrypoint := detectExtType(absPath, name)
	ui.PrintDebug(fmt.Sprintf("detected type=%s entrypoint=%s", extType, entrypoint))

	// For binary type, verify the compiled binary exists
	binaryName := ""
	if extType == ExtTypeBinary {
		binaryName = extensionBinaryName(name)
		srcBinary := filepath.Join(absPath, binaryName)
		ui.PrintDebug(fmt.Sprintf("looking for binary at %s", srcBinary))
		if _, err := os.Stat(srcBinary); err != nil {
			return nil, fmt.Errorf("binary not found at %s (build your extension first)", srcBinary)
		}
	}

	// Symlink the entire source directory so all files (scripts/, etc.) are available
	ui.PrintDebug(fmt.Sprintf("symlinking %s -> %s", extDir, absPath))
	if err := os.Symlink(absPath, extDir); err != nil {
		return nil, fmt.Errorf("unable to symlink extension directory: %w", err)
	}

	// Write manifest.json into the source directory (via the symlink)
	now := time.Now().UTC().Format(time.RFC3339)
	installed := &InstalledExtension{
		Name:            name,
		Description:     manifest.Extension.Description,
		Repo:            "local:" + absPath,
		Version:         "dev",
		Tag:             "dev",
		InstalledAt:     now,
		UpdatedAt:       now,
		Binary:          binaryName,
		Type:            extType,
		Entrypoint:      entrypoint,
		Platform:        runtime.GOOS + "/" + runtime.GOARCH,
		MinCLIVersion:   manifest.Extension.MinCLIVersion,
		RequiresToken:   manifest.Extension.Auth.RequiresToken,
		RequiresOrg:     manifest.Extension.Auth.RequiresOrg,
		RequiresApp:     manifest.Extension.Auth.RequiresApp,
		RequiresInstall: manifest.Extension.Auth.RequiresInstall,
	}

	if err := writeManifestJSON(extDir, installed); err != nil {
		os.Remove(extDir) // remove symlink, not source
		return nil, fmt.Errorf("unable to write manifest: %w", err)
	}

	ui.PrintDebug(fmt.Sprintf("installed %s (%s, dev) from %s to %s", name, extType, absPath, extDir))
	return installed, nil
}
