package outputs

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/log"
	"github.com/powertoolsdev/mono/pkg/terraform/outputs"
	terraformworkspace "github.com/powertoolsdev/mono/pkg/terraform/workspace"
)

// TerraformOutputs is a helper method that returns outputs in a compatible way with the way they are currently stored
// in s3
func TerraformOutputs(ctx context.Context, workspace terraformworkspace.Workspace) (map[string]interface{}, error) {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, err
	}

	hlog := log.NewHClog(l)
	outs, err := workspace.Output(ctx, hlog)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get workspace outputs")
	}

	outsPB, err := outputs.TFOutputMetaToStructPB(outs)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to standard outputs: %w", err)
	}

	return outsPB.AsMap(), nil
}
