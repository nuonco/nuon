package installdelegationdns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53_types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2/google"
	googledns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/nuonco/nuon/pkg/generics"
)

type DelegateDNSRequest struct {
	DNSAccessIAMRoleARN string `validate:"required"`
	ZoneID              string `validate:"required"`

	Domain      string   `validate:"required"`
	NameServers []string `validate:"required"`
}

func (d DelegateDNSRequest) validate() error {
	validate := validator.New()
	return validate.Struct(d)
}

type DelegateDNSResponse struct{}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 1m
func (a *Activities) DelegateDNS(ctx context.Context, req DelegateDNSRequest) (DelegateDNSResponse, error) {
	if a.cfg.IsGCP() {
		if err := a.upsertCloudDNSRecords(ctx, req); err != nil {
			return DelegateDNSResponse{}, fmt.Errorf("unable to upsert cloud dns records: %w", err)
		}
		return DelegateDNSResponse{}, nil
	}

	client, err := a.getRoute53Client(ctx, req.DNSAccessIAMRoleARN)
	if err != nil {
		return DelegateDNSResponse{}, fmt.Errorf("unable to upsert dns records: %w", err)
	}

	if err := a.upsertDNSRecords(ctx, client, req); err != nil {
		return DelegateDNSResponse{}, fmt.Errorf("unable to upsert dns records: %w", err)
	}

	return DelegateDNSResponse{}, nil
}

func (a *Activities) upsertCloudDNSRecords(ctx context.Context, req DelegateDNSRequest) error {
	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return fmt.Errorf("unable to get GCP token source: %w", err)
	}

	svc, err := googledns.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("unable to create Cloud DNS client: %w", err)
	}

	// Cloud DNS requires both the record name and NS rrdata to be
	// fully-qualified with a trailing dot; Route53 hands back nameservers
	// without one.
	nameservers := make([]string, len(req.NameServers))
	for i, ns := range req.NameServers {
		nameservers[i] = ensureTrailingDot(ns)
	}

	domain := ensureTrailingDot(req.Domain)
	record := &googledns.ResourceRecordSet{
		Name:    domain,
		Type:    "NS",
		Ttl:     3600,
		Rrdatas: nameservers,
	}
	change := &googledns.Change{Additions: []*googledns.ResourceRecordSet{record}}

	// Cloud DNS has no native upsert; if an NS record set already exists for
	// this name, replace it by deleting the current one in the same change so
	// retries and reprovisions are idempotent.
	// req.ZoneID holds the Cloud DNS managed zone name (set via DNS_ZONE_ID on GCP)
	existing, err := svc.ResourceRecordSets.Get(a.cfg.ManagementAccountID, req.ZoneID, domain, "NS").Context(ctx).Do()
	switch {
	case err == nil:
		if existing.Ttl == record.Ttl && sameRecords(existing.Rrdatas, nameservers) {
			return nil
		}
		change.Deletions = []*googledns.ResourceRecordSet{existing}
	case !isNotFound(err):
		return fmt.Errorf("unable to look up existing Cloud DNS NS record: %w", err)
	}

	if _, err := svc.Changes.Create(a.cfg.ManagementAccountID, req.ZoneID, change).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to upsert Cloud DNS NS record: %w", err)
	}

	return nil
}

func isNotFound(err error) bool {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		return gerr.Code == http.StatusNotFound
	}
	return false
}

func sameRecords(a, b []string) bool {
	a = slices.Clone(a)
	b = slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

func (a *Activities) upsertDNSRecords(ctx context.Context, client route53Client, req DelegateDNSRequest) error {
	records := make([]route53_types.ResourceRecord, 0)
	for _, ns := range req.NameServers {
		records = append(records, route53_types.ResourceRecord{
			Value: generics.ToPtr(ns),
		})
	}
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53_types.ChangeBatch{
			Changes: []route53_types.Change{
				{
					Action: route53_types.ChangeActionUpsert,
					ResourceRecordSet: &route53_types.ResourceRecordSet{
						Name:            generics.ToPtr(req.Domain),
						Type:            route53_types.RRTypeNs,
						ResourceRecords: records,
						TTL:             generics.ToPtr(int64(3600)),
					},
				},
			},
		},
		HostedZoneId: generics.ToPtr(req.ZoneID),
	}

	_, err := client.ChangeResourceRecordSets(ctx, params)
	if err != nil {
		return fmt.Errorf("unable to change resource record sets: %w", err)
	}

	return nil
}
