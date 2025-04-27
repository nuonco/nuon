package variables

import "context"

type VarFile struct {
	Filename string
	Contents []byte
}

// Variables configures a way to set up a terraform workspace with appropriate variables, that can come from either
// settings in a file, or the environment.
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=variables_mock.go -source=variables.go -package=variables
type Variables interface {
	Init(context.Context) error

	// Env vars represent environment variables that need to be set in the environment before any runs
	GetEnv(context.Context) (map[string]string, error)

	// File vars represent files that should be written into a variables file in terraform
	GetFiles(context.Context) ([]VarFile, error)
}
