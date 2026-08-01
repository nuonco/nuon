package orgs

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// ParseClaimConditions parses repeated `claim=pattern` flag values.
func ParseClaimConditions(raw []string) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	conditions := map[string]string{}
	for _, entry := range raw {
		claim, pattern, found := strings.Cut(entry, "=")
		if !found || claim == "" {
			return nil, fmt.Errorf("invalid --claim %q: expected claim=pattern", entry)
		}
		conditions[claim] = pattern
	}

	return conditions, nil
}

func renderClaimConditions(conditions map[string]string) string {
	if len(conditions) == 0 {
		return ""
	}

	parts := make([]string, 0, len(conditions))
	for claim, pattern := range conditions {
		parts = append(parts, fmt.Sprintf("%s=%s", claim, pattern))
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

func (s *Service) CreateOIDCTrustPolicy(ctx context.Context, name, issuer, audience, role string, ttl int64, conditions map[string]string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	policy, err := s.api.CreateOIDCTrustPolicy(ctx, &models.ServiceCreateOIDCTrustPolicyRequest{
		Name:                 &name,
		IssuerURL:            &issuer,
		Audience:             &audience,
		ClaimConditions:      conditions,
		Role:                 role,
		TokenDurationSeconds: ttl,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(policy)
		return nil
	}

	view.Render([][]string{
		{"id", policy.ID},
		{"name", policy.Name},
		{"issuer", policy.IssuerURL},
		{"audience", policy.Audience},
		{"claims", renderClaimConditions(policy.ClaimConditions)},
		{"role", policy.Role},
		{"token duration", fmt.Sprintf("%ds", policy.TokenDurationSeconds)},
	})
	return nil
}

func (s *Service) ListOIDCTrustPolicies(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewListView()

	policies, err := s.api.ListOIDCTrustPolicies(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(policies)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"ENABLED",
			"ISSUER",
			"ROLE",
			"LAST USED",
		},
	}
	for _, p := range policies {
		data = append(data, []string{
			p.ID,
			p.Name,
			strconv.FormatBool(p.Enabled),
			p.IssuerURL,
			p.Role,
			p.LastUsedAt,
		})
	}

	view.Render(data)
	return nil
}

func (s *Service) GetOIDCTrustPolicy(ctx context.Context, policyID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	policy, err := s.api.GetOIDCTrustPolicy(ctx, policyID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(policy)
		return nil
	}

	view.Render([][]string{
		{"id", policy.ID},
		{"name", policy.Name},
		{"enabled", strconv.FormatBool(policy.Enabled)},
		{"issuer", policy.IssuerURL},
		{"audience", policy.Audience},
		{"claims", renderClaimConditions(policy.ClaimConditions)},
		{"role", policy.Role},
		{"token duration", fmt.Sprintf("%ds", policy.TokenDurationSeconds)},
		{"service account", policy.ServiceAccountID},
		{"last used", policy.LastUsedAt},
		{"created at", policy.CreatedAt},
	})
	return nil
}

func (s *Service) UpdateOIDCTrustPolicy(ctx context.Context, policyID string, req *models.ServiceUpdateOIDCTrustPolicyRequest, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	policy, err := s.api.UpdateOIDCTrustPolicy(ctx, policyID, req)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(policy)
		return nil
	}

	view.Render([][]string{
		{"id", policy.ID},
		{"name", policy.Name},
		{"enabled", strconv.FormatBool(policy.Enabled)},
		{"role", policy.Role},
	})
	return nil
}

func (s *Service) DeleteOIDCTrustPolicy(ctx context.Context, policyID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	if err := s.api.DeleteOIDCTrustPolicy(ctx, policyID); err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(map[string]string{"id": policyID, "status": "deleted"})
		return nil
	}

	ui.PrintLn(fmt.Sprintf("deleted OIDC trust policy %s", policyID))
	return nil
}
