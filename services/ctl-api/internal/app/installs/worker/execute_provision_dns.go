package worker

import (
	"fmt"
	"strings"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

type NuonDNSChart struct {
	Enabled  bool   `mapstructure:"enabled,omitempty"`
	ID       string `mapstructure:"id,omitempty"`
	Chart    string `mapstructure:"chart,omitempty"`
	Revision string `mapstructure:"revision,omitempty"`
}

type NuonDNSSandboxOutputs struct {
	DNS NuonDNSOutputs `mapstructure:"nuon_dns"`
}

type NuonDNSDomain struct {
	ZoneID      string   `mapstructure:"zone_id,omitempty"`
	Name        string   `mapstructure:"name,omitempty"`
	Nameservers []string `mapstructure:"nameservers,omitempty"`
}

type NuonDNSOutputs struct {
	Enabled bool `mapstructure:"enabled,omitempty"`

	PublicDomain   NuonDNSDomain `mapstructure:"public_domain,omitempty"`
	InternalDomain NuonDNSDomain `mapstructure:"internal_domain,omitempty"`

	ALBIngressController NuonDNSChart `mapstructure:"alb_ingress_controller,omitempty"`
	ExternalDNS          NuonDNSChart `mapstructure:"external_dns,omitempty"`
	CertManager          NuonDNSChart `mapstructure:"cert_manager,omitempty"`
	IngressNginx         NuonDNSChart `mapstructure:"ingress_nginx,omitempty"`
}

// @temporal-gen workflow
// @execution-timeout 60m
// @execution-timeout 30m
func (w *Workflows) ProvisionDNS(ctx workflow.Context, sreq signals.RequestSignal) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("configuring DNS for nuon.run domain if enabled")
	state, err := activities.AwaitGetInstallStateByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install state")
	}

	install, err := activities.AwaitGetByInstallID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get install id")
	}

	var outputs NuonDNSSandboxOutputs
	if err := mapstructure.Decode(state.Sandbox.Outputs, &outputs); err != nil {
		return errors.Wrap(err, "unable to parse nuon dns")
	}

	if !outputs.DNS.Enabled {
		return nil
	}
	if !strings.Contains(outputs.DNS.PublicDomain.Name, "nuon.run") {
		return nil
	}

	dnsReq := &executors.ProvisionDNSDelegationRequest{
		WorkflowID:  fmt.Sprintf("%s-provision-dns", workflow.GetInfo(ctx).WorkflowExecution.ID),
		Domain:      outputs.DNS.PublicDomain.Name,
		Nameservers: outputs.DNS.PublicDomain.Nameservers,
	}

	l.Info("provisioning nuon.run root domain")
	if !sreq.SandboxMode {
		_, err = executors.AwaitProvisionDNSDelegation(ctx, dnsReq)
		if err != nil {
			return errors.Wrap(err, "unable to provision dns delegation")
		}
		l.Info("successfully provisioned dns delegation")
	} else {
		l.Info("skipping dns delegation provisioning",
			zap.Any("install_id", install.ID))
	}

	return nil
}
