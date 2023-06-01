package exec

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=fns_mock_test.go -source=fns_test.go -package=exec
type ui interface {
	terminal.UI
}

type hcLog interface {
	hclog.Logger
}

type testExecFns interface {
	Init(context.Context) error

	// terraform functions
	TerraformOutput(context.Context) (map[string]tfexec.OutputMeta, error)
	TerraformState(context.Context) (*tfjson.State, error)
	TerraformPlan(context.Context) (*tfjson.Plan, error)
}
