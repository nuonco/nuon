package dns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53_types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
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

func (a *Activities) DelegateDNS(ctx context.Context, req DelegateDNSRequest) (DelegateDNSResponse, error) {
	client, err := a.getRoute53Client(ctx, req.DNSAccessIAMRoleARN)
	if err != nil {
		return DelegateDNSResponse{}, fmt.Errorf("unable to upsert dns records: %w", err)
	}

	if err := a.upsertDNSRecords(ctx, client, req); err != nil {
		return DelegateDNSResponse{}, fmt.Errorf("unable to upsert dns records: %w", err)
	}

	return DelegateDNSResponse{}, nil
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
