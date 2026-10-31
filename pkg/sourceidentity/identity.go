package sourceidentity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/opencontainers/go-digest"
)

const Version = 1
const Complete = "complete"
const Unknown = "unknown"

const DefaultTerraformVersion = "1.7.5"

func TerraformVersion(version string) string {
	if version == "" {
		return DefaultTerraformVersion
	}
	return version
}

type File struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Executable bool   `json:"executable,omitempty"`
}

type Identity struct {
	Version        int      `json:"version"`
	Checksum       string   `json:"checksum"`
	Coverage       string   `json:"coverage"`
	Reasons        []string `json:"reasons,omitempty"`
	ConfigChecksum string   `json:"config_checksum"`
	Kind           string   `json:"kind"`
	Root           string   `json:"root,omitempty"`
	Repository     string   `json:"repository,omitempty"`
	Commit         string   `json:"commit,omitempty"`
	Digest         string   `json:"digest,omitempty"`
	ChartVersion   string   `json:"chart_version,omitempty"`
	Files          []File   `json:"files,omitempty"`
	Captured       bool     `json:"captured"`
}

func Hash(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return sha256Bytes(data)
}

func sha256Bytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func New(kind, configChecksum string) *Identity {
	return &Identity{Version: Version, Kind: kind, ConfigChecksum: configChecksum, Coverage: Complete}
}

func (i *Identity) MarkUnknown(reason string) {
	i.Coverage = Unknown
	for _, existing := range i.Reasons {
		if existing == reason {
			return
		}
	}
	i.Reasons = append(i.Reasons, reason)
	sort.Strings(i.Reasons)
}

func (i *Identity) Seal() {
	sort.Slice(i.Files, func(a, b int) bool { return i.Files[a].Path < i.Files[b].Path })
	files := i.Files
	if len(files) == 0 {
		files = nil
	}
	i.Checksum = Hash(struct {
		Version                            int
		ConfigChecksum, Kind, Root, Digest string
		Files                              []File
		ChartVersion                       string `json:",omitempty"`
	}{i.Version, i.ConfigChecksum, i.Kind, i.Root, i.Digest, files, i.ChartVersion})
}

func (i *Identity) EligibleChecksum() string {
	if i == nil || i.Version != Version || i.Coverage != Complete || !i.Captured || i.ConfigChecksum == "" || i.Checksum == "" {
		return ""
	}
	if i.Validate() != nil {
		return ""
	}
	return i.Checksum
}

func (i *Identity) Validate() error {
	if i == nil || i.Version != Version || i.ConfigChecksum == "" || (i.Coverage != Complete && i.Coverage != Unknown) {
		return fmt.Errorf("invalid source identity")
	}
	copy := *i
	copy.Files = append([]File(nil), i.Files...)
	copy.Seal()
	if copy.Checksum != i.Checksum {
		return fmt.Errorf("source identity checksum mismatch")
	}
	if i.Coverage == Unknown && len(i.Reasons) == 0 {
		return fmt.Errorf("unknown source coverage requires a reason")
	}
	if i.Coverage == Complete && len(i.Reasons) != 0 {
		return fmt.Errorf("complete source coverage cannot have unknown reasons")
	}
	if len(i.Files) > 20000 {
		return fmt.Errorf("source manifest exceeds file limit")
	}
	if i.Commit != "" && !IsCommit(i.Commit) {
		return fmt.Errorf("source commit must be immutable")
	}
	if strings.HasPrefix(i.Digest, "git:") {
		parts := strings.Split(i.Digest, ":")
		if len(parts) != 3 || !IsCommit(parts[2]) || i.Commit == "" || i.Repository == "" || !i.Captured || len(i.Files) != 0 {
			return fmt.Errorf("invalid Git source identity")
		}
		switch parts[1] {
		case "040000", "120000":
		case "100644", "100755":
			if i.Kind != "helm" {
				return fmt.Errorf("source root is not a directory")
			}
		default:
			return fmt.Errorf("invalid Git source mode")
		}
	}
	if i.Kind == "image" && i.Captured {
		if err := digest.Digest(i.Digest).Validate(); err != nil {
			return fmt.Errorf("invalid image source digest: %w", err)
		}
	}
	seen := make(map[string]bool, len(i.Files))
	for _, f := range i.Files {
		if f.Path == "" || path.IsAbs(f.Path) || path.Clean(f.Path) != f.Path || f.Path == ".." || strings.HasPrefix(f.Path, "../") || seen[f.Path] {
			return fmt.Errorf("invalid source manifest path")
		}
		if _, err := hex.DecodeString(f.SHA256); err != nil || len(f.SHA256) != 64 {
			return fmt.Errorf("invalid source file checksum")
		}
		seen[f.Path] = true
	}
	return nil
}

