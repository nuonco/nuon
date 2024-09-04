package workspace

type Source struct {
	// the user provided directory
	Path string

	// the root of the workspace
	Root string

	// whether this is a git source
	IsGit bool
}

func (w *workspace) Source() *Source {
	return &Source{
		Path:  w.Src.Path,
		IsGit: w.isGit(),
		Root:  w.rootDir(),
	}
}
