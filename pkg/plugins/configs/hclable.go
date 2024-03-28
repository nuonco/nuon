package configs

type Hclable interface {
	ToHCL() []byte
}
