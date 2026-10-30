package sourceidentity

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

func IsCommit(ref string) bool { return plumbing.IsHash(ref) && ref == strings.ToLower(ref) }

func Checkout(ctx context.Context, repo *git.Repository, ref string, auth transport.AuthMethod) error {
	if !IsCommit(ref) {
		return fmt.Errorf("source requires a full immutable commit")
	}
	hash := plumbing.NewHash(ref)
	if _, err := repo.CommitObject(hash); err != nil {
		fetchErr := repo.FetchContext(ctx, &git.FetchOptions{Auth: auth, RefSpecs: []config.RefSpec{config.RefSpec("+" + ref + ":refs/nuon/captured")}})
		if fetchErr != nil && fetchErr != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("fetch captured commit: %w", fetchErr)
		}
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if err := wt.Checkout(&git.CheckoutOptions{Hash: hash, Force: true}); err != nil {
		return fmt.Errorf("checkout captured commit: %w", err)
	}
	head, err := repo.Head()
	if err != nil {
		return err
	}
	if head.Hash() != hash {
		return fmt.Errorf("checkout does not match captured commit")
	}
	return nil
}

func gitDigest(source *sourcefs.FS, root string) (string, error) {
	root, err := source.Relative(filepath.FromSlash(root))
	if err != nil {
		return "", err
	}
	name := "."
	for _, part := range strings.Split(root, string(filepath.Separator)) {
		name = filepath.Join(name, part)
		info, err := source.Lstat(name)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			root = name
			break
		}
	}
	entry, err := gitEntry(source, root)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("git:%06o:%s", entry.Mode, entry.Hash), nil
}

func gitEntry(source *sourcefs.FS, name string) (object.TreeEntry, error) {
	entry := object.TreeEntry{Name: filepath.Base(name)}
	info, err := source.Lstat(name)
	if err != nil {
		return entry, err
	}
	entry.Mode, err = filemode.NewFromOSFileMode(info.Mode())
	if err != nil {
		return entry, err
	}
	switch entry.Mode {
	case filemode.Symlink:
		target, err := source.Readlink(name)
		if err != nil {
			return entry, err
		}
		entry.Hash = plumbing.ComputeHash(plumbing.BlobObject, []byte(target))
	case filemode.Dir:
		children, err := source.ReadDir(name)
		if err != nil {
			return entry, err
		}
		tree := &object.Tree{}
		for _, child := range children {
			if child.Name() == ".git" {
				continue
			}
			next, err := gitEntry(source, filepath.Join(name, child.Name()))
			if err != nil {
				return entry, err
			}
			tree.Entries = append(tree.Entries, next)
		}
		sort.Sort(object.TreeEntrySorter(tree.Entries))
		encoded := &plumbing.MemoryObject{}
		if err := tree.Encode(encoded); err != nil {
			return entry, err
		}
		entry.Hash = encoded.Hash()
	default:
		file, err := source.Open(name)
		if err != nil {
			return entry, err
		}
		hasher := plumbing.NewHasher(plumbing.BlobObject, info.Size())
		_, err = io.Copy(hasher, file)
		if err := errors.Join(err, file.Close()); err != nil {
			return entry, err
		}
		entry.Hash = hasher.Sum()
	}
	return entry, nil
}
