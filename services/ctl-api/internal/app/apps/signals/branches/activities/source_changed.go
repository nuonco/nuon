package activities

import (
	"path/filepath"
	"strings"
)

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

// pathMatchesDirectory reports whether path is under directory (or equal to it).
// Empty / "." directories never match — they are not a usable source root for
// rebuild decisions (inputs-only changes would otherwise mark every component).
func pathMatchesDirectory(path, directory string) bool {
	path = normalizeRepoPath(path)
	directory = normalizeRepoPath(directory)
	if directory == "" || path == "" {
		return false
	}
	return path == directory || strings.HasPrefix(path, directory+"/")
}

// anyPathMatchesDirectory returns true if any changed path falls under directory.
func anyPathMatchesDirectory(changedPaths []string, directory string) bool {
	for _, p := range changedPaths {
		if pathMatchesDirectory(p, directory) {
			return true
		}
	}
	return false
}

// enrichConfigDiffWithSourceChanged copies FullDiff into a ConfigDiff blob shape
// and sets source_changed on component entries whose Directory intersects changedPaths.
// Non-component sections get source_changed=false.
func enrichConfigDiffWithSourceChanged(
	full *ComputeAppConfigDiffOutput,
	componentDirs map[string]string,
	changedPaths []string,
) *ConfigDiffWithSourceOutput {
	if full == nil {
		return &ConfigDiffWithSourceOutput{}
	}

	out := &ConfigDiffWithSourceOutput{
		ConfigFile: full.ConfigFile,
		Additions:  full.Additions,
		Removals:   full.Removals,
		Changed:    full.Changed,
		Sections:   make([]ConfigDiffSectionWithSource, 0, len(full.Sections)),
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
				dir, ok := componentDirs[e.Name]
				if ok {
					entry.SourceChanged = anyPathMatchesDirectory(changedPaths, dir)
				}
			}
			section.Entries = append(section.Entries, entry)
		}
		out.Sections = append(out.Sections, section)
	}

	return out
}