func Equivalent(before, after *Identity) bool {
	checksum := before.EligibleChecksum()
	return checksum != "" && checksum == after.EligibleChecksum()
}

func Matches(expected, actual *Identity) bool {
	return expected != nil && actual != nil && expected.Validate() == nil && actual.Validate() == nil && expected.Checksum == actual.Checksum && expected.Coverage == actual.Coverage && slices.Equal(expected.Reasons, actual.Reasons) && expected.Commit == actual.Commit && expected.Repository == actual.Repository && expected.Captured == actual.Captured
}

func ValidateTransition(expected, actual *Identity, allowCapture bool) error {
	if actual == nil {
		return nil
	}
	if err := actual.Validate(); err != nil {
		return err
	}
	if expected == nil {
		return fmt.Errorf("source metadata was not prepared")
	}
	if err := expected.Validate(); err != nil {
		return err
	}
	if expected.Captured || !allowCapture {
		if !Matches(expected, actual) {
			return fmt.Errorf("runner source differs from prepared inputs")
		}
		return nil
	}
	if actual.ConfigChecksum != expected.ConfigChecksum || actual.Kind != expected.Kind || actual.Root != expected.Root || !actual.Captured || actual.Commit != expected.Commit || actual.Repository != expected.Repository {
		return fmt.Errorf("invalid source capture")
	}
	if expected.Digest != "" && actual.Digest != expected.Digest {
		return fmt.Errorf("source capture changed a prepared digest")
	}
	if expected.ChartVersion != "" && actual.ChartVersion != expected.ChartVersion {
		return fmt.Errorf("source capture changed a prepared chart version")
	}
	if actual.Kind == "image" {
		if actual.Digest == "" {
			return fmt.Errorf("image digest is required")
		}
	} else if actual.Coverage != Unknown {
		return fmt.Errorf("unprepared dependency closure cannot become complete")
	}
	for _, reason := range expected.Reasons {
		if reason != "source_requires_runner_resolution" && !slices.Contains(actual.Reasons, reason) {
			return fmt.Errorf("source capture lost an unknown input")
		}
	}
	return nil
}

func ResolveImage(expected *Identity, digest string) (*Identity, error) {
	if expected == nil {
		return nil, nil
	}
	if err := expected.Validate(); err != nil {
		return nil, err
	}
	if expected.Kind != "image" {
		return nil, fmt.Errorf("image source identity is required")
	}
	if expected.Digest != "" && expected.Digest != digest {
		return nil, fmt.Errorf("image digest differs from prepared source")
	}
	if expected.Captured {
		return expected, nil
	}
	i := *expected
	i.Files = slices.Clone(expected.Files)
	i.Coverage, i.Reasons = Complete, nil
	for _, reason := range expected.Reasons {
		if reason != "source_requires_runner_resolution" {
			i.MarkUnknown(reason)
		}
	}
	i.Digest = digest
	i.Captured = true
	i.Seal()
	if err := ValidateTransition(expected, &i, true); err != nil {
		return nil, err
	}
	return &i, nil
}
