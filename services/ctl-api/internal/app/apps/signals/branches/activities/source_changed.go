package activities

import (
	"path/filepath"
	"strings"
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

// enrichConfigDiffWithSourceChanged copies FullDiff into a ConfigDiff blob shape
// and sets source_changed on component entries whose tracked repo+directory
// intersect changedPaths. A top-level map covers every component, including
// those with source-only changes that do not appear in the config diff.
func enrichConfigDiffWithSourceChanged(
	full *ComputeAppConfigDiffOutput,
	sources []componentSource,
	branchRepo string,
	changedPaths []string,
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
			section.Entries = append(section.Entries, entry)
		}
		out.Sections = append(out.Sections, section)
	}

	return out
}
