package stacks

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/generics"
)

// AWSCloudFormationOutputs are used to define the stack outputs for a stack.
// This type can be used from anywhere, and is intentionally set up for testing here.
type AWSCloudFormationOutputs struct {
	VPCID          string `mapstructure:"vpc_id"`
	AccountID      string `mapstructure:"account_id"`
	RunnerSubnet   string `mapstructure:"runner_subnet"`
	PrivateSubnets string `mapstructure:"runner_subnet"`

	ProvisionIAMRoleARN   string `mapstructure:"provision_iam_role_arn"`
	MaintenanceIAMRoleARN string `mapstructure:"maintenance_subnet"`
	DeprovisionIAMRoleARN string `mapstructure:"deprovision_iam_role_arn"`
}

// aWSStackOutputDataDecodeHook converts %v printed maps into proper map
func aWSStackOutputDataDecodeHook() mapstructure.DecodeHookFunc {
	return generics.StringToMapDecodeHook()
}

func DecodeAWSStackOutputData(raw pgtype.Hstore) (map[string]interface{}, error) {
	var result map[string]interface{}
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
			aWSStackOutputDataDecodeHook(),
		),
		WeaklyTypedInput: true,
		Result:           &result,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create aws decoder")
	}
	// Convert pgtype.Hstore to map[string]string first
	outputsMap := make(map[string]string)
	for key, value := range raw {
		if value != nil {
			outputsMap[key] = generics.FromPtrStr(value)
		}
	}

	if err := decoder.Decode(outputsMap); err != nil {
		return nil, errors.Wrap(err, "unable to decode input data")
	}

	return result, nil
}
