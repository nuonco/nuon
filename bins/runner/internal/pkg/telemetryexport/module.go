package telemetryexport

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/audit"
)

var Module = fx.Options(
	fx.Provide(newAWSFactory, newAzureFactory, newGCPFactory, newConfigSourceResolver, New, NewVendor, asAuditRouteLifecycle),
	fx.Invoke(func(*Supervisor, *VendorSupervisor) {}),
)

type auditRouteLifecycle struct{}

func (auditRouteLifecycle) AuditRouteLifecycle() {}

func asAuditRouteLifecycle(*Supervisor, *VendorSupervisor) audit.LocalRouteLifecycle {
	return auditRouteLifecycle{}
}
