package extensions

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
	RequiresToken bool `toml:"requires_token"`
	RequiresOrg   bool `toml:"requires_org"`
}

// InstalledExtension represents a locally installed extension.
// This is the schema for manifest.json stored in each extension's directory.
type InstalledExtension struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Repo          string `json:"repo"`
	Version       string `json:"version"`
	Tag           string `json:"tag"`
	InstalledAt   string `json:"installed_at"`
	UpdatedAt     string `json:"updated_at"`
	Binary        string `json:"binary"`
	Platform      string `json:"platform"`
	MinCLIVersion string `json:"min_cli_version"`
	RequiresToken bool   `json:"requires_token"`
	RequiresOrg   bool   `json:"requires_org"`
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
