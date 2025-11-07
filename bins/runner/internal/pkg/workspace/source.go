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
	var path string
	if w.Src != nil {
		path = w.Src.Path
	}
	
	return &Source{
		Path:  path,
		IsGit: w.isGit(),
		Root:  w.rootDir(),
	}
}
