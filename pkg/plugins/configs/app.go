package configs

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Apps[B, D any] struct {
	Project string `hcl:"project"`

	App App[B, D] `hcl:"app,block"`
}

func (a *Apps[B, D]) ToHCL() []byte {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(a, f.Body())
	return f.Bytes()
}

type App[B, D any] struct {
	Name string `hcl:"name,label"`

	Build  B `hcl:"build,block"`
	Deploy D `hcl:"deploy,block"`
}
