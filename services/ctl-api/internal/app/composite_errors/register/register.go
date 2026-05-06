// Package register is a convenience umbrella that registers every built-in
// CompositeError type into the catalog at process start.
//
// Importing this package with the blank identifier:
//
//	import _ "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/register"
//
// runs the init() of every type subpackage. Callers that only want a subset
// of types can import the individual packages directly.
package register

import (
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/types/aws_missing_iam_permission"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/types/terraform_apply_failed"
	_ "github.com/nuonco/nuon/services/ctl-api/internal/app/composite_errors/types/unknown_error"
)
