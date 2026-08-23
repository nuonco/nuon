module github.com/nuonco/nuon/sdks/stack

go 1.25.0

require github.com/stretchr/testify v1.11.1

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/nuonco/nuon/sdks/auth v0.0.0-00010101000000-000000000000
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Local development only: Go ignores replace directives in dependency modules, so
// anything consuming this SDK resolves sdks/auth from its require above. That version
// must be a real published one before this module is consumed outside the repo.
replace github.com/nuonco/nuon/sdks/auth => ../auth
