package configs

type Deploy[T any] struct {
	Use T `hcl:"use,block"`
}
