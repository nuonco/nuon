package customermanaged

import (
	"mime"
	"path/filepath"
	"sort"

	"github.com/opencontainers/go-digest"
)

type ReleaseArchive struct {
	SchemaVersion int               `json:"schema_version"`
	Members       map[string]string `json:"members"`
	Files         map[string]string `json:"files,omitempty"`
}

type ReleaseFile struct {
	Path      string `json:"path"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	MediaType string `json:"media_type"`
}

func (a ReleaseArchive) FileList() []ReleaseFile {
	files := make([]ReleaseFile, 0, len(a.Files))
	for path, contents := range a.Files {
		files = append(files, ReleaseFile{
			Path: path, Digest: digest.FromString(contents).String(), Size: int64(len(contents)), MediaType: releaseFileMediaType(path),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files
}

func releaseFileMediaType(path string) string {
	switch filepath.Ext(path) {
	case ".toml":
		return "application/toml"
	case ".rego":
		return "text/x-rego"
	case ".md":
		return "text/markdown"
	case ".sh":
		return "text/x-shellscript"
	}
	if mediaType := mime.TypeByExtension(filepath.Ext(path)); mediaType != "" {
		return mediaType
	}
	return "text/plain"
}
