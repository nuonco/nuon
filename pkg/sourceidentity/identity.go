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
	"time"

	"github.com/opencontainers/go-digest"
)

const Version = 1

func CleanRoot(root string) string {
	root = path.Clean(strings.TrimLeft(root, "/"))
	if root == "." {
		return ""
	}
	return root
}

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

// Input is a repository path outside Root that the build read (or only
// statted, when Stat is set), with its Git entry at Commit. Entry is empty
// when the path is absent; stat inputs keep only the mode unless the entry is
// a symlink or submodule.
type Input struct {
	Path  string `json:"path"`
	Stat  bool   `json:"stat,omitempty"`
	Entry string `json:"entry,omitempty"`
}

type Reads struct {
	Read []string `json:"read,omitempty"`
	Stat []string `json:"stat,omitempty"`
}

const MaxInputs = 1000

// How a Git source's commit was chosen. Provenance is recorded for display and
// is not part of the checksum.
const (
	ResolvedFromCommit    = "commit"
	ResolvedFromBranchRun = "branch_run"
	ResolvedFromLookup    = "lookup"
)

const InputsOverLimit = "source_inputs_over_limit"

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
	Inputs         []Input  `json:"inputs,omitempty"`
	InputsRecorded bool     `json:"inputs_recorded,omitempty"`

	Ref          string     `json:"ref,omitempty"`
	ResolvedFrom string     `json:"resolved_from,omitempty"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
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
	sort.Slice(i.Inputs, func(a, b int) bool { return i.Inputs[a].Path < i.Inputs[b].Path })
	files := i.Files
	if len(files) == 0 {
		files = nil
	}
	i.Checksum = Hash(struct {
		Version                            int
		ConfigChecksum, Kind, Root, Digest string
		Files                              []File
		ChartVersion                       string  `json:",omitempty"`
		Inputs                             []Input `json:",omitempty"`
		InputsRecorded                     bool    `json:",omitempty"`
	}{i.Version, i.ConfigChecksum, i.Kind, i.Root, i.Digest, files, i.ChartVersion, i.Inputs, i.InputsRecorded})
}

// WithInputs records the entries a build read outside Root.
func (i *Identity) WithInputs(inputs []Input) (*Identity, error) {
	out := *i.WithoutInputs()
	out.Inputs = slices.Clone(inputs)
	out.InputsRecorded = true
	out.Seal()
	return &out, out.Validate()
}

// OverInputLimit records that the build read more than MaxInputs paths
// outside Root, so it can never be reused.
func (i *Identity) OverInputLimit() (*Identity, error) {
	out := *i.WithoutInputs()
	out.InputsRecorded = true
	out.MarkUnknown(InputsOverLimit)
	out.Seal()
	return &out, out.Validate()
}

func (i *Identity) WithoutInputs() *Identity {
	out := *i
	out.Files = slices.Clone(i.Files)
	out.Reasons = slices.Clone(i.Reasons)
	if i.Inputs == nil && !i.InputsRecorded {
		return &out
	}
	if i.InputsRecorded && i.Coverage == Unknown {
		out.Coverage, out.Reasons = Complete, nil
	}
	out.Inputs, out.InputsRecorded = nil, false
	out.Seal()
	return &out
}

// InputEntry reduces a Git entry to what a stat observes: the mode, plus the
// object for symlinks and submodules.
func InputEntry(entry string, stat bool) string {
	parts := strings.Split(entry, ":")
	if !stat || len(parts) != 3 || parts[1] == "120000" || parts[1] == "160000" {
		return entry
	}
	return parts[0] + ":" + parts[1]
}

// ExternalReads keeps the paths outside root, dropping any path covered by a
// directory or file whose contents were read. When root's digest is a
// directory, stats of root's ancestors are dropped too, because that digest
// already pins each of them as a directory. A symlink digest holds only the
// link target, not which ancestor is the link, so those stats are kept.
func ExternalReads(root, digest string, read, stat []string) *Reads {
	root = CleanRoot(root)
	outside := func(name string) bool {
		return root != "" && name != root && !strings.HasPrefix(name, root+"/")
	}
	directory := strings.HasPrefix(digest, "git:040000:")
	ancestor := func(name string) bool {
		return directory && (name == "." || strings.HasPrefix(root, name+"/"))
	}
	out := &Reads{}
	kept := make(map[string]bool)
	covered := func(name string) bool {
		for dir := name; ; dir = path.Dir(dir) {
			if kept[dir] {
				return true
			}
			if dir == "." || dir == "/" {
				return false
			}
		}
	}
	read = slices.Clone(read)
	sort.Strings(read)
	for _, name := range read {
		if outside(name) && !covered(name) {
			kept[name] = true
			out.Read = append(out.Read, name)
		}
	}
	stat = slices.Clone(stat)
	sort.Strings(stat)
	for _, name := range slices.Compact(stat) {
		if outside(name) && !ancestor(name) && !covered(name) {
			out.Stat = append(out.Stat, name)
		}
	}
	return out
}

func ValidInputPath(name string) bool {
	if name == "" || path.IsAbs(name) || path.Clean(name) != name || name == ".." || strings.HasPrefix(name, "../") {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if strings.EqualFold(part, ".git") {
			return false
		}
	}
	return true
}

func (i *Identity) EligibleChecksum() string {
	if i == nil || i.Version != Version || i.Coverage != Complete || !i.Captured || i.ConfigChecksum == "" || i.Checksum == "" {
		return ""
	}
	if strings.HasPrefix(i.Digest, "git:") && !i.InputsRecorded {
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
	copy.Inputs = slices.Clone(i.Inputs)
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
	switch i.ResolvedFrom {
	case "", ResolvedFromCommit, ResolvedFromBranchRun, ResolvedFromLookup:
	default:
		return fmt.Errorf("invalid source resolution %q", i.ResolvedFrom)
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
	if len(i.Inputs) != 0 || i.InputsRecorded {
		if !strings.HasPrefix(i.Digest, "git:") || !i.InputsRecorded || len(i.Inputs) > MaxInputs {
			return fmt.Errorf("invalid source inputs")
		}
		if i.Coverage == Unknown && (len(i.Inputs) != 0 || !slices.Equal(i.Reasons, []string{InputsOverLimit})) {
			return fmt.Errorf("invalid source inputs")
		}
		for n, input := range i.Inputs {
			if !ValidInputPath(input.Path) || n > 0 && i.Inputs[n-1].Path >= input.Path || !validInputEntry(input) {
				return fmt.Errorf("invalid source input %q", input.Path)
			}
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

func validInputEntry(input Input) bool {
	if input.Entry == "" {
		return true
	}
	parts := strings.Split(input.Entry, ":")
	if len(parts) < 2 || len(parts) > 3 || parts[0] != "git" || len(parts[1]) != 6 || strings.Trim(parts[1], "01234567") != "" {
		return false
	}
	objectless := input.Stat && parts[1] != "120000" && parts[1] != "160000"
	if len(parts) == 2 {
		return objectless
	}
	return !objectless && IsCommit(parts[2])
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
	if actual.InputsRecorded {
		if expected.InputsRecorded {
			return fmt.Errorf("prepared source cannot record build inputs")
		}
		actual = actual.WithoutInputs()
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
	} else if actual.Coverage != Unknown && (actual.Kind != "helm" || expected.Commit != "" || actual.ChartVersion == "") {
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
