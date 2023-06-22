package configs

type Build[T, R any] struct {
	Use      T `hcl:"use,block"`
	Registry R `hcl:"registry,block"`
}

type NoRegistryBuild[T any] struct {
	Use T `hcl:"use,block"`
}
