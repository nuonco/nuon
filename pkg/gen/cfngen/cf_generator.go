package cfngen

import (
	"encoding/json"
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/autoscaling"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/lambda"
	"github.com/awslabs/goformation/v7/cloudformation/secretsmanager"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
	awseks "github.com/nuonco/sandboxes/pkg/sandboxes/aws-eks"
	"github.com/nuonco/sandboxes/pkg/sandboxes/permissions"
)

type InternalValues struct {
	// Environment indicates the env we're creating for. Only current effect here is turning on certain behaviors, like SSH access, if this is "dev"
	Environment string `toml:"environment"`
	// InstallID is the ID of the install that the stack is being created for.
	InstallID string `toml:"install_id"`
	// RunnerID is the ID that the provisioned runner will run as.
	RunnerID string `toml:"runner_id"`
	// RunnerApiToken is the token the runner will use to authenticate with the API.
	RunnerApiToken string `toml:"runner_api_token"`
	// RunnerApiUrl is the base URL of the API the runner will interact with. Defaults to https://runner.nuon.co
	RunnerApiUrl string `toml:"runner_api_url"`
	// NotifyUrl is the URL that will be notified webhook-style when the stack is created, updated, or deleted.
	NotifyUrl string `toml:"notify_url"`
}

func GenerateCloudformation(cfg *AppConfigValues, intv InternalValues) (*cloudformation.Template, error) {
	tmpl := cloudformation.NewTemplate()

	if cfg.InstanceType == "" {
		cfg.InstanceType = "t3a.medium"
	}
	if intv.RunnerApiUrl == "" {
		intv.RunnerApiUrl = "https://runner.nuon.co"
	}
	if intv.NotifyUrl == "" {
		return nil, fmt.Errorf("notify url is required; use a pastebin like webhook.site to quickly probe outputs")
	}

	t := tagger{
		installID:  intv.InstallID,
		additional: cfg.AdditionalTags,
	}

	paramlabels := map[string]any{}
	var vpcpgroups map[string]any
	var roleparams []string
	var secretparams []string
	var subnets []string

	var sgs []string

	tmpl.Resources["RunnerEgressGroup"] = &ec2.SecurityGroup{
		GroupDescription: "Egress security group for the runner - allow all outbound traffic",
		VpcId:            ptr(cloudformation.Ref("RunnerVPC")),
		SecurityGroupEgress: []ec2.SecurityGroup_Egress{
			{
				CidrIp:     ptr("0.0.0.0/0"),
				FromPort:   ptr(-1),
				ToPort:     ptr(-1),
				IpProtocol: "-1",
			},
		},
		Tags: t.apply(nil, "egress-sg"),
	}

	if intv.Environment == "dev" {
		tmpl.Resources["RunnerSSHIngress"] = &ec2.SecurityGroup{
			GroupDescription: "Security group to allow ssh ingress traffic to the runner",
			VpcId:            ptr(cloudformation.Ref("RunnerVPC")),
			SecurityGroupIngress: []ec2.SecurityGroup_Ingress{
				{
					IpProtocol: "tcp",
					FromPort:   ptr(22),
					ToPort:     ptr(22),
					CidrIp:     ptr("0.0.0.0/0"),
				},
			},
			Tags: t.apply(nil, "ssh-ingress-sg"),
		}
		sgs = append(sgs, cloudformation.Ref("RunnerSSHIngress"))
	}
	sgs = append(sgs, cloudformation.Ref("RunnerEgressGroup"))

	inst := &ec2.LaunchTemplate_LaunchTemplateData{
		InstanceType: ptr(cfg.InstanceType),
		ImageId:      ptr(cloudformation.Sub("{{resolve:ssm:/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64}}")),
		IamInstanceProfile: &ec2.LaunchTemplate_IamInstanceProfile{
			Name: cloudformation.RefPtr("RunnerInstanceProfile"),
		},
		TagSpecifications: []ec2.LaunchTemplate_TagSpecification{
			{
				ResourceType: ptr("instance"),
				Tags: t.apply([]tags.Tag{
					{
						Key:   "nuon_runner_id",
						Value: intv.RunnerID,
					},
					{
						Key:   "nuon_runner_api_url",
						Value: intv.RunnerApiUrl,
					},
					{
						// TODO(sdboyer) remove this in favor of an API call to get the token
						Key:   "nuon_runner_api_token",
						Value: intv.RunnerApiToken,
					},
				}, "runner-instance"),
			},
			{
				ResourceType: ptr("network-interface"),
				Tags:         t.apply(nil, "runner-eni"),
			},
		},
		// in the beginning, there was a curlbash
		UserData: cloudformation.Base64Ptr(`#!/bin/bash
curl https://raw.githubusercontent.com/nuonco/aws-runner-init/refs/heads/main/init.sh | bash
`),
	}

	if cfg.BYOVPC {
		tmpl.Parameters["RunnerVPC"] = cloudformation.Parameter{
			Type:        "AWS::EC2::VPC::Id",
			Description: ptr("The VPC into which the app should be installed."),
			Default:     "",
		}
		tmpl.Parameters["SubnetId"] = cloudformation.Parameter{
			Type:        "AWS::EC2::Subnet::Id",
			Description: ptr("The subnet on which the app will run within the selected VPC."),
		}

		paramlabels["RunnerVPC"] = map[string]any{
			"default": "Target VPC",
		}
		paramlabels["SubnetId"] = map[string]any{
			"default": "Target Subnet",
		}

		inst.NetworkInterfaces = []ec2.LaunchTemplate_NetworkInterface{
			{
				AssociatePublicIpAddress: ptr(intv.Environment == "dev"),
				DeviceIndex:              ptr(0),
				SubnetId:                 ptr(cloudformation.Ref("SubnetId")),
				Groups:                   sgs,
				DeleteOnTermination:      ptr(true),
			},
		}

		subnets = []string{
			cloudformation.Ref("SubnetId"),
		}

		vpcpgroups = map[string]any{
			"Label": map[string]any{
				"default": "VPC",
			},
			"Parameters": []string{"RunnerVPC", "SubnetId"},
		}
	} else {
		// NOTE(sdboyer): this is how we expose VPC CIDR as a parameter, if we ever want to go back to that
		// tmpl.Parameters["VpcCIDR"] = cloudformation.Parameter{
		// 	Type:           "String",
		// 	Description:    ptr("CIDR block for a new VPC that will contain the app."),
		// 	AllowedPattern: ptr("^([0-9]{1,3}\\.){3}[0-9]{1,3}\\/(1[0-9]|2[0-4]|[1-9])$"),
		// 	Default:        "10.128.0.0/16",
		// }

		tmpl.Resources["RunnerVPC"] = &ec2.VPC{
			CidrBlock:          ptr("10.128.0.0/16"),
			EnableDnsHostnames: ptr(true),
			EnableDnsSupport:   ptr(true),
			Tags:               t.apply(nil, ""),
		}

		tmpl.Resources["RunnerGateway"] = &ec2.InternetGateway{
			Tags: t.apply(nil, ""),
		}

		tmpl.Resources["RunnerGatewayAttachment"] = &ec2.VPCGatewayAttachment{
			VpcId:             cloudformation.Ref("RunnerVPC"),
			InternetGatewayId: ptr(cloudformation.Ref("RunnerGateway")),
		}

		tmpl.Resources["RunnerPublicRouteTable"] = &ec2.RouteTable{
			VpcId: cloudformation.Ref("RunnerVPC"),
			Tags:  t.apply(nil, "public"),
		}

		tmpl.Resources["RunnerDefaultPublicRoute"] = &ec2.Route{
			RouteTableId:               cloudformation.Ref("RunnerPublicRouteTable"),
			DestinationCidrBlock:       ptr("0.0.0.0/0"),
			GatewayId:                  ptr(cloudformation.Ref("RunnerGateway")),
			AWSCloudFormationDependsOn: []string{"RunnerGatewayAttachment"},
		}

		tmpl.Resources["RunnerPrivateRouteTable"] = &ec2.RouteTable{
			VpcId: cloudformation.Ref("RunnerVPC"),
			Tags:  t.apply(nil, "private"),
		}

		tmpl.Resources["RunnerDefaultPrivateRoute"] = &ec2.Route{
			RouteTableId:               cloudformation.Ref("RunnerPrivateRouteTable"),
			DestinationCidrBlock:       ptr("0.0.0.0/0"),
			NatGatewayId:               ptr(cloudformation.Ref("RunnerNATGateway0")),
			AWSCloudFormationDependsOn: []string{"RunnerGatewayAttachment"},
		}

		for i, addr := range []string{"0", "64", "128"} {
			subnet := fmt.Sprintf("RunnerPublicSubnet%d", i)
			subtableassoc := fmt.Sprintf("RunnerPublicSubnet%dRouteTableAssociation", i)
			eip := fmt.Sprintf("RunnerEIP%d", i)
			ngwy := fmt.Sprintf("RunnerNATGateway%d", i)

			tmpl.Resources[subnet] = &ec2.Subnet{
				VpcId:            cloudformation.Ref("RunnerVPC"),
				CidrBlock:        ptr(fmt.Sprintf("10.128.0.%s/26", addr)),
				AvailabilityZone: cloudformation.SelectPtr(i, cloudformation.GetAZs("")),
				Tags: t.apply([]tags.Tag{
					{
						Key:   fmt.Sprintf("kubernetes.io/cluster/%s", intv.InstallID),
						Value: "shared",
					},
					{
						Key:   "kubernetes.io/role/elb",
						Value: "1",
					},
					{
						Key:   "visibility",
						Value: "public",
					},
				}, fmt.Sprintf("public-%d", i)),
			}
			tmpl.Resources[eip] = &ec2.EIP{
				Domain: ptr("vpc"),
				Tags:   t.apply(nil, fmt.Sprintf("eip-%d", i)),
			}

			tmpl.Resources[ngwy] = &ec2.NatGateway{
				SubnetId:     cloudformation.Ref(subnet),
				AllocationId: cloudformation.GetAttPtr(eip, "AllocationId"),
				Tags:         t.apply(nil, fmt.Sprintf("ngwy-%d", i)),
			}

			tmpl.Resources[subtableassoc] = &ec2.SubnetRouteTableAssociation{
				RouteTableId: cloudformation.Ref("RunnerPublicRouteTable"),
				SubnetId:     cloudformation.Ref(subnet),
			}
		}

		for i, addr := range []string{"128", "129", "130"} {
			subnet := fmt.Sprintf("RunnerPrivateSubnet%d", i)
			subtableassoc := fmt.Sprintf("RunnerPrivateSubnet%dRouteTableAssociation", i)

			tmpl.Resources[subnet] = &ec2.Subnet{
				VpcId:            cloudformation.Ref("RunnerVPC"),
				CidrBlock:        ptr(fmt.Sprintf("10.128.%s.0/24", addr)),
				AvailabilityZone: cloudformation.SelectPtr(i, cloudformation.GetAZs("")),
				Tags: t.apply([]tags.Tag{
					{
						Key:   fmt.Sprintf("kubernetes.io/cluster/%s", intv.InstallID),
						Value: "shared",
					},
					{
						Key:   "kubernetes.io/role/internal-elb",
						Value: "1",
					},
					{
						Key:   "visibility",
						Value: "private",
					},
				}, fmt.Sprintf("private-%d", i)),
			}

			tmpl.Resources[subtableassoc] = &ec2.SubnetRouteTableAssociation{
				RouteTableId: cloudformation.Ref("RunnerPrivateRouteTable"),
				SubnetId:     cloudformation.Ref(subnet),
			}

			inst.NetworkInterfaces = append(inst.NetworkInterfaces, ec2.LaunchTemplate_NetworkInterface{
				AssociatePublicIpAddress: ptr(intv.Environment == "dev"),
				DeviceIndex:              ptr(i),
				SubnetId:                 ptr(cloudformation.Ref(subnet)),
				Groups:                   sgs,
			})

			subnets = append(subnets, cloudformation.Ref(subnet))
		}

		// Just take the first ENI
		// enis = enis[0:1]
		inst.NetworkInterfaces = inst.NetworkInterfaces[0:1]
	}

	if intv.Environment == "dev" {
		inst.KeyName = ptr("hack-keypair")
	}

	tmpl.Resources["RunnerLaunchTemplate"] = &ec2.LaunchTemplate{
		LaunchTemplateName: ptr(cloudformation.Sub("${AWS::StackName}-runner")),
		LaunchTemplateData: inst,
	}

	tmpl.Resources["RunnerASG"] = &autoscaling.AutoScalingGroup{
		VPCZoneIdentifier: subnets,
		LaunchTemplate: &autoscaling.AutoScalingGroup_LaunchTemplateSpecification{
			LaunchTemplateId: cloudformation.RefPtr("RunnerLaunchTemplate"),
			Version:          cloudformation.GetAtt("RunnerLaunchTemplate", "LatestVersionNumber"),
		},
		MaxSize: "1",
		MinSize: "1",
		Tags: []autoscaling.AutoScalingGroup_TagProperty{
			{
				Key:               "nuon_install_id",
				Value:             intv.InstallID,
				PropagateAtLaunch: false, // handled directly
			},
			{
				Key:               "Name",
				Value:             cloudformation.Sub("${AWS::StackName}-runner"),
				PropagateAtLaunch: false,
			},
		},
	}

	tmpl.Resources["RunnerInstanceProfile"] = &iam.InstanceProfile{
		InstanceProfileName: ptr(cloudformation.Sub("${AWS::StackName}--runner-profile")),
		Roles: []string{
			cloudformation.Ref("RunnerInstanceRole"),
		},
	}

	tmpl.Resources["RunnerInstanceRole"] = &iam.Role{
		Description: ptr("Instance role for the runner ec2 instance and ASG. Used to assume Provision, Deprovision, and Maintenance roles as needed by the app."),
		AssumeRolePolicyDocument: map[string]any{
			"Statement": []map[string]any{
				{
					"Effect": "Allow",
					"Principal": map[string]any{
						"Service": "ec2.amazonaws.com",
					},
					"Action": "sts:AssumeRole",
				},
			},
		},
		Policies: []iam.Role_Policy{
			{
				PolicyName: "RunnerInstancePolicy",
				PolicyDocument: map[string]any{
					"Version": "2012-10-17",
					"Statement": []map[string]any{
						{
							"Effect": "Allow",
							"Action": []string{
								"sts:AssumeIdentity",
							},
							"Resource": "*",
						},
					},
				},
			},
			// TODO(sdboyer) remove this after we've gotten all the role assumption working right
			{
				PolicyName: "RunnerInstancePolicyAdmin",
				PolicyDocument: map[string]any{
					"Version": "2012-10-17",
					"Statement": []map[string]any{
						{
							"Effect":   "Allow",
							"Action":   "*",
							"Resource": "*",
						},
					},
				},
			},
		},
		Tags: t.apply(nil, "runner-instance"),
	}

	for _, role := range cfg.Roles {
		roleparams = append(roleparams, role.ParamName())
		paramlabels[role.ParamName()] = map[string]any{
			"default": role.DisplayName,
		}

		tmpl.Conditions[role.StrEnabled()] = cloudformation.Equals(cloudformation.Ref(role.ParamName()), "true")

		tmpl.Resources[role.StrRole()] = &iam.Role{
			AWSCloudFormationCondition: role.StrEnabled(),
			RoleName:                   ptr(cloudformation.Sub(fmt.Sprintf("${AWS::StackName}-sandbox-%s", role.Name))),
			AssumeRolePolicyDocument: map[string]any{
				"Statement": []map[string]any{
					{
						"Effect": "Allow",
						"Principal": map[string]any{
							"AWS": cloudformation.GetAttPtr("RunnerInstanceRole", "Arn"),
						},
						"Action": "sts:AssumeRole",
					},
				},
			},
			Tags: t.apply(nil, fmt.Sprintf("%s-role", role.Name)),
		}

		// TODO(sdboyer) this is NOT what the intent of "builtin" was, rather it was for specifying an AWS managed policy
		for _, policy := range role.Policies {
			if policy.Builtin {
				var pol permissions.Policy

				switch role.KnownRoleName() {
				// TODO this is hardcoded to aws-eks
				case "Provision":
					pol = awseks.ProvisionPolicy
				case "Deprovision":
					pol = awseks.DeprovisionPolicy
				case "BreakGlass", "Maintenance":
					// isn't one for these, shouldn't be possible to hit b/c we should validate earlier
					continue
				default:
					panic(fmt.Sprintf("unknown role name %s", role.Name))
				}
				tmpl.Resources[role.StrPolicy()] = &iam.Policy{
					AWSCloudFormationCondition: role.StrEnabled(),
					PolicyName: cloudformation.SubVars(
						fmt.Sprintf("${AWS::StackName}-sandbox-%s", role.Name),
						map[string]any{"RoleName": cloudformation.Ref(role.StrRole())}),
					PolicyDocument: pol,
					Roles:          []string{cloudformation.Ref(role.StrRole())},
				}
			} else {
				tmpl.Resources[policy.Name] = &iam.Policy{
					AWSCloudFormationCondition: role.StrEnabled(),
					PolicyName: cloudformation.SubVars(
						fmt.Sprintf("${AWS::StackName}-sandbox-%s", role.Name),
						map[string]any{"RoleName": cloudformation.Ref(role.StrRole())}),
					PolicyDocument: json.RawMessage([]byte(policy.PolicyJSON)),
					Roles:          []string{cloudformation.Ref(role.StrRole())},
				}
			}
		}

		tmpl.Parameters[fmt.Sprintf("Enable%s", role.KnownRoleName())] = cloudformation.Parameter{
			Type:    "String",
			Default: role.DefaultParam(),
			AllowedValues: []any{
				"true",
				"false",
			},
			Description: &role.Description,
		}
	}

	// Lambda func to phone home with ARNs for roles, etc.

	lambdaprops := map[string]any{
		"ServiceToken": cloudformation.GetAttPtr("RunnerPhoneHome", "Arn"),
		"RunnerId":     intv.RunnerID,
		"url":          intv.NotifyUrl,
	}
	for _, role := range cfg.Roles {
		lambdaprops[role.StrRole()] = cloudformation.If(role.StrEnabled(), cloudformation.GetAttPtr(role.StrRole(), "Arn"), "")
	}

	tmpl.Resources["PhoneHomeProps"] = &cloudformation.CustomResource{
		Type:       "AWS::CloudFormation::CustomResource",
		Properties: lambdaprops,
	}

	tmpl.Resources["RunnerPhoneHome"] = &lambda.Function{
		Handler:     ptr("phonehome.lambda_handler"),
		Runtime:     ptr("python3.9"),
		Tags:        t.apply(nil, "phone-home-lambda"),
		Description: ptr("Notify the Nuon API of the stack state."),
		Code: &lambda.Function_Code{
			S3Bucket: ptr("nuon-artifacts"),
			S3Key:    ptr("cfngen/runner-phone-home.zip"),
		},
		Role: cloudformation.GetAtt("RunnerPhoneHomeRole", "Arn"),
	}

	tmpl.Resources["RunnerPhoneHomeRole"] = &iam.Role{
		Tags: t.apply(nil, "phone-home-lambda"),
		AssumeRolePolicyDocument: map[string]any{
			"Statement": []map[string]any{
				{
					"Effect": "Allow",
					"Principal": map[string]any{
						"Service": "lambda.amazonaws.com",
					},
					"Action": "sts:AssumeRole",
				},
			},
		},
		Policies: []iam.Role_Policy{
			{
				PolicyName: "CloudwatchPolicy",
				PolicyDocument: map[string]any{
					"Version": "2012-10-17",
					"Statement": []map[string]any{
						{
							"Effect": "Allow",
							"Action": []string{
								"logs:CreateLogGroup",
								"logs:CreateLogStream",
								"logs:PutLogEvents",
							},
							"Resource": "*",
						},
					},
				},
			},
		},
		ManagedPolicyArns: []string{
			"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
		},
	}

	for _, secret := range cfg.Secrets {
		paramlabels[secret.CamelSecret()+"Param"] = map[string]any{
			"default": secret.DisplayName,
		}
		secretparams = append(secretparams, secret.CamelSecret()+"Param")
		tmpl.Parameters[secret.CamelSecret()+"Param"] = cloudformation.Parameter{
			Type:        "String",
			Description: &secret.Description,
			NoEcho:      ptr(true),
		}

		tmpl.Resources[secret.CamelSecret()] = &secretsmanager.Secret{
			Name:        ptr(cloudformation.Sub(fmt.Sprintf("${AWS::StackName}-%s", secret.Name))),
			Description: &secret.Description,
			Tags:        t.apply(nil, ""),
		}
	}

	_ = vpcpgroups

	var pgs []map[string]any
	if cfg.BYOVPC {
		pgs = append(pgs, vpcpgroups)
	}

	pgs = append(pgs, []map[string]any{
		{
			"Label": map[string]any{
				"default": "Application Secrets",
			},
			"Parameters": secretparams,
		},
		{
			"Label": map[string]any{
				"default": "Access Permissions",
			},
			"Parameters": roleparams,
		},
	}...)

	tmpl.Metadata["AWS::CloudFormation::Interface"] = map[string]any{
		"ParameterLabels": paramlabels,
		"ParameterGroups": pgs,
	}

	return tmpl, nil
}

func ptr[T any](v T) *T {
	return &v
}

type tagger struct {
	installID  string
	additional map[string]string
}

func (t tagger) apply(existing []tags.Tag, name string) []tags.Tag {
	existingMap := make(map[string]string)
	for _, tag := range existing {
		existingMap[tag.Key] = tag.Value
	}

	existingMap["nuon_install_id"] = t.installID
	if _, has := existingMap[name]; !has {
		if name != "" {
			existingMap["Name"] = fmt.Sprintf("%s-%s", t.installID, name)
		} else {
			existingMap["Name"] = t.installID
		}
	}
	for k, v := range t.additional {
		if _, has := existingMap[k]; !has {
			existingMap[k] = v
		}
	}

	ret := []tags.Tag{}
	for k, v := range existingMap {
		ret = append(ret, tags.Tag{
			Key:   k,
			Value: v,
		})
	}

	return ret
}
