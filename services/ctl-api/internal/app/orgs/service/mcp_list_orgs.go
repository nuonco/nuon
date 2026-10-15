package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type mcpListOrgsInput struct {
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum orgs to return (default 20, max 100)"`
	Offset int    `json:"offset,omitempty" jsonschema:"skip this many orgs; use next_offset from a previous response when has_more is true"`
	Query  string `json:"query,omitempty" jsonschema:"optional case-insensitive match against org name or id"`
}

type mcpOrgListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mcpListOrgsResult struct {
	Orgs       []mcpOrgListItem `json:"orgs"`
	HasMore    bool             `json:"has_more"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
	NextOffset int              `json:"next_offset,omitempty"`
}

func (s *service) mcpListOrgs(ctx context.Context, _ *mcp.CallToolRequest, in mcpListOrgsInput) (*mcp.CallToolResult, any, error) {
	accountID := keys.CreatedByIDFromContext(ctx)

	var acct app.Account
	if res := s.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Org").
		First(&acct, "id = ?", accountID); res.Error != nil {
		return nil, nil, fmt.Errorf("unable to fetch account: %w", res.Error)
	}

	limit, offset, err := apiPkg.MCPListPage(in.Limit, in.Offset)
	if err != nil {
		return nil, nil, err
	}

	matched := filterOrgs(accountOrgs(acct), in.Query)
	page := []app.Org{}
	if offset < len(matched) {
		page = matched[offset:]
	}
	page, hasMore := apiPkg.MCPClipList(page, limit)

	orgs := make([]mcpOrgListItem, 0, len(page))
	for _, o := range page {
		orgs = append(orgs, mcpOrgListItem{ID: o.ID, Name: o.Name})
	}

	return apiPkg.MCPJSONResult(mcpListOrgsResult{
		Orgs:       orgs,
		HasMore:    hasMore,
		Limit:      limit,
		Offset:     offset,
		NextOffset: apiPkg.MCPNextOffset(offset, limit, hasMore),
	})
}

func accountOrgs(acct app.Account) []app.Org {
	orgs := make([]app.Org, 0, len(acct.Orgs))
	for _, o := range acct.Orgs {
		if o == nil || o.ID == "" {
			continue
		}
		orgs = append(orgs, *o)
	}
	slices.SortFunc(orgs, func(a, b app.Org) int {
		if c := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	return orgs
}

func filterOrgs(orgs []app.Org, query string) []app.Org {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return orgs
	}
	matched := make([]app.Org, 0, len(orgs))
	for _, o := range orgs {
		if strings.Contains(strings.ToLower(o.Name), q) || strings.Contains(strings.ToLower(o.ID), q) {
			matched = append(matched, o)
		}
	}
	return matched
}
