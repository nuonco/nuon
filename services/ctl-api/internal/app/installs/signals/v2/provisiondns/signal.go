package provisiondns

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/activities/installcommon"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/activities/installdelegationdns"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queues/signal"
)

type Signal struct {
	InstallID string
}

var _ signal.Signal = &Signal{}

func (s *Signal) Type() signal.SignalType {
	return signal.SignalTypeInstallProvisionDNS
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

type NuonDNSOutputs struct {
	Delegation []string `mapstructure:"delegation"`
	Domain     string   `mapstructure:"domain"`
	HostedZone string   `mapstructure:"hosted_zone"`
	Nameserver string   `mapstructure:"nameserver"`
}

type NuonDNSSandboxOutputs struct {
	DNS NuonDNSOutputs `mapstructure:"nuon_dns"`
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)
	l.Info("executing provision dns signal", zap.String("install_id", s.InstallID))

	// Get install state
	installStateResp, err := installcommon.AwaitGetState(ctx, &installcommon.GetStateRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install state: %w", err)
	}

	if installStateResp.Install.Metadata.SandboxMode {
		l.Info("sandbox mode enabled, skipping DNS delegation provisioning")
		return nil
	}

	if !installStateResp.Install.Metadata.EnableDNS {
		l.Info("DNS not enabled, skipping DNS delegation provisioning")
		return nil
	}

	if installStateResp.Install.Metadata.DNSRootDomain == "" {
		return fmt.Errorf("dns is enabled but dns root domain is not set")
	}

	// Parse DNS outputs from sandbox
	var dnsOutputs NuonDNSSandboxOutputs
	err = mapstructure.Decode(installStateResp.Install.SandboxOutputs, &dnsOutputs)
	if err != nil {
		return fmt.Errorf("unable to decode dns outputs: %w", err)
	}

	// Check if DNS domain matches root domain
	if dnsOutputs.DNS.Domain != installStateResp.Install.Metadata.DNSRootDomain {
		l.Info("dns domain does not match root domain, skipping DNS delegation provisioning",
			zap.String("dns_domain", dnsOutputs.DNS.Domain),
			zap.String("root_domain", installStateResp.Install.Metadata.DNSRootDomain))
		return nil
	}

	// Provision DNS delegation
	_, err = installdelegationdns.AwaitProvisionDNSDelegation(ctx, &installdelegationdns.ProvisionDNSDelegationRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to provision dns delegation: %w", err)
	}

	l.Info("successfully provisioned dns delegation", zap.String("install_id", s.InstallID))
	return nil
}
