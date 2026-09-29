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

const defaultOrg = "nuonco"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func isLocalPath(input string) bool {
	input = strings.TrimSpace(input)
	return strings.HasPrefix(input, ".") || strings.HasPrefix(input, "/") || strings.HasPrefix(input, "~")
}

func cloneRepo(repo, destDir, ref string) error {
	url := fmt.Sprintf("https://github.com/%s.git", repo)

	if ref == "" {
		cmd := exec.Command("git", "clone", "--depth", "1", url, destDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", ref, url, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
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

func detectExtType(dir, name string) (ExtType, string) {
	if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		entrypoint := "nuon-ext-" + name
		if _, err := os.Stat(filepath.Join(dir, entrypoint)); err != nil {
			return ExtTypePython, entrypoint
		}
	}

	entrypoint := extensionBinaryName(name)
	if _, err := os.Stat(filepath.Join(dir, entrypoint)); err == nil {
		return ExtTypeScript, entrypoint
	}

	return ExtTypeBinary, ""
}

func findReleaseAsset(release *githubRelease, name string) (downloadURL, assetName string) {
	baseName := fmt.Sprintf("nuon-ext-%s-%s-%s", name, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		baseName += ".exe"
	}

	for _, asset := range release.Assets {
		if asset.Name == baseName {
			return asset.BrowserDownloadURL, asset.Name
		}
	}

	for _, asset := range release.Assets {
		if asset.Name == baseName+".tar.gz" {
			return asset.BrowserDownloadURL, asset.Name
		}
	}

	for _, asset := range release.Assets {
		if asset.Name == baseName+".zip" {
			return asset.BrowserDownloadURL, asset.Name
		}
	}

	return "", baseName
}

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

	extDir := filepath.Join(m.dir, "nuon-ext-"+name)
	if _, err := os.Stat(extDir); err == nil {
		return nil, fmt.Errorf("extension %q is already installed (use `nuon ext upgrade %s` to update)", name, name)
	}

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

func (m *Manager) installByRelease(repo, name, extDir string, manifest *ExtensionManifest, release *githubRelease, downloadURL string) (*InstalledExtension, error) {
	binaryName := extensionBinaryName(name)

	ui.PrintDebug(fmt.Sprintf("creating extension directory: %s", extDir))
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		return nil, fmt.Errorf("unable to create extension directory: %w", err)
	}

	ui.PrintDebug(fmt.Sprintf("downloading binary from %s", downloadURL))
	binaryPath := filepath.Join(extDir, binaryName)
	if err := downloadAndExtractBinary(downloadURL, binaryPath, binaryName); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to download extension binary: %w", err)
	}

	tomlData, err := fetchRawManifest(repo, "")
	if err == nil {
		os.WriteFile(filepath.Join(extDir, "nuon-ext.toml"), tomlData, 0o644)
	}

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

func normalizeRepo(input string) (repo, name, ref string, err error) {
	input = strings.TrimSpace(input)

	if idx := strings.LastIndex(input, "@"); idx > 0 {
		ref = input[idx+1:]
		input = input[:idx]
		if ref == "" {
			return "", "", "", fmt.Errorf("empty ref after @")
		}
	}

	if strings.Contains(input, "/") {
		parts := strings.SplitN(input, "/", 2)
		repoName := parts[1]

		if !strings.HasPrefix(repoName, "nuon-ext-") {
			return "", "", "", fmt.Errorf("extension repository must use nuon-ext- prefix (got: %s)", repoName)
		}

		name = strings.TrimPrefix(repoName, "nuon-ext-")
		return input, name, ref, nil
	}

	name = strings.TrimPrefix(input, "nuon-ext-")
	repo = defaultOrg + "/nuon-ext-" + name
	return repo, name, ref, nil
}

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

func downloadAndExtractBinary(url, binaryPath, binaryName string) error {
	if strings.HasSuffix(url, ".tar.gz") {
		return downloadAndExtractTarGz(url, binaryPath, binaryName)
	}

	if err := downloadFile(url, binaryPath); err != nil {
		return err
	}
	return os.Chmod(binaryPath, 0o755)
}

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

func writeManifestJSON(extDir string, ext *InstalledExtension) error {
	data, err := json.MarshalIndent(ext, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(extDir, "manifest.json"), data, 0o644)
}

func extensionBinaryName(name string) string {
	binName := "nuon-ext-" + name
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	return binName
}

func (m *Manager) InstallLocal(path string) (*InstalledExtension, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve path: %w", err)
	}
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		if home != "" {
			absPath = filepath.Join(home, path[1:])
		}
	}
	ui.PrintDebug(fmt.Sprintf("resolved local path: %s", absPath))

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %w", err)
	}

	if !info.IsDir() {
		return m.installLocalBinary(absPath, info)
	}

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

	dirName := filepath.Base(absPath)
	expectedDir := "nuon-ext-" + name
	if dirName != expectedDir {
		return nil, fmt.Errorf("directory name %q does not match extension name %q (expected directory %s)", dirName, name, expectedDir)
	}

	if err := CheckCLIVersion(manifest); err != nil {
		return nil, err
	}

	extDir := filepath.Join(m.dir, "nuon-ext-"+name)
	if _, err := os.Stat(extDir); err == nil {
		return nil, fmt.Errorf("extension %q is already installed (use `nuon ext remove %s` first)", name, name)
	}

	extType, entrypoint := detectExtType(absPath, name)
	ui.PrintDebug(fmt.Sprintf("detected type=%s entrypoint=%s", extType, entrypoint))

	binaryName := ""
	if extType == ExtTypeBinary {
		binaryName = extensionBinaryName(name)
		srcBinary := filepath.Join(absPath, binaryName)
		ui.PrintDebug(fmt.Sprintf("looking for binary at %s", srcBinary))
		if _, err := os.Stat(srcBinary); err != nil {
			return nil, fmt.Errorf("binary not found at %s (build your extension first)", srcBinary)
		}
	}

	ui.PrintDebug(fmt.Sprintf("symlinking %s -> %s", extDir, absPath))
	if err := os.Symlink(absPath, extDir); err != nil {
		return nil, fmt.Errorf("unable to symlink extension directory: %w", err)
	}

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
		os.Remove(extDir)
		return nil, fmt.Errorf("unable to write manifest: %w", err)
	}

	ui.PrintDebug(fmt.Sprintf("installed %s (%s, dev) from %s to %s", name, extType, absPath, extDir))
	return installed, nil
}

