package deprovisiondns

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/activities/installcommon"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queues/signal"
)

type Signal struct {
	InstallID string
}

var _ signal.Signal = &Signal{}

func (s *Signal) Type() signal.SignalType {
	return signal.SignalTypeInstallDeprovisionDNS
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install id is required")
	}

	// Validate install exists
	_, err := installcommon.AwaitGet(ctx, &installcommon.GetRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)
	l.Info("deprovision dns is a no-op, domains must be manually deleted", zap.String("install_id", s.InstallID))
	return nil
}
