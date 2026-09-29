package variables

import "context"

type VarFile struct {
	Filename string
	Contents []byte
}

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=variables_mock.go -source=variables.go -package=variables
type Variables interface {
	Init(context.Context) error

	GetEnv(context.Context) (map[string]string, error)

	GetFiles(context.Context) ([]VarFile, error)
}
