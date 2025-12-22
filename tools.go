//go:build tools
// +build tools

package tools

import (
        _ "github.com/dave/dst/decorator"
        _ "github.com/go-swagger/go-swagger/cmd/swagger"
        _ "github.com/go-toolsmith/astfmt"
        _ "github.com/golang/mock/mockgen"
        _ "github.com/grafana/codejen"
        _ "github.com/swaggo/swag/cmd/swag"
        _ "go.opentelemetry.io/collector/cmd/builder"
)
