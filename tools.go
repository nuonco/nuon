//go:build tools
// +build tools

package tools

import (
	_ "github.com/a-h/templ/cmd/templ"
	_ "github.com/go-openapi/errors"
	_ "github.com/go-openapi/runtime"
	_ "github.com/go-openapi/strfmt"
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/swaggo/swag/cmd/swag"
	_ "go.opentelemetry.io/collector/cmd/builder"
)
