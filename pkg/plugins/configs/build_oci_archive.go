package configs

type OCIArchiveBuild struct {
	Plugin string `hcl:"plugin,label"`

	Labels      map[string]string `hcl:"labels,optional"`
	ArchiveType string            `hcl:"archive_type"`
}
