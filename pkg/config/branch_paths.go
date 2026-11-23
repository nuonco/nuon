package config

import (
	"fmt"
	"path"
	"strings"
	"unicode"

	"github.com/gobwas/glob"
)

const (
	maxRunPathLength = 512

	maxDoubleStarSegments = 3
)

type runPathMatcher []glob.Glob

func (m runPathMatcher) match(p string) bool {
	for _, g := range m {
		if g.Match(p) {
			return true
		}
	}
	return false
}

// RunPathFilter is either an allow list (paths) or a deny list (paths_ignore),
// never both.
type RunPathFilter struct {
	include []runPathMatcher
	ignore  []runPathMatcher
}

type RunPathDecision struct {
	Run    bool
	Reason string
}

func ValidateRunPaths(paths, pathsIgnore []string) error {
	_, err := NewRunPathFilter(paths, pathsIgnore)
	return err
}

// NewRunPathFilter returns nil when no filter applies: with neither paths nor
// paths_ignore set, every change runs.
func NewRunPathFilter(paths, pathsIgnore []string) (*RunPathFilter, error) {
	if len(paths) > 0 && len(pathsIgnore) > 0 {
		return nil, fmt.Errorf("paths and paths_ignore are mutually exclusive")
	}
	if len(paths) == 0 && len(pathsIgnore) == 0 {
		return nil, nil
	}

	f := &RunPathFilter{}
	for _, p := range paths {
		if strings.HasPrefix(p, "!") {
			return nil, fmt.Errorf("paths: pattern %q cannot be negated; use paths_ignore to deny paths", p)
		}
		m, err := compileRunPath(p)
		if err != nil {
			return nil, fmt.Errorf("paths: %w", err)
		}
		f.include = append(f.include, m)
	}
	for _, p := range pathsIgnore {
		if strings.HasPrefix(p, "!") {
			return nil, fmt.Errorf("paths_ignore: pattern %q cannot be negated", p)
		}
		m, err := compileRunPath(p)
		if err != nil {
			return nil, fmt.Errorf("paths_ignore: %w", err)
		}
		f.ignore = append(f.ignore, m)
	}
	return f, nil
}

func compileRunPath(p string) (runPathMatcher, error) {
	if len(p) > maxRunPathLength {
		return nil, fmt.Errorf("pattern exceeds %d bytes", maxRunPathLength)
	}
	p, _ = strings.CutPrefix(p, "./")
	if strings.TrimSpace(p) == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}
	if path.IsAbs(p) || strings.Contains(p, `\`) || strings.Contains(p, "://") || hasDriveLetter(p) {
		return nil, fmt.Errorf("pattern %q must be relative to the repository root", p)
	}
	if strings.ContainsFunc(p, unicode.IsControl) {
		return nil, fmt.Errorf("pattern %q contains control characters", p)
	}
	segs := strings.Split(p, "/")
	optional := 0
	for i, seg := range segs {
		if seg == ".." {
			return nil, fmt.Errorf("pattern %q cannot contain '..'", p)
		}
		if seg == "**" && i < len(segs)-1 {
			optional++
		}
	}
	if optional > maxDoubleStarSegments {
		return nil, fmt.Errorf("pattern %q has more than %d '**/' segments", p, maxDoubleStarSegments)
	}

	var m runPathMatcher
	for _, variant := range expandDoubleStar(segs) {
		g, err := glob.Compile(variant, '/')
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", p, err)
		}
		m = append(m, g)
	}
	return m, nil
}

// expandDoubleStar lets each non-trailing "**" segment also match zero
// directories, which gobwas/glob does not do on its own. path.Match cannot
// express "**".
func expandDoubleStar(segs []string) []string {
	variants := [][]string{{}}
	for i, seg := range segs {
		next := make([][]string, 0, len(variants)*2)
		for _, v := range variants {
			next = append(next, append(append([]string{}, v...), seg))
			if seg == "**" && i < len(segs)-1 {
				next = append(next, append([]string{}, v...))
			}
		}
		variants = next
	}
	out := make([]string, 0, len(variants))
	for _, v := range variants {
		out = append(out, strings.Join(v, "/"))
	}
	return out
}

// Paths that cannot be evaluated safely make Match fail open.
func (f *RunPathFilter) Match(files []string) RunPathDecision {
	if f == nil {
		return RunPathDecision{Run: true}
	}
	if len(files) == 0 {
		return RunPathDecision{Reason: "no changed files"}
	}

	cleaned := make([]string, 0, len(files))
	for _, file := range files {
		p, ok := normalizeRunPath(file)
		if !ok || p == "" {
			return RunPathDecision{Run: true}
		}
		cleaned = append(cleaned, p)
	}

	if len(f.include) > 0 {
		for _, p := range cleaned {
			if f.included(p) {
				return RunPathDecision{Run: true}
			}
		}
		return RunPathDecision{Reason: fmt.Sprintf("no changed files match run.paths (%d files checked)", len(cleaned))}
	}

	for _, p := range cleaned {
		if !f.ignored(p) {
			return RunPathDecision{Run: true}
		}
	}
	return RunPathDecision{Reason: fmt.Sprintf("all %d changed files match run.paths_ignore", len(cleaned))}
}

func (f *RunPathFilter) included(p string) bool {
	for _, m := range f.include {
		if m.match(p) {
			return true
		}
	}
	return false
}

func (f *RunPathFilter) ignored(p string) bool {
	for _, m := range f.ignore {
		if m.match(p) {
			return true
		}
	}
	return false
}

func normalizeRunPath(p string) (string, bool) {
	p = strings.TrimSpace(p)
	if strings.ContainsFunc(p, unicode.IsControl) || strings.Contains(p, `\`) || path.IsAbs(p) {
		return "", false
	}
	p, _ = strings.CutPrefix(p, "./")
	p = path.Clean(p)
	if p == "." {
		return "", true
	}
	if p == ".." || strings.HasPrefix(p, "../") {
		return "", false
	}
	return p, true
}

func hasDriveLetter(p string) bool {
	return len(p) >= 2 && p[1] == ':' && ((p[0] >= 'a' && p[0] <= 'z') || (p[0] >= 'A' && p[0] <= 'Z'))
}
