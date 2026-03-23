package extensions

// Extension types determine how an extension is installed and executed.
const (
	ExtTypeBinary ExtType = "binary" // precompiled platform binaries via GitHub Releases
	ExtTypeScript ExtType = "script" // executable script at repo root (bash, etc.)
	ExtTypePython ExtType = "python" // python project managed by uv
)

// ExtType represents the type of an extension.
type ExtType string

// ExtensionManifest represents the parsed nuon-ext.toml file from an extension repo.
type ExtensionManifest struct {
	Extension ExtensionMeta `toml:"extension"`
}

// ExtensionMeta holds the metadata from the [extension] section of nuon-ext.toml.
type ExtensionMeta struct {
	Name          string        `toml:"name"`
	Description   string        `toml:"description"`
	MinCLIVersion string        `toml:"min_cli_version"`
	Auth          ExtensionAuth `toml:"auth"`
}

// ExtensionAuth holds the auth requirements from [extension.auth] in nuon-ext.toml.
type ExtensionAuth struct {
	RequiresToken   bool `toml:"requires_token"`
	RequiresOrg     bool `toml:"requires_org"`
	RequiresApp     bool `toml:"requires_app"`
	RequiresInstall bool `toml:"requires_install"`
}

// InstalledExtension represents a locally installed extension.
// This is the schema for manifest.json stored in each extension's directory.
type InstalledExtension struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Repo            string  `json:"repo"`
	Version         string  `json:"version"`
	Tag             string  `json:"tag"`
	Ref             string  `json:"ref,omitempty"`
	InstalledAt     string  `json:"installed_at"`
	UpdatedAt       string  `json:"updated_at"`
	Binary          string  `json:"binary"`
	Type            ExtType `json:"type"`
	Entrypoint      string  `json:"entrypoint,omitempty"`
	Platform        string  `json:"platform"`
	MinCLIVersion   string  `json:"min_cli_version"`
	RequiresToken   bool    `json:"requires_token"`
	RequiresOrg     bool    `json:"requires_org"`
	RequiresApp     bool    `json:"requires_app"`
	RequiresInstall bool    `json:"requires_install"`
}

// AvailableExtension represents an extension available for installation from GitHub.
type AvailableExtension struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repo        string `json:"repo"`
	LatestTag   string `json:"latest_tag"`
	Installed   bool   `json:"installed"`
}

// UpgradeResult represents the result of upgrading a single extension.
type UpgradeResult struct {
	Name       string `json:"name"`
	OldVersion string `json:"old_version"`
	NewVersion string `json:"new_version"`
	Error      error  `json:"-"`
}
