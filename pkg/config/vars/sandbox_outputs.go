package vars

import (
	"context"

	"github.com/mitchellh/mapstructure"
	sandboxes "github.com/nuonco/sandboxes/pkg/sandboxes"
	awsecs "github.com/nuonco/sandboxes/pkg/sandboxes/aws-ecs"
	awseks "github.com/nuonco/sandboxes/pkg/sandboxes/aws-eks"
	azureaks "github.com/nuonco/sandboxes/pkg/sandboxes/azure-aks"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (v *varsValidator) getSandboxOutputs(ctx context.Context) (map[string]interface{}, error) {
	vcsCfg := v.cfg.Sandbox.PublicRepo
	if vcsCfg == nil {
		// TODO(jm): print to stdout from here to show log messages
		return map[string]interface{}{}, nil
	}
	switch vcsCfg.Directory {
	case "aws-ecs", "aws-ecs-byovpc":
		obj, err := v.awsECSSandboxOutputs()
		if err != nil {
			return nil, errors.Wrap(err, "unable to create aws-ecs sandbox outputs")
		}
		return obj, nil
	case "aws-eks", "aws-eks-byovpc":
		obj, err := v.awsEKSSandboxOutputs()
		if err != nil {
			return nil, errors.Wrap(err, "unable to create aws-eks sandbox outputs")
		}
		return obj, nil
	case "azure-aks":
		obj, err := v.azureAKSSandboxOutputs()
		if err != nil {
			return nil, errors.Wrap(err, "unable to create azure-aks sandbox outputs")
		}
		return obj, nil
	}

	// TODO(jm): add a warning statement here once the cli logger is in this context
	return map[string]interface{}{}, nil
}

func (v *varsValidator) awsECSSandboxOutputs() (map[string]interface{}, error) {
	obj := &awsecs.TerraformOutputs{
		PublicDomain:   v.domainOutputs(),
		InternalDomain: v.domainOutputs(),
		ECSCluster:     generics.GetFakeObj[awsecs.ECSClusterOutputs](),
		VPC:            v.vpcOutputs(),
		ECR:            generics.GetFakeObj[sandboxes.ECROutputs](),
		Runner:         generics.GetFakeObj[awsecs.RunnerOutputs](),
	}

	data := make(map[string]interface{})
	if err := mapstructure.Decode(obj, &data); err != nil {
		return nil, errors.Wrap(err, "unable to convert to mapstructure")
	}

	return data, nil
}

func (v *varsValidator) domainOutputs() sandboxes.DomainOutputs {
	return sandboxes.DomainOutputs{
		Nameservers: []interface{}{"abc"},
		Name:        generics.GetFakeObj[string](),
		ZoneID:      generics.GetFakeObj[string](),
		ID:          generics.GetFakeObj[string](),
	}
}

func (v *varsValidator) vpcOutputs() sandboxes.VPCOutputs {
	return sandboxes.VPCOutputs{
		Name:                    generics.GetFakeObj[string](),
		ID:                      generics.GetFakeObj[string](),
		CIDR:                    generics.GetFakeObj[string](),
		AZs:                     []interface{}{"a"},
		PrivateSubnetCidrBlocks: []interface{}{"a"},
		PrivateSubnetIDs:        []interface{}{"a"},
		PublicSubnetIDs:         []interface{}{"a"},
		PublicSubnetCidrBlocks:  []interface{}{"a"},
		DefaultSecurityGroupID:  generics.GetFakeObj[string](),
	}
}

func (v *varsValidator) awsEKSSandboxOutputs() (map[string]interface{}, error) {
	obj := &awseks.TerraformOutputs{
		PublicDomain:   v.domainOutputs(),
		InternalDomain: v.domainOutputs(),
		Cluster:        generics.GetFakeObj[awseks.ClusterOutputs](),
		VPC:            v.vpcOutputs(),
		ECR:            generics.GetFakeObj[sandboxes.ECROutputs](),
		Runner:         generics.GetFakeObj[awseks.RunnerOutputs](),
	}

	data := make(map[string]interface{})
	if err := mapstructure.Decode(obj, &data); err != nil {
		return nil, errors.Wrap(err, "unable to convert to mapstructure")
	}

	return data, nil
}

func (v *varsValidator) azureAKSSandboxOutputs() (map[string]interface{}, error) {
	obj := generics.GetFakeObj[azureaks.TerraformOutputs]()
	data := make(map[string]interface{})
	if err := mapstructure.Decode(obj, &data); err != nil {
		return nil, errors.Wrap(err, "unable to convert to mapstructure")
	}

	return data, nil
}
