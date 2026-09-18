package vcserrors

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const SourceDirectoryNotFoundType compositeerrors.Type = "vcs.source_directory_not_found"

type SourceDirectoryNotFoundError struct {
	Directory string `json:"directory,omitempty"`
}

var (
	_ compositeerrors.CompositeError = (*SourceDirectoryNotFoundError)(nil)
	_ compositeerrors.HintsProvider  = (*SourceDirectoryNotFoundError)(nil)
)

func (e *SourceDirectoryNotFoundError) Error() string {
	return SourceDirectoryNotFoundMessage(e.Directory)
}

func (e *SourceDirectoryNotFoundError) Type() compositeerrors.Type {
	return SourceDirectoryNotFoundType
}

func (e *SourceDirectoryNotFoundError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityFatal
}

func (e *SourceDirectoryNotFoundError) Sections() []compositeerrors.Section {
	return []compositeerrors.Section{
		compositeerrors.MarkdownSection("Why", "The directory configured as the source for this build does not exist at the checked out git ref. It may have been renamed or moved, or the ref may not contain it yet."),
		compositeerrors.MarkdownSection("How to fix", "Point the config's directory at a path that exists at the configured ref, or push the directory to that ref, then run again."),
	}
}

func (e *SourceDirectoryNotFoundError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithTerminal()
}

func SourceDirectoryNotFoundMessage(directory string) string {
	if directory == "" {
		return "The configured source directory was not found in the repository"
	}
	return fmt.Sprintf("Source directory %q was not found in the repository", directory)
}
