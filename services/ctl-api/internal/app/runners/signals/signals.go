package signals

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	TemporalNamespace string = "runners"
)

const (
	OperationCreated           eventloop.SignalType = "created"
	OperationRestart           eventloop.SignalType = "restart"
	OperationDelete            eventloop.SignalType = "delete"
	OperationForceDelete       eventloop.SignalType = "force_delete"
	OperationProcessJob        eventloop.SignalType = "process_job"
	OperationUpdateVersion     eventloop.SignalType = "update_version"
	OperationHealthcheck       eventloop.SignalType = "healthcheck"
	OperationOfflineCheck      eventloop.SignalType = "offline_check"
	OperationGracefulShutdown  eventloop.SignalType = "graceful_shutdown"
	OperationForceShutdown     eventloop.SignalType = "force_shutdown"
	OperationFlushOrphanedJobs eventloop.SignalType = "flush_orphaned_jobs"

	// used for v2 provisioning
	OperationProvisionServiceAccount   eventloop.SignalType = "provision_service_account"
	OperationReprovisionServiceAccount eventloop.SignalType = "reprovision_service_account"
	OperationInstallStackVersionRun    eventloop.SignalType = "install_stack_version_run"

	// used for management operations
	OperationMngVMShutDown eventloop.SignalType = "mng-shutdown-vm"
	OperationMngShutDown   eventloop.SignalType = "mng-shutdown"
	OperationMngUpdate     eventloop.SignalType = "mng-update"

	// used for internal provisioning
	OperationProvision   eventloop.SignalType = "provision"
	OperationDeprovision eventloop.SignalType = "deprovision"
	OperationReprovision eventloop.SignalType = "reprovision"
)

type Signal struct {
	Type eventloop.SignalType `validate:"required"`
	eventloop.BaseSignal

	JobID                    string `validate:"required_if=Type job_queued"`
	HealthCheckID            string `validate:"required_if=Type update_version"`
	InstallStackVersionRunID string `validate:"required_if=Type install_stack_version_run"`
	ForceDelete              bool
}

type RequestSignal struct {
	*Signal
	eventloop.EventLoopRequest
}

func NewRequestSignal(req eventloop.EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: req,
	}
}

var _ eventloop.Signal = (*Signal)(nil)

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

func (s *Signal) SignalType() eventloop.SignalType {
	return s.Type
}

func (s *Signal) Namespace() string {
	return TemporalNamespace
}

func (s *Signal) Name() string {
	return string(s.Type)
}

func (s *Signal) Stop() bool {
	switch s.Type {
	case OperationDelete:
		return true
	case OperationForceDelete:
		return true
	default:
	}

	return false
}

func (s *Signal) Restart() bool {
	switch s.Type {
	case OperationRestart:
		return true
	default:
	}

	return false
}

func (s *Signal) Start() bool {
	switch s.SignalType() {
	case OperationCreated:
		return true
	default:
	}

	return false
}

func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	runner := app.Runner{}
	res := db.WithContext(ctx).
		Preload("Org").
		First(&runner, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner.Org, nil
}
