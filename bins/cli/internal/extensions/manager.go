package extensions

import "os"

// Manager handles extension lifecycle operations.
type Manager struct {
	dir string // e.g. ~/.nuon/extensions/
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

// List returns all installed extensions.
func (m *Manager) List() ([]InstalledExtension, error) {
	return nil, nil
}

// Get returns a specific installed extension by name.
func (m *Manager) Get(name string) (*InstalledExtension, error) {
	return nil, nil
}

// Install installs an extension from a GitHub repository.
func (m *Manager) Install(repo string) (*InstalledExtension, error) {
	return nil, nil
}

// Remove uninstalls an extension by name.
func (m *Manager) Remove(name string) error {
	return nil
}

// Upgrade upgrades a specific extension to the latest version.
func (m *Manager) Upgrade(name string) error {
	return nil
}

// UpgradeAll upgrades all installed extensions.
func (m *Manager) UpgradeAll() ([]UpgradeResult, error) {
	return nil, nil
}

// Browse lists available extensions from the nuonco GitHub org.
func (m *Manager) Browse() ([]AvailableExtension, error) {
	return nil, nil
}

// Exec runs an installed extension with the given arguments and environment.
func (m *Manager) Exec(name string, args []string, env map[string]string) error {
	return nil
}
