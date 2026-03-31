package installdelegationdns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53_types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2/google"
	googledns "google.golang.org/api/dns/v1"
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
	if a.cfg.CloudProvider == "gcp" {
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

	// ensure domain has trailing dot (Cloud DNS format)
	domain := req.Domain
	if len(domain) > 0 && domain[len(domain)-1] != '.' {
		domain += "."
	}

	change := &googledns.Change{
		Additions: []*googledns.ResourceRecordSet{
			{
				Name:    domain,
				Type:    "NS",
				Ttl:     3600,
				Rrdatas: req.NameServers,
			},
		},
	}

	// req.ZoneID holds the Cloud DNS managed zone name (set via DNS_ZONE_ID on GCP)
	_, err = svc.Changes.Create(a.cfg.ManagementAccountID, req.ZoneID, change).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to create Cloud DNS NS record: %w", err)
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
