package sourceidentity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

func Capture(source *sourcefs.FS, selected *Identity) (*Identity, error) {
	if selected == nil {
		return nil, fmt.Errorf("source identity is required")
	}
	out := *selected
	out.Files = nil
	out.Reasons = append([]string(nil), selected.Reasons...)
	root := "."
	path, rootLinks, err := source.ResolveWithLinks(filepath.FromSlash(selected.Root))
	if err != nil {
		return nil, err
	}
	addFile := func(file File) error {
		out.Files = append(out.Files, file)
		if len(out.Files) > 20000 {
			return fmt.Errorf("source manifest exceeds file limit")
		}
		return nil
	}
	directories := make(map[string]bool)
	var walk func(string, string, []sourcefs.Link) error
	walk = func(name, rel string, links []sourcefs.Link) error {
		info, err := source.Lstat(name)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			name, links, err = source.ResolveWithLinks(name)
			if err != nil {
				return err
			}
			info, err = source.Lstat(name)
			if err != nil {
				return err
			}
		}
		sum := ""
		if len(links) > 0 {
			target, err := filepath.Rel(root, name)
			if err != nil {
				return err
			}
			sum = Hash(struct {
				Links  []sourcefs.Link
				Target string
			}{links, filepath.ToSlash(target)})
		}
		if info.IsDir() {
			if directories[name] {
				return fmt.Errorf("source directory symlink loop at %s", name)
			}
			directories[name] = true
			defer delete(directories, name)
			if len(links) > 0 {
				out.MarkUnknown("source_directory_symlink_closure_uninspected")
				if err := addFile(File{Path: filepath.ToSlash(rel), SHA256: sum}); err != nil {
					return err
				}
			}
			entries, err := source.ReadDir(name)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if entry.Name() == ".git" {
					continue
				}
				if err := walk(filepath.Join(name, entry.Name()), filepath.Join(rel, entry.Name()), nil); err != nil {
					return err
				}
			}
			return nil
		}
		if rel == "." && out.Kind != "helm" {
			return fmt.Errorf("source root is not a directory")
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source file %s is not regular", name)
		}
		data, err := source.ReadFile(name)
		if err != nil {
			return err
		}
		if len(links) > 0 {
			// Equal materialized bytes must not hide a changed link chain or target.
			sum = Hash(struct{ Link, Content string }{sum, sha256Bytes(data)})
		} else {
			sum = sha256Bytes(data)
		}
		if err := addFile(File{Path: filepath.ToSlash(rel), SHA256: sum, Executable: info.Mode().Perm()&0111 != 0}); err != nil {
			return err
		}
		return nil
	}
	if err := walk(path, ".", rootLinks); err != nil {
		return nil, err
	}
	out.Captured = true
	out.Seal()
	return &out, nil
}

func Verify(source *sourcefs.FS, expected *Identity) error {
	if expected == nil {
		return nil
	}
	if err := expected.Validate(); err != nil {
		return err
	}
	if !expected.Captured {
		return nil
	}
	if strings.HasPrefix(expected.Digest, "git:") {
		actual, err := gitDigest(source, expected.Root)
		if err != nil {
			return err
		}
		if actual != expected.Digest {
			return fmt.Errorf("Git source differs from prepared directory tree")
		}
		return nil
	}
	actual, err := Capture(source, expected)
	if err != nil {
		return err
	}
	if !Matches(expected, actual) {
		return fmt.Errorf("captured source differs from prepared build inputs")
	}
	return nil
}
