package configs

type Registry[T any] struct {
	Use T `hcl:"use,block"`
}
