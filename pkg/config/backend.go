package config

type BackendType string

const (
	BackendTypeS3    BackendType = "s3"
	BackendTypeLocal BackendType = "local"
)
