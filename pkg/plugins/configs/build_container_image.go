package configs

type ContainerImageBuild struct {
	Plugin string `hcl:"plugin,label"`

	Tag string `hcl:"tag"`

	Source OCIRegistryRepository `hcl:"oci_registry_repository,block"`
}
