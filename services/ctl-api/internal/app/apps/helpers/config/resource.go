package config

type resource interface {
	ToResourceType() string
	ToResource() (map[string]interface{}, error)
}
