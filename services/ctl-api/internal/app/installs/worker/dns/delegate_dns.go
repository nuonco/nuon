package installdelegationdns

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53_types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/activity"
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

type DeleteDNSRequest struct {
	DNSAccessIAMRoleARN string `validate:"required"`
	ZoneID              string `validate:"required"`

	Domain string `validate:"required"`
}

type DeleteDNSResponse struct{}

// @temporal-gen-v2 activity
// @schedule-to-close-timeout 1m
func (a *Activities) DeleteDNS(ctx context.Context, req DeleteDNSRequest) (DeleteDNSResponse, error) {
	switch a.cfg.CloudProvider {
	case cloudProviderGCP:
		if err := a.deleteCloudDNSRecords(ctx, req); err != nil {
			return DeleteDNSResponse{}, fmt.Errorf("unable to delete cloud dns records: %w", err)
		}
		return DeleteDNSResponse{}, nil
	case cloudProviderAzure:
		activity.GetLogger(ctx).Info("dns delegation not implemented for azure, skipping", "domain", req.Domain)
		return DeleteDNSResponse{}, nil
	case cloudProviderAWS, "":
		if err := a.deleteRoute53(ctx, req); err != nil {
			return DeleteDNSResponse{}, fmt.Errorf("unable to delete cloud dns records: %w", err)
		}
		return DeleteDNSResponse{}, nil
	default:
		return DeleteDNSResponse{}, fmt.Errorf("cloud provider not supported for dns delegation: %q", a.cfg.CloudProvider)
	}
}

func (a *Activities) deleteRoute53(ctx context.Context, req DeleteDNSRequest) error {
	client, err := a.getRoute53Client(ctx, req.DNSAccessIAMRoleARN)
	if err != nil {
		return fmt.Errorf("unable to delete dns records: %w", err)
	}

	if err := a.deleteDNSRecords(ctx, client, req); err != nil {
		return fmt.Errorf("unable to delete dns records: %w", err)
	}

	return nil
}

func (a *Activities) deleteCloudDNSRecords(ctx context.Context, req DeleteDNSRequest) error {
	ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return fmt.Errorf("unable to get GCP token source: %w", err)
	}

	svc, err := googledns.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("unable to create Cloud DNS client: %w", err)
	}

	domain := ensureTrailingDot(req.Domain)

	// Cloud DNS delete requires the current record set; if it's already gone the
	// delegation is clean, so treat NotFound as success to keep retries idempotent.
	existing, err := svc.ResourceRecordSets.Get(a.cfg.ManagementAccountID, req.ZoneID, domain, "NS").Context(ctx).Do()
	switch {
	case isNotFound(err):
		return nil
	case err != nil:
		return fmt.Errorf("unable to look up existing Cloud DNS NS record: %w", err)
	}

	change := &googledns.Change{Deletions: []*googledns.ResourceRecordSet{existing}}
	if _, err := svc.Changes.Create(a.cfg.ManagementAccountID, req.ZoneID, change).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to delete Cloud DNS NS record: %w", err)
	}

	return nil
}

func (a *Activities) deleteDNSRecords(ctx context.Context, client route53Client, req DeleteDNSRequest) error {
	name := ensureTrailingDot(req.Domain)

	// Route53 delete must match the existing record set exactly (values + TTL),
	// so look it up first. A missing record means the delegation is already
	// clean, so return without error to keep retries idempotent.
	out, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
		HostedZoneId:    generics.ToPtr(req.ZoneID),
		StartRecordName: generics.ToPtr(name),
		StartRecordType: route53_types.RRTypeNs,
		MaxItems:        generics.ToPtr(int32(1)),
	})
	if err != nil {
		return fmt.Errorf("unable to list resource record sets: %w", err)
	}

	var target *route53_types.ResourceRecordSet
	for i := range out.ResourceRecordSets {
		rr := out.ResourceRecordSets[i]
		if rr.Type == route53_types.RRTypeNs && strings.EqualFold(generics.FromPtrStr(rr.Name), name) {
			target = &out.ResourceRecordSets[i]
			break
		}
	}
	if target == nil {
		return nil
	}

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53_types.ChangeBatch{
			Changes: []route53_types.Change{
				{
					Action:            route53_types.ChangeActionDelete,
					ResourceRecordSet: target,
				},
			},
		},
		HostedZoneId: generics.ToPtr(req.ZoneID),
	}

	if _, err := client.ChangeResourceRecordSets(ctx, params); err != nil {
		return fmt.Errorf("unable to change resource record sets: %w", err)
	}

	return nil
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
