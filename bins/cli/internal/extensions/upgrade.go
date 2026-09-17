package extensions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Upgrade upgrades a specific installed extension to the latest version.
// If force is true, a compiled binary is re-downloaded even when the tag matches.
// Clone-based extensions (python/script, or releases with no platform asset) are
// always refreshed by re-cloning, matching an unpinned install.
func (m *Manager) Upgrade(name string, force bool) error {
	ext, err := m.Get(name)
	if err != nil {
		return err
	}
	if ext == nil {
		return fmt.Errorf("extension %q is not installed", name)
	}
	if strings.HasPrefix(ext.Repo, "local:") {
		return fmt.Errorf("extension %q is a local install; remove and re-install to update", name)
	}

	extDir := filepath.Join(m.dir, "nuon-ext-"+name)

	release, err := getLatestRelease(ext.Repo)
	var downloadURL string
	if err == nil {
		downloadURL, _ = findReleaseAsset(release, name)
	}

	if downloadURL != "" {
		return m.upgradeByRelease(ext, extDir, release, downloadURL, force)
	}

	if err != nil && ext.Type == ExtTypeBinary {
		return fmt.Errorf("unable to check for updates: %w", err)
	}

	return m.upgradeByClone(ext, extDir, release)
}

func (m *Manager) upgradeByRelease(ext *InstalledExtension, extDir string, release *githubRelease, downloadURL string, force bool) error {
	if !force && release.TagName == ext.Tag {
		return fmt.Errorf("extension %q is already at the latest version (%s)", ext.Name, ext.Tag)
	}

	manifest, err := FetchManifest(ext.Repo, release.TagName)
	if err != nil {
		return fmt.Errorf("unable to fetch manifest for %s: %w", release.TagName, err)
	}

	if err := CheckCLIVersion(manifest); err != nil {
		return err
	}

	binaryName := extensionBinaryName(ext.Name)
	binaryPath := filepath.Join(extDir, binaryName)
	if err := downloadAndExtractBinary(downloadURL, binaryPath, binaryName); err != nil {
		return fmt.Errorf("unable to download updated binary: %w", err)
	}

	tomlData, err := fetchRawManifest(ext.Repo, release.TagName)
	if err == nil {
		os.WriteFile(filepath.Join(extDir, "nuon-ext.toml"), tomlData, 0o644)
	}

	ext.Version = release.TagName
	ext.Tag = release.TagName
	ext.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	ext.Description = manifest.Extension.Description
	ext.MinCLIVersion = manifest.Extension.MinCLIVersion
	ext.RequiresToken = manifest.Extension.Auth.RequiresToken
	ext.RequiresOrg = manifest.Extension.Auth.RequiresOrg
	ext.RequiresApp = manifest.Extension.Auth.RequiresApp
	ext.RequiresInstall = manifest.Extension.Auth.RequiresInstall

	if err := writeManifestJSON(extDir, ext); err != nil {
		return fmt.Errorf("unable to update manifest: %w", err)
	}

	return nil
}

func (m *Manager) upgradeByClone(ext *InstalledExtension, extDir string, release *githubRelease) error {
	manifestRef := ""
	if release != nil {
		manifestRef = release.TagName
	}

	manifest, err := FetchManifest(ext.Repo, manifestRef)
	if err != nil {
		return fmt.Errorf("unable to fetch extension manifest: %w", err)
	}

	if err := CheckCLIVersion(manifest); err != nil {
		return err
	}

	installedAt := ext.InstalledAt
	if err := os.RemoveAll(extDir); err != nil {
		return fmt.Errorf("unable to replace extension directory: %w", err)
	}

	installed, err := m.installByClone(ext.Repo, ext.Name, "", extDir, manifest)
	if err != nil {
		return err
	}

	installed.InstalledAt = installedAt
	if err := writeManifestJSON(extDir, installed); err != nil {
		return fmt.Errorf("unable to update manifest: %w", err)
	}

	return nil
}

// UpgradeAll upgrades all installed extensions and returns results.
func (m *Manager) UpgradeAll() ([]UpgradeResult, error) {
	exts, err := m.List()
	if err != nil {
		return nil, err
	}

	var results []UpgradeResult
	for _, ext := range exts {
		oldVersion := ext.Version
		result := UpgradeResult{
			Name:       ext.Name,
			OldVersion: oldVersion,
		}

		err := m.Upgrade(ext.Name, false)
		if err != nil {
			result.Error = err
			result.NewVersion = oldVersion
		} else {
			// Re-read to get the new version
			updated, _ := m.Get(ext.Name)
			if updated != nil {
				result.NewVersion = updated.Version
			}
		}

		results = append(results, result)
	}

	return results, nil
}
