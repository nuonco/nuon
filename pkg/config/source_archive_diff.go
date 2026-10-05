package config

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

const (
	SourceFileOpUnchanged = "unchanged"
	SourceFileOpAdded     = "added"
	SourceFileOpRemoved   = "removed"
	SourceFileOpModified  = "modified"

	// MaxSourceDiffPatchBytes caps the inline unified patch per file. Larger
	// diffs are fetched per-file through the source-files endpoint.
	MaxSourceDiffPatchBytes = 32 << 10
)

type SourceFileDiff struct {
	Path           string `json:"path"`
	Op             string `json:"op"`
	BeforeSHA256   string `json:"before_sha256,omitempty"`
	AfterSHA256    string `json:"after_sha256,omitempty"`
	BeforeSize     int64  `json:"before_size,omitempty"`
	AfterSize      int64  `json:"after_size,omitempty"`
	Patch          string `json:"patch,omitempty"`
	PatchTruncated bool   `json:"patch_truncated,omitempty"`
}

// SourceArchiveDiff describes the delta between two source archives. Unchanged
// files carry only path + hash + size so the full file tree can be rendered
// from this payload alone; contents are fetched per-file on demand.
type SourceArchiveDiff struct {
	TotalFiles int              `json:"total_files"`
	Unchanged  int              `json:"unchanged"`
	Files      []SourceFileDiff `json:"files"`
}

func sourceFileSHA256(contents string) string {
	sum := sha256.Sum256([]byte(contents))
	return hex.EncodeToString(sum[:])
}

// Diff compares the receiver (head) against old (base). A nil old treats every
// file as added, which yields the full file listing for a single config.
// Output is deterministically sorted by path.
func (a *SourceArchive) Diff(old *SourceArchive) *SourceArchiveDiff {
	if old == nil {
		old = NewSourceArchive()
	}

	paths := make(map[string]struct{}, len(a.Files)+len(old.Files))
	for path := range a.Files {
		paths[path] = struct{}{}
	}
	for path := range old.Files {
		paths[path] = struct{}{}
	}
	sorted := make([]string, 0, len(paths))
	for path := range paths {
		sorted = append(sorted, path)
	}
	sort.Strings(sorted)

	out := &SourceArchiveDiff{TotalFiles: len(sorted)}
	for _, path := range sorted {
		after, hasAfter := a.Files[path]
		before, hasBefore := old.Files[path]
		switch {
		case hasAfter && !hasBefore:
			patch, truncated := sourceFilePatch("", after, path)
			out.Files = append(out.Files, SourceFileDiff{
				Path:           path,
				Op:             SourceFileOpAdded,
				AfterSHA256:    sourceFileSHA256(after),
				AfterSize:      int64(len(after)),
				Patch:          patch,
				PatchTruncated: truncated,
			})
		case hasBefore && !hasAfter:
			out.Files = append(out.Files, SourceFileDiff{
				Path:         path,
				Op:           SourceFileOpRemoved,
				BeforeSHA256: sourceFileSHA256(before),
				BeforeSize:   int64(len(before)),
			})
		case before != after:
			patch, truncated := sourceFilePatch(before, after, path)
			out.Files = append(out.Files, SourceFileDiff{
				Path:           path,
				Op:             SourceFileOpModified,
				BeforeSHA256:   sourceFileSHA256(before),
				AfterSHA256:    sourceFileSHA256(after),
				BeforeSize:     int64(len(before)),
				AfterSize:      int64(len(after)),
				Patch:          patch,
				PatchTruncated: truncated,
			})
		default:
			sha := sourceFileSHA256(after)
			out.Unchanged++
			out.Files = append(out.Files, SourceFileDiff{
				Path:         path,
				Op:           SourceFileOpUnchanged,
				BeforeSHA256: sha,
				AfterSHA256:  sha,
				BeforeSize:   int64(len(after)),
				AfterSize:    int64(len(after)),
			})
		}
	}
	return out
}

// sourceFilePatch renders a unified diff between before and after, truncated
// to MaxSourceDiffPatchBytes when larger.
func sourceFilePatch(before, after, path string) (string, bool) {
	if before == after {
		return "", false
	}
	patch, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(before),
		B:        difflib.SplitLines(after),
		FromFile: "a/" + path,
		ToFile:   "b/" + path,
		Context:  3,
	})
	if err != nil {
		return "", false
	}
	if len(patch) > MaxSourceDiffPatchBytes {
		patch = patch[:MaxSourceDiffPatchBytes]
		if i := strings.LastIndex(patch, "\n"); i >= 0 {
			patch = patch[:i+1]
		}
		return patch, true
	}
	return patch, false
}
