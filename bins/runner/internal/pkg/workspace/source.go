package workspace

import "path/filepath"

type Source struct {
	// the user provided directory
	Path string

	// the root of the workspace
	Root string

	// whether this is a git source
	IsGit bool
}

func (s Source) AbsPath() string {
	return filepath.Join(s.Root, s.Path)
}

func (w *workspace) Source() *Source {
	return &Source{
		Path:  w.Src.Path,
		IsGit: w.isGit(),
		Root:  w.rootDir(),
	}
}
