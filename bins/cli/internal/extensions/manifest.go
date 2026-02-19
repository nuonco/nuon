package extensions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver/v3"

	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

// ParseManifest parses a nuon-ext.toml file from raw bytes.
func ParseManifest(data []byte) (*ExtensionManifest, error) {
	var m ExtensionManifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unable to parse nuon-ext.toml: %w", err)
	}
	return &m, nil
}

// FetchManifest fetches and parses nuon-ext.toml from a GitHub repo at a given ref.
// It uses the GitHub contents API to get the file content.
func FetchManifest(repo, ref string) (*ExtensionManifest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/nuon-ext.toml", repo)
	if ref != "" {
		url += "?ref=" + ref
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nuon-cli/"+version.Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch nuon-ext.toml: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("nuon-ext.toml not found in %s (ref: %s)", repo, ref)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching nuon-ext.toml from %s", resp.StatusCode, repo)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response: %w", err)
	}

	// GitHub contents API returns JSON with base64-encoded content
	var contents struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.Unmarshal(body, &contents); err != nil {
		return nil, fmt.Errorf("unable to parse GitHub response: %w", err)
	}

	if contents.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected encoding: %s", contents.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(contents.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("unable to decode content: %w", err)
	}

	return ParseManifest(decoded)
}

// ValidateManifest checks that a manifest has required fields and the name matches the repo suffix.
func ValidateManifest(m *ExtensionManifest, repoName string) error {
	if m.Extension.Name == "" {
		return fmt.Errorf("extension.name is required in nuon-ext.toml")
	}
	if m.Extension.Description == "" {
		return fmt.Errorf("extension.description is required in nuon-ext.toml")
	}

	// Check that the name matches the repo suffix (nuon-ext-<name>)
	expectedSuffix := "nuon-ext-" + m.Extension.Name
	// repoName may be "nuonco/nuon-ext-foo" or just "nuon-ext-foo"
	parts := strings.Split(repoName, "/")
	actual := parts[len(parts)-1]
	if actual != expectedSuffix {
		return fmt.Errorf("extension.name %q does not match repo name %q (expected repo %s)", m.Extension.Name, actual, expectedSuffix)
	}

	return nil
}

// CheckCLIVersion checks if the current CLI version meets the extension's minimum version requirement.
func CheckCLIVersion(m *ExtensionManifest) error {
	if m.Extension.MinCLIVersion == "" {
		return nil
	}

	if version.Version == "development" {
		return nil
	}

	minVer, err := semver.NewVersion(m.Extension.MinCLIVersion)
	if err != nil {
		return fmt.Errorf("unable to parse min_cli_version %q: %w", m.Extension.MinCLIVersion, err)
	}

	cliVer, err := semver.NewVersion(version.Version)
	if err != nil {
		return fmt.Errorf("unable to parse CLI version %q: %w", version.Version, err)
	}

	if cliVer.LessThan(minVer) {
		return fmt.Errorf("extension requires CLI version >= %s (current: %s)", m.Extension.MinCLIVersion, version.Version)
	}

	return nil
}
