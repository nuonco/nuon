package configs

// OCIArchiveBuild is used by the terraform plugin to create an OCI archive with the build parameters
type OCIArchiveBuild struct {
	Plugin string `hcl:"plugin,label"`

	Labels      map[string]string `hcl:"labels,optional"`
	ArchiveType string            `hcl:"archive_type"`
}
