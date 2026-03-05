package extensions

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Upgrade upgrades a specific installed extension to the latest version.
// If force is true, the binary is re-downloaded even when the tag matches.
func (m *Manager) Upgrade(name string, force bool) error {
	ext, err := m.Get(name)
	if err != nil {
		return err
	}
	if ext == nil {
		return fmt.Errorf("extension %q is not installed", name)
	}

	// Check latest release
	release, err := getLatestRelease(ext.Repo)
	if err != nil {
		return fmt.Errorf("unable to check for updates: %w", err)
	}

	if !force && release.TagName == ext.Tag {
		return fmt.Errorf("extension %q is already at the latest version (%s)", name, ext.Tag)
	}

	// Re-fetch manifest from the new tag
	manifest, err := FetchManifest(ext.Repo, release.TagName)
	if err != nil {
		return fmt.Errorf("unable to fetch manifest for %s: %w", release.TagName, err)
	}

	if err := CheckCLIVersion(manifest); err != nil {
		return err
	}

	// Find the right binary asset
	downloadURL, _ := findReleaseAsset(release, name)
	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
	}

	// Download new binary (handling archives if needed)
	extDir := filepath.Join(m.dir, "nuon-ext-"+name)
	binaryName := extensionBinaryName(name)
	binaryPath := filepath.Join(extDir, binaryName)
	if err := downloadAndExtractBinary(downloadURL, binaryPath, binaryName); err != nil {
		return fmt.Errorf("unable to download updated binary: %w", err)
	}

	// Update cached nuon-ext.toml
	tomlData, err := fetchRawManifest(ext.Repo, release.TagName)
	if err == nil {
		os.WriteFile(filepath.Join(extDir, "nuon-ext.toml"), tomlData, 0o644)
	}

	// Update manifest.json
	ext.Version = release.TagName
	ext.Tag = release.TagName
	ext.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	ext.Description = manifest.Extension.Description
	ext.MinCLIVersion = manifest.Extension.MinCLIVersion
	ext.RequiresToken = manifest.Extension.Auth.RequiresToken
	ext.RequiresOrg = manifest.Extension.Auth.RequiresOrg

	if err := writeManifestJSON(extDir, ext); err != nil {
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
