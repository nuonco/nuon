package extensions

const (
	ExtTypeBinary ExtType = "binary"
	ExtTypeScript ExtType = "script"
	ExtTypePython ExtType = "python"
)

type ExtType string

type ExtensionManifest struct {
	Extension ExtensionMeta `toml:"extension"`
}

type ExtensionMeta struct {
	Name          string        `toml:"name"`
	Description   string        `toml:"description"`
	MinCLIVersion string        `toml:"min_cli_version"`
	Auth          ExtensionAuth `toml:"auth"`
}

type ExtensionAuth struct {
	RequiresToken   bool `toml:"requires_token"`
	RequiresOrg     bool `toml:"requires_org"`
	RequiresApp     bool `toml:"requires_app"`
	RequiresInstall bool `toml:"requires_install"`
}

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

type AvailableExtension struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repo        string `json:"repo"`
	LatestTag   string `json:"latest_tag"`
	Installed   bool   `json:"installed"`
}

type UpgradeResult struct {
	Name       string `json:"name"`
	OldVersion string `json:"old_version"`
	NewVersion string `json:"new_version"`
	Error      error  `json:"-"`
}
