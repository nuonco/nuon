package activities

import (
	"path/filepath"
	"strings"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
)

type componentSource struct {
	Name      string
	Repo      string
	Directory string
}

// normalizeRepoPath cleans a repo-relative path for prefix matching.
func normalizeRepoPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "./")
	p = filepath.ToSlash(filepath.Clean(p))
	if p == "." {
		return ""
	}
	return p
}

func normalizeRepoURL(repo string) string {
	repo = strings.TrimSuffix(repo, ".git")
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimPrefix(repo, "http://github.com/")
	repo = strings.TrimPrefix(repo, "git@github.com:")
	repo = strings.TrimSuffix(repo, "/")
	return strings.ToLower(repo)
}

func repoURLsEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return normalizeRepoURL(a) == normalizeRepoURL(b)
}

// pathMatchesDirectory reports whether path is under directory (or equal to it).
// Empty / "." directories match any path in the same repo.
func pathMatchesDirectory(path, directory string) bool {
	path = normalizeRepoPath(path)
	directory = normalizeRepoPath(directory)
	if path == "" {
		return false
	}
	if directory == "" {
		return true
	}
	return path == directory || strings.HasPrefix(path, directory+"/")
}

// anyPathMatchesDirectory returns true if any changed path falls under directory.
// An empty directory (repo root) matches any changed path.
func anyPathMatchesDirectory(changedPaths []string, directory string) bool {
	if normalizeRepoPath(directory) == "" {
		return len(changedPaths) > 0
	}
	for _, p := range changedPaths {
		if pathMatchesDirectory(p, directory) {
			return true
		}
	}
	return false
}

func componentSourceChanged(src componentSource, branchRepo string, changedPaths []string) bool {
	if !repoURLsEqual(src.Repo, branchRepo) {
		return false
	}
	return anyPathMatchesDirectory(changedPaths, src.Directory)
}

// sectionMemberKind maps a config diff section name to the source archive
// member kind that holds each entity's defining file.
var sectionMemberKind = map[string]string{
	"Components":      "component",
	"Actions":         "action",
	"Runbooks":        "runbook",
	"Policies":        "policy",
	"Permissions":     "permission",
	"Sandbox":         "sandbox",
	"Runner":          "runner",
	"Break glass":     "break_glass",
	"Operation roles": "operation_roles",
}

// enrichConfigDiffWithSourceChanged copies FullDiff into a ConfigDiff blob shape
// and sets source_changed on component entries whose tracked repo+directory
// intersect changedPaths. A top-level map covers every component, including
// those with source-only changes that do not appear in the config diff.
// Entry File paths resolve through the head config's source archive members
// so the UI can link an entity to the file it was parsed from.
func enrichConfigDiffWithSourceChanged(
	full *ComputeAppConfigDiffOutput,
	sources []componentSource,
	branchRepo string,
	changedPaths []string,
	headArchive *pkgconfig.SourceArchive,
	baseArchive *pkgconfig.SourceArchive,
) *ConfigDiffWithSourceOutput {
	if full == nil {
		full = &ComputeAppConfigDiffOutput{}
	}

	componentSourceChangedMap := make(map[string]bool, len(sources))
	for _, src := range sources {
		if src.Name == "" {
			continue
		}
		componentSourceChangedMap[src.Name] = componentSourceChanged(src, branchRepo, changedPaths)
	}

	out := &ConfigDiffWithSourceOutput{
		ConfigFile:             full.ConfigFile,
		Additions:              full.Additions,
		Removals:               full.Removals,
		Changed:                full.Changed,
		ComponentSourceChanged: componentSourceChangedMap,
		Sections:               make([]ConfigDiffSectionWithSource, 0, len(full.Sections)),
	}

	for _, sec := range full.Sections {
		section := ConfigDiffSectionWithSource{
			Name:      sec.Name,
			Additions: sec.Additions,
			Removals:  sec.Removals,
			Changed:   sec.Changed,
			Entries:   make([]ConfigDiffEntryWithSource, 0, len(sec.Entries)),
		}
		for _, e := range sec.Entries {
			entry := ConfigDiffEntryWithSource{
				Op:          e.Op,
				Name:        e.Name,
				Description: e.Description,
			}
			if sec.Name == "Components" {
				entry.SourceChanged = componentSourceChangedMap[e.Name]
			}
			entry.File = memberFilePath(headArchive, sec.Name, e.Name)
			if entry.File == "" {
				entry.File = memberFilePath(baseArchive, sec.Name, e.Name)
			}
			section.Entries = append(section.Entries, entry)
		}
		out.Sections = append(out.Sections, section)
	}

	return out
}

// sectionMemberFallbackKeys returns extra member keys to try for a section
// after the primary kind:name lookup fails.
var sectionMemberFallbackKeys = map[string][]string{
	// All policies declared in a single policies.toml share one member.
	"Policies": {"policy:policies"},
	// Archives captured before the member naming was normalized use a
	// hyphenated name for break_glass.toml.
	"Break glass": {"break_glass:break-glass"},
}

// permissionMemberKeys returns archive member keys to try for a Permissions
// entry, most specific first. Standard roles are indexed by type
// (provision_role → provision), custom roles and named policies by their bare
// name after the diff key prefix is stripped.
func permissionMemberKeys(name string) []string {
	if n := strings.TrimPrefix(name, "named_policy."); n != name {
		return []string{"permission_policy:" + n}
	}
	if n := strings.TrimPrefix(name, "custom_role."); n != name {
		return []string{"permission:" + n, "permission:" + name}
	}
	if trimmed := strings.TrimSuffix(name, "_role"); trimmed != name {
		return []string{"permission:" + trimmed, "permission:" + name}
	}
	return []string{"permission:" + name}
}

// memberFilePath resolves an entity's defining config file from the source
// archive member index. Policy entry names carry a "policy." prefix that the
// member keys do not.
func memberFilePath(archive *pkgconfig.SourceArchive, section, name string) string {
	if archive == nil {
		return ""
	}
	if section == "Permissions" {
		for _, key := range permissionMemberKeys(name) {
			if path := archive.Members[key]; path != "" {
				return path
			}
		}
	}
	if kind := sectionMemberKind[section]; kind != "" {
		if path := archive.Members[kind+":"+strings.TrimPrefix(name, kind+".")]; path != "" {
			return path
		}
	}
	for _, key := range sectionMemberFallbackKeys[section] {
		if path := archive.Members[key]; path != "" {
			return path
		}
	}
	return ""
}