func (m *Manager) installLocalBinary(absPath string, info os.FileInfo) (*InstalledExtension, error) {
	baseName := filepath.Base(absPath)

	cleanName := strings.TrimSuffix(baseName, ".exe")

	if !strings.HasPrefix(cleanName, "nuon-ext-") {
		return nil, fmt.Errorf("binary name %q must use the nuon-ext-<name> convention (e.g. nuon-ext-linter)", baseName)
	}

	name := strings.TrimPrefix(cleanName, "nuon-ext-")
	ui.PrintDebug(fmt.Sprintf("detected extension name %q from binary %s", name, baseName))

	if info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("binary %s is not executable", absPath)
	}

	extDir := filepath.Join(m.dir, "nuon-ext-"+name)
	if _, err := os.Stat(extDir); err == nil {
		return nil, fmt.Errorf("extension %q is already installed (use `nuon ext remove %s` first)", name, name)
	}

	ui.PrintDebug(fmt.Sprintf("creating extension directory: %s", extDir))
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		return nil, fmt.Errorf("unable to create extension directory: %w", err)
	}

	binaryName := extensionBinaryName(name)
	destPath := filepath.Join(extDir, binaryName)
	ui.PrintDebug(fmt.Sprintf("copying binary %s -> %s", absPath, destPath))

	srcFile, err := os.Open(absPath)
	if err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to read binary: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(destPath)
	if err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to create binary: %w", err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to copy binary: %w", err)
	}
	dstFile.Close()

	if err := os.Chmod(destPath, 0o755); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to make binary executable: %w", err)
	}

	description := fmt.Sprintf("Extension: %s", name)
	var requiresToken, requiresOrg, requiresApp, requiresInstall bool

	tomlPath := filepath.Join(filepath.Dir(absPath), "nuon-ext.toml")
	if tomlData, err := os.ReadFile(tomlPath); err == nil {
		ui.PrintDebug(fmt.Sprintf("found nuon-ext.toml at %s", tomlPath))
		if manifest, err := ParseManifest(tomlData); err == nil {
			if manifest.Extension.Description != "" {
				description = manifest.Extension.Description
			}
			requiresToken = manifest.Extension.Auth.RequiresToken
			requiresOrg = manifest.Extension.Auth.RequiresOrg
			requiresApp = manifest.Extension.Auth.RequiresApp
			requiresInstall = manifest.Extension.Auth.RequiresInstall
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	installed := &InstalledExtension{
		Name:            name,
		Description:     description,
		Repo:            "local:" + absPath,
		Version:         "dev",
		Tag:             "dev",
		InstalledAt:     now,
		UpdatedAt:       now,
		Binary:          binaryName,
		Type:            ExtTypeBinary,
		Platform:        runtime.GOOS + "/" + runtime.GOARCH,
		RequiresToken:   requiresToken,
		RequiresOrg:     requiresOrg,
		RequiresApp:     requiresApp,
		RequiresInstall: requiresInstall,
	}

	if err := writeManifestJSON(extDir, installed); err != nil {
		os.RemoveAll(extDir)
		return nil, fmt.Errorf("unable to write manifest: %w", err)
	}

	ui.PrintDebug(fmt.Sprintf("installed %s (binary, dev) from %s to %s", name, absPath, extDir))
	return installed, nil
}
