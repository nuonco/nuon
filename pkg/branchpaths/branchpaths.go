package branchpaths

import (
	"fmt"
	"path"
	"strings"
	"unicode"

	"github.com/gobwas/glob"
)

const (
	MaxPatterns      = 100
	MaxPatternLength = 512

	maxDoubleStarSegments = 3
)

type matcher []glob.Glob

func (m matcher) Match(p string) bool {
	for _, g := range m {
		if g.Match(p) {
			return true
		}
	}
	return false
}

type rule struct {
	matcher matcher
	negate  bool
}

type Filter struct {
	include []rule
	ignore  []matcher
}

type Decision struct {
	Run    bool
	Reason string
}

func Validate(paths, pathsIgnore []string) error {
	_, err := compile(paths, pathsIgnore)
	return err
}

// New returns nil when no filter applies. Without paths or paths_ignore, a
// non-root directory becomes the implicit filter "<directory>/**".
func New(paths, pathsIgnore []string, directory string) (*Filter, error) {
	if len(paths) == 0 && len(pathsIgnore) == 0 {
		dir, ok := normalizePath(directory)
		if !ok || dir == "" {
			return nil, nil
		}
		g, err := glob.Compile(glob.QuoteMeta(dir)+"/**", '/')
		if err != nil {
			return nil, fmt.Errorf("invalid directory %q: %w", directory, err)
		}
		return &Filter{include: []rule{{matcher: matcher{g}}}}, nil
	}
	return compile(paths, pathsIgnore)
}

func compile(paths, pathsIgnore []string) (*Filter, error) {
	if len(paths) > 0 && len(pathsIgnore) > 0 {
		return nil, fmt.Errorf("paths and paths_ignore are mutually exclusive")
	}
	if len(paths) == 0 && len(pathsIgnore) == 0 {
		return nil, nil
	}
	if len(paths) > MaxPatterns || len(pathsIgnore) > MaxPatterns {
		return nil, fmt.Errorf("at most %d path patterns are allowed", MaxPatterns)
	}

	f := &Filter{}
	for _, p := range paths {
		negate := strings.HasPrefix(p, "!")
		m, err := compilePattern(strings.TrimPrefix(p, "!"))
		if err != nil {
			return nil, fmt.Errorf("paths: %w", err)
		}
		f.include = append(f.include, rule{matcher: m, negate: negate})
	}
	for _, p := range pathsIgnore {
		if strings.HasPrefix(p, "!") {
			return nil, fmt.Errorf("paths_ignore: pattern %q cannot be negated", p)
		}
		m, err := compilePattern(p)
		if err != nil {
			return nil, fmt.Errorf("paths_ignore: %w", err)
		}
		f.ignore = append(f.ignore, m)
	}
	return f, nil
}

func compilePattern(p string) (matcher, error) {
	if len(p) > MaxPatternLength {
		return nil, fmt.Errorf("pattern exceeds %d bytes", MaxPatternLength)
	}
	p = strings.TrimPrefix(p, "./")
	if strings.TrimSpace(p) == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, `\`) || strings.Contains(p, "://") || hasDriveLetter(p) {
		return nil, fmt.Errorf("pattern %q must be relative to the repository root", p)
	}
	if hasControlChar(p) {
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

	var m matcher
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
// directories, which gobwas/glob does not do on its own.
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
func (f *Filter) Match(files []string) Decision {
	if f == nil {
		return Decision{Run: true}
	}
	if len(files) == 0 {
		return Decision{Reason: "no changed files"}
	}

	cleaned := make([]string, 0, len(files))
	for _, file := range files {
		p, ok := normalizePath(file)
		if !ok || p == "" {
			return Decision{Run: true}
		}
		cleaned = append(cleaned, p)
	}

	if len(f.include) > 0 {
		for _, p := range cleaned {
			if f.included(p) {
				return Decision{Run: true}
			}
		}
		return Decision{Reason: fmt.Sprintf("no changed files match run.paths (%d files checked)", len(cleaned))}
	}

	for _, p := range cleaned {
		if !f.ignored(p) {
			return Decision{Run: true}
		}
	}
	return Decision{Reason: fmt.Sprintf("all %d changed files match run.paths_ignore", len(cleaned))}
}

func (f *Filter) included(p string) bool {
	matched := false
	for _, r := range f.include {
		if r.matcher.Match(p) {
			matched = !r.negate
		}
	}
	return matched
}

func (f *Filter) ignored(p string) bool {
	for _, m := range f.ignore {
		if m.Match(p) {
			return true
		}
	}
	return false
}

func normalizePath(p string) (string, bool) {
	p = strings.TrimSpace(p)
	if hasControlChar(p) || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") {
		return "", false
	}
	p = path.Clean(strings.TrimPrefix(p, "./"))
	if p == "." {
		return "", true
	}
	if p == ".." || strings.HasPrefix(p, "../") {
		return "", false
	}
	return p, true
}

func hasControlChar(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func hasDriveLetter(p string) bool {
	return len(p) >= 2 && p[1] == ':' && ((p[0] >= 'a' && p[0] <= 'z') || (p[0] >= 'A' && p[0] <= 'Z'))
}
