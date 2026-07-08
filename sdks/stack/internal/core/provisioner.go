package core

import (
	"context"
	"log/slog"
)

// Provisioner is the contract a single provisioning method implements (AWS
// SDK, Terraform module, CloudFormation template). The public stack package
// drives it: it owns run reporting and log-stream wiring, and delegates the
// actual resource lifecycle to the selected method.
//
// log emits under the "oteljob" scope (user-visible job output); sysLog emits
// under "system" (internal chatter the dashboard hides by default).
type Provisioner interface {
	// Provision creates or reconciles the install stack. It must be
	// idempotent: a re-run (kind reprovision) on existing resources should
	// converge rather than duplicate. On success it returns fully-resolved
	// Outputs for the phone-home payload.
	Provision(ctx context.Context, log, sysLog *slog.Logger, cfg *Config, kind Kind) (*Outputs, error)

	// Deprovision tears down everything the method created for the install.
	Deprovision(ctx context.Context, log, sysLog *slog.Logger, cfg *Config) error

	// Status reports the current persisted Outputs for the install without
	// mutating any resources. Used for offline inspection.
	Status(ctx context.Context, cfg *Config) (*Outputs, error)

	// ReportsOwnRun indicates the method reports its own run status to ctl-api
	// (e.g. the Terraform module's phone-home). When true, the public stack
	// package skips its run_client reporting to avoid double-reporting. When
	// false, the stack package owns reporting via the stack-run endpoints.
	ReportsOwnRun() bool
}
