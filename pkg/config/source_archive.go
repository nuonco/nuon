package config

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
)

type SourceArchive struct {
	SchemaVersion int               `json:"schema_version"`
	Files         map[string]string `json:"files"`
	Members       map[string]string `json:"members"`
}

const maxSourceArchiveBytes = 250 << 20

func NewSourceArchive() *SourceArchive {
	return &SourceArchive{SchemaVersion: 3, Files: map[string]string{}, Members: map[string]string{}}
}

func (a *SourceArchive) AddFile(relative string, contents []byte) error {
	relative = filepath.ToSlash(filepath.Clean(relative))
	if relative == "" || strings.HasPrefix(relative, "installs/") || hasHiddenPathSegment(relative) {
		return nil
	}
	if relative == ".." || strings.HasPrefix(relative, "../") || filepath.IsAbs(relative) {
		return fmt.Errorf("source file %s is outside the config directory", relative)
	}
	if _, exists := a.Files[relative]; exists {
		return nil
	}
	if !utf8.Valid(contents) {
		return fmt.Errorf("source file %s is not UTF-8 text", relative)
	}
	total := len(contents)
	for _, existing := range a.Files {
		total += len(existing)
	}
	if total > maxSourceArchiveBytes {
		return fmt.Errorf("source archive exceeds %d bytes", maxSourceArchiveBytes)
	}
	a.Files[relative] = string(contents)
	return nil
}

func (a *SourceArchive) AddMember(kind, logicalName, path string) error {
	if kind == "" || logicalName == "" {
		return nil
	}
	path = filepath.ToSlash(path)
	key := kind + ":" + logicalName
	if existing := a.Members[key]; existing != "" && existing != path {
		return fmt.Errorf("source files %s and %s both define %s", existing, path, key)
	}
	a.Members[key] = path
	return nil
}

func hasHiddenPathSegment(path string) bool {
	for _, segment := range strings.Split(path, "/") {
		if strings.HasPrefix(segment, ".") {
			return true
		}
	}
	return false
}

func (a SourceArchive) MemberSource(key string) (string, bool) {
	path := a.Members[key]
	contents, ok := a.Files[path]
	return contents, ok && path != ""
}

func (a *SourceArchive) ReindexMembers() error {
	if a.SchemaVersion >= 3 {
		for key, path := range a.Members {
			kind, name, found := strings.Cut(key, ":")
			if !found || kind == "" || name == "" || path == "" {
				return fmt.Errorf("source archive contains invalid member identity %q", key)
			}
			if _, ok := a.Files[path]; !ok {
				return fmt.Errorf("source member %s references missing file %s", key, path)
			}
		}
		return nil
	}

	paths := make([]string, 0, len(a.Files))
	for path := range a.Files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	members := make(map[string]string)
	for _, path := range paths {
		memberKey, err := sourceMemberKey(path, []byte(a.Files[path]))
		if err != nil {
			return fmt.Errorf("index source file %s: %w", path, err)
		}
		if memberKey == "" {
			continue
		}
		if existing := members[memberKey]; existing != "" {
			return fmt.Errorf("source files %s and %s both define %s", existing, path, memberKey)
		}
		members[memberKey] = path
	}
	a.Members = members
	return nil
}

func sourceMemberKey(path string, contents []byte) (string, error) {
	if filepath.Ext(path) != ".toml" {
		return "", nil
	}
	switch path {
	case "metadata.toml":
		return "metadata:metadata", nil
	case "inputs.toml":
		return "input:inputs", nil
	case "policies.toml":
		return "policy:policies", nil
	case "break_glass.toml":
		return "break_glass:break-glass", nil
	case "sandbox.toml":
		return "sandbox:sandbox", nil
	case "stack.toml":
		return "stack:stack", nil
	case "runner.toml":
		return "runner:runner", nil
	}
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", nil
	}
	kind := strings.TrimSuffix(parts[0], "s")
	if kind != "component" && kind != "action" && kind != "runbook" && kind != "permission" {
		return "", nil
	}
	var header struct {
		Name string `toml:"name"`
	}
	if err := toml.Unmarshal(contents, &header); err != nil {
		return "", err
	}
	if header.Name == "" {
		return "", fmt.Errorf("source file %s has no name field", path)
	}
	return kind + ":" + header.Name, nil
}
