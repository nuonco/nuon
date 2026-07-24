package deprovisiondns

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/cockroachdb/errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	installdelegationdns "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/dns"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "deprovision-dns"

type Signal struct {
	InstallID string
}

var (
	_ signal.Signal              = &Signal{}
	_ signal.SignalWithAutoRetry = (*Signal)(nil)
)

func (s *Signal) AutoRetry() bool { return true }

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install id is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	state, err := activities.AwaitGetInstallStateByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install state")
	}

	if state == nil || state.Sandbox == nil {
		l.Info("no sandbox state found, skipping dns delegation deprovisioning", "install_id", s.InstallID)
		return nil
	}

	var outputs nuonDNSSandboxOutputs
	if err := mapstructure.Decode(state.Sandbox.Outputs, &outputs); err != nil {
		return errors.Wrap(err, "unable to parse nuon dns outputs")
	}

	if !outputs.DNS.Enabled || outputs.DNS.PublicDomain.Name == "" {
		l.Info("DNS not enabled or public domain not configured, skipping", "install_id", s.InstallID)
		return nil
	}

	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	if install.SandboxMode.Bool {
		l.Info("skipping dns delegation deprovisioning for sandbox install", "install_id", s.InstallID)
		return nil
	}

	l.Info("deprovisioning DNS delegation", "install_id", s.InstallID, "domain", outputs.DNS.PublicDomain.Name)
	_, err = installdelegationdns.AwaitDeprovisionDNSDelegation(ctx, &installdelegationdns.DeprovisionDNSDelegationRequest{
		Domain: outputs.DNS.PublicDomain.Name,
	})
	if err != nil {
		return errors.Wrap(err, "unable to deprovision dns delegation")
	}

	l.Info("successfully deprovisioned dns delegation", "install_id", s.InstallID)
	return nil
}
