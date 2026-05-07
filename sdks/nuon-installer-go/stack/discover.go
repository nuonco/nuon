package stack

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

// installIDTagKey is the tag we set on every taggable AWS resource so we can
// discover it without consulting the local state file.
const installIDTagKey = "nuon:install_id"

// IsAWSErrCode reports whether err is an AWS API error with the given code.
func IsAWSErrCode(err error, code string) bool {
	var ae smithy.APIError
	return errors.As(err, &ae) && ae.ErrorCode() == code
}

// findEC2ResourcesByInstallID returns IDs of EC2 resources of the given type
// tagged for this install. Order is not guaranteed.
func findEC2ResourcesByInstallID(ctx context.Context, c *ec2.Client, rt ec2types.ResourceType, installID string) ([]string, error) {
	out, err := c.DescribeTags(ctx, &ec2.DescribeTagsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("resource-type"), Values: []string{string(rt)}},
			{Name: aws.String("key"), Values: []string{installIDTagKey}},
			{Name: aws.String("value"), Values: []string{installID}},
		},
	})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(out.Tags))
	for _, t := range out.Tags {
		ids = append(ids, aws.ToString(t.ResourceId))
	}
	return ids, nil
}
