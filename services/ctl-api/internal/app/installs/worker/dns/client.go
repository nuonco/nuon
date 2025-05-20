package installdelegationdns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

const (
	defaultSessionName string = "workers-installs-dns"
)

type route53Client interface {
	ChangeResourceRecordSets(context.Context, *route53.ChangeResourceRecordSetsInput, ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
}

func (a *Activities) getRoute53Client(ctx context.Context, iamRoleARN string) (*route53.Client, error) {
	cfg, err := credentials.Fetch(ctx, &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     iamRoleARN,
			SessionName: defaultSessionName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get aws config: %w", err)
	}

	return route53.NewFromConfig(cfg), nil
}
