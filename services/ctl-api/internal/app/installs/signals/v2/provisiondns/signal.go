package provisiondns

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	installdelegationdns "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/dns"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "provision-dns"

type Signal struct {
	InstallID string
}

var _ signal.Signal = &Signal{}
var _ signal.SignalWithAutoRetry = (*Signal)(nil)

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

type nuonDNSDomain struct {
	ZoneID      string   `mapstructure:"zone_id,omitempty"`
	Name        string   `mapstructure:"name,omitempty"`
	Nameservers []string `mapstructure:"nameservers,omitempty"`
}

type nuonDNSOutputs struct {
	Enabled      bool          `mapstructure:"enabled,omitempty"`
	PublicDomain nuonDNSDomain `mapstructure:"public_domain,omitempty"`
}

type nuonDNSSandboxOutputs struct {
	DNS nuonDNSOutputs `mapstructure:"nuon_dns"`
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	state, err := activities.AwaitGetInstallStateByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install state")
	}

	var outputs nuonDNSSandboxOutputs
	if err := mapstructure.Decode(state.Sandbox.Outputs, &outputs); err != nil {
		return errors.Wrap(err, "unable to parse nuon dns outputs")
	}

	if !outputs.DNS.Enabled {
		l.Info("DNS not enabled, skipping", "install_id", s.InstallID)
		return nil
	}

	if outputs.DNS.PublicDomain.Name == "" || len(outputs.DNS.PublicDomain.Nameservers) == 0 {
		l.Info("DNS public domain not configured, skipping", "install_id", s.InstallID)
		return nil
	}

	// Check if sandbox - skip DNS provisioning for sandbox installs
	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	if install.SandboxMode.Bool {
		l.Info("skipping dns delegation provisioning for sandbox install", "install_id", s.InstallID)
		return nil
	}

	// The ProvisionDNSDelegation child workflow handles the root domain check internally
	dnsReq := &installdelegationdns.ProvisionDNSDelegationRequest{
		Domain:      outputs.DNS.PublicDomain.Name,
		Nameservers: outputs.DNS.PublicDomain.Nameservers,
	}

	l.Info("provisioning DNS delegation", "install_id", s.InstallID, "domain", outputs.DNS.PublicDomain.Name)
	_, err = installdelegationdns.AwaitProvisionDNSDelegation(ctx, dnsReq, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-provision-dns", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		return errors.Wrap(err, "unable to provision dns delegation")
	}

	l.Info("successfully provisioned dns delegation", "install_id", s.InstallID)
	return nil
}
