package extensions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manager handles extension lifecycle operations.
type Manager struct {
	dir string // e.g. ~/.config/nuon/extensions/
}

// New creates a new extension manager.
func New(extensionsDir string) *Manager {
	return &Manager{dir: extensionsDir}
}

// ExtensionDir returns the base extensions directory.
func (m *Manager) ExtensionDir() string {
	return m.dir
}

// EnsureDir creates the extensions directory if it doesn't exist.
func (m *Manager) EnsureDir() error {
	return os.MkdirAll(m.dir, 0o755)
}

// List returns all installed extensions by reading manifest.json from each subdirectory.
func (m *Manager) List() ([]InstalledExtension, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to read extensions directory: %w", err)
	}

	var exts []InstalledExtension
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "nuon-ext-") {
			continue
		}
		// Use os.Stat (not entry.IsDir()) so symlinks to directories — the
		// shape that InstallLocal produces — are followed instead of skipped.
		info, err := os.Stat(filepath.Join(m.dir, entry.Name()))
		if err != nil || !info.IsDir() {
			continue
		}

		manifestPath := filepath.Join(m.dir, entry.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var ext InstalledExtension
		if err := json.Unmarshal(data, &ext); err != nil {
			continue
		}

		exts = append(exts, ext)
	}

	return exts, nil
}

// Get returns a specific installed extension by name.
func (m *Manager) Get(name string) (*InstalledExtension, error) {
	manifestPath := filepath.Join(m.dir, "nuon-ext-"+name, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to read manifest for %q: %w", name, err)
	}

	var ext InstalledExtension
	if err := json.Unmarshal(data, &ext); err != nil {
		return nil, fmt.Errorf("unable to parse manifest for %q: %w", name, err)
	}

	return &ext, nil
}

// Remove uninstalls an extension by removing its directory.
// If the extension directory is a symlink (local install), only the symlink is removed.
func (m *Manager) Remove(name string) error {
	extDir := filepath.Join(m.dir, "nuon-ext-"+name)

	info, err := os.Lstat(extDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("extension %q is not installed", name)
		}
		return err
	}

	// If it's a symlink, just remove the symlink (don't follow it)
	if info.Mode()&os.ModeSymlink != 0 {
		return os.Remove(extDir)
	}

	return os.RemoveAll(extDir)
}

// BinaryPath returns the full path to an extension's binary.
func (m *Manager) BinaryPath(name string) string {
	return filepath.Join(m.dir, "nuon-ext-"+name, extensionBinaryName(name))
}
