package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func writeDiagnosticsErr(ctx context.Context, diagnostics *diag.Diagnostics, err error, op string) {
	tflog.Trace(ctx, fmt.Sprintf("unable to %s: %s", op, err))
	diagnostics.AddError(
		fmt.Sprintf("Unable to %s", op),
		fmt.Sprintf("Error: %s\nPlease make sure your configuration is correct, and that the auth token has permissions for this org.", err),
	)
}
