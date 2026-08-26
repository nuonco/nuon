package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type dnsCheckDomain struct {
	Name        string   `mapstructure:"name,omitempty"`
	Nameservers []string `mapstructure:"nameservers,omitempty"`
}

type dnsCheckOutputs struct {
	Enabled      bool           `mapstructure:"enabled,omitempty"`
	PublicDomain dnsCheckDomain `mapstructure:"public_domain,omitempty"`
}

type dnsCheckSandboxOutputs struct {
	DNS dnsCheckOutputs `mapstructure:"nuon_dns"`
}

type CheckInstallDNSDelegationResponse struct {
	Domain              string   `json:"domain"`
	Enabled             bool     `json:"enabled"`
	Delegated           bool     `json:"delegated"`
	ExpectedNameservers []string `json:"expected_nameservers"`
	ObservedNameservers []string `json:"observed_nameservers"`
	MissingNameservers  []string `json:"missing_nameservers"`
	ExtraNameservers    []string `json:"extra_nameservers"`
	Message             string   `json:"message"`
}

// @ID						CheckInstallDNSDelegation
// @Summary				Check whether an install's public DNS delegation is live.
// @Description			Resolves the install's public domain nameservers from the public internet and compares them to the nameservers Nuon provisioned, confirming whether the customer's registrar delegation has taken effect.
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	CheckInstallDNSDelegationResponse
// @Router					/v1/installs/{install_id}/dns/check [get]
func (s *service) CheckInstallDNSDelegation(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	is, err := s.helpers.GetInstallState(ctx, install.ID, true, true)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install state: %w", err))
		return
	}

	var outputs dnsCheckSandboxOutputs
	if is.Sandbox != nil {
		if err := mapstructure.Decode(is.Sandbox.Outputs, &outputs); err != nil {
			ctx.Error(fmt.Errorf("unable to parse nuon dns outputs: %w", err))
			return
		}
	}

	resp := s.checkDNSDelegation(ctx, outputs.DNS)
	ctx.JSON(http.StatusOK, resp)
}

func (s *service) checkDNSDelegation(ctx context.Context, dns dnsCheckOutputs) CheckInstallDNSDelegationResponse {
	resp := CheckInstallDNSDelegationResponse{
		Domain:              dns.PublicDomain.Name,
		Enabled:             dns.Enabled,
		ExpectedNameservers: normalizeNameservers(dns.PublicDomain.Nameservers),
	}

	if !dns.Enabled {
		resp.Message = "DNS is not enabled for this install; nothing to delegate."
		return resp
	}
	if resp.Domain == "" || len(resp.ExpectedNameservers) == 0 {
		resp.Message = "install has no public domain provisioned yet; wait for the sandbox to finish provisioning."
		return resp
	}

	lookupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	nss, err := net.DefaultResolver.LookupNS(lookupCtx, resp.Domain)
	if err != nil {
		resp.Message = fmt.Sprintf("no NS records resolve for %s yet — the registrar delegation is not live: %v", resp.Domain, err)
		resp.MissingNameservers = resp.ExpectedNameservers
		return resp
	}

	observed := make([]string, 0, len(nss))
	for _, ns := range nss {
		observed = append(observed, ns.Host)
	}
	resp.ObservedNameservers = normalizeNameservers(observed)

	expectedSet := toSet(resp.ExpectedNameservers)
	observedSet := toSet(resp.ObservedNameservers)
	for _, ns := range resp.ExpectedNameservers {
		if _, ok := observedSet[ns]; !ok {
			resp.MissingNameservers = append(resp.MissingNameservers, ns)
		}
	}
	for _, ns := range resp.ObservedNameservers {
		if _, ok := expectedSet[ns]; !ok {
			resp.ExtraNameservers = append(resp.ExtraNameservers, ns)
		}
	}

	resp.Delegated = len(resp.MissingNameservers) == 0 && len(resp.ObservedNameservers) > 0
	if resp.Delegated {
		resp.Message = fmt.Sprintf("delegation is live: %s resolves to the expected nameservers.", resp.Domain)
	} else {
		resp.Message = fmt.Sprintf("delegation is NOT live: %s is missing %d expected nameserver(s) — the customer still needs to add the NS records at their registrar.", resp.Domain, len(resp.MissingNameservers))
	}

	return resp
}

func normalizeNameservers(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, ns := range in {
		n := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(ns), "."))
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func toSet(in []string) map[string]struct{} {
	set := make(map[string]struct{}, len(in))
	for _, v := range in {
		set[v] = struct{}{}
	}
	return set
}
