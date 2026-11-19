// interface.go pins the OpenTofu workspace to the terraform workspace
// interface. The pkg/terraform/run and pkg/terraform/teardown pipelines are
// orchestrated against pkg/terraform/workspace.Workspace and contain no
// terraform-CLI calls, so an OpenTofu workspace implementing that exact
// interface can flow through them unchanged. See the package comment in
// workspace.go for why the interface is implemented cross-package rather than
// re-declared here.
package workspace

import (
	tfworkspace "github.com/nuonco/nuon/pkg/terraform/workspace"
)

var _ tfworkspace.Workspace = (*workspace)(nil)
