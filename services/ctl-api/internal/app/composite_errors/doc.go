// Package composite_errors is the umbrella for the typed CompositeError
// implementations registered into the catalog at process start.
//
// Each error type lives in its own subpackage under types/<name>/ and
// registers itself via init():
//
//   - aws_missing_iam_permission
//   - terraform_apply_failed
//   - unknown_error
//
// Importing this package directly does NOT register types — callers must
// import the desired type packages individually (or the umbrella `register`
// package, see register.go) so init() runs.
package composite_errors
