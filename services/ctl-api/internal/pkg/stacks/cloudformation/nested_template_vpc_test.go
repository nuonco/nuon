package cloudformation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const mockVPCTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Deploys a VPC w/ 1 to 3 public subnets (one per az), 1 to 3 private subnets (one per az), a Nat gateway, and an internet gateway.
Parameters:
  ClusterName:
    Description: The name for the EKS Cluster that will be deployed on this VPC.
    Type: String
  NuonInstallID:
    Description: The Nuon Install ID; prefixed to resource names.
    Type: String
  NuonOrgID:
    Description: The Nuon Org ID. Used in tags.
    Type: String
  NuonAppID:
    Description: The Nuon Install ID. Used in tags.
    Type: String
  VpcCIDR:
    Description: Please enter the IP range (CIDR notation) for this VPC.
    Type: String
    Default: 10.128.0.0/16
  PublicSubnet1CIDR:
    Description: Please enter the IP range (CIDR notation) for the public subnet in the first Availability Zone
    Type: String
    Default: 10.128.0.0/26
  PublicSubnet2CIDR:
    Description: Please enter the IP range (CIDR notation) for the public subnet in the second Availability Zone
    Type: String
    Default: 10.128.0.64/26
  PublicSubnet3CIDR:
    Description: Please enter the IP range (CIDR notation) for the public subnet in the third Availability Zone
    Type: String
    Default: 10.128.0.128/26
  RunnerSubnetCIDR:
    Description: Please enter the IP range (CIDR notation) for the dedicated private subnet for the runner.
    Type: String
    Default: 10.128.128.0/24
  PrivateSubnet1CIDR:
    Description: Please enter the IP range (CIDR notation) for the private subnet in the first Availability Zone
    Type: String
    Default: 10.128.130.0/24
  PrivateSubnet2CIDR:
    Description: Please enter the IP range (CIDR notation) for the private subnet in the second Availability Zone
    Type: String
    Default: 10.128.132.0/24
  PrivateSubnet3CIDR:
    Description: Please enter the IP range (CIDR notation) for the private subnet in the third Availability Zone
    Type: String
    Default: 10.128.134.0/24
Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: !Ref VpcCIDR
`

func TestExtractNestedStackParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(mockVPCTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	params, defaultParams, reservedInTemplate, _, err := tpl.extractNestedStackParameters(server.URL + "/stack.yaml")
	require.NoError(t, err)

	expectedParams := []string{
		"VpcCIDR",
		"PublicSubnet1CIDR",
		"PublicSubnet2CIDR",
		"PublicSubnet3CIDR",
		"RunnerSubnetCIDR",
		"PrivateSubnet1CIDR",
		"PrivateSubnet2CIDR",
		"PrivateSubnet3CIDR",
	}

	reservedParams := []string{"ClusterName", "NuonInstallID", "NuonOrgID", "NuonAppID"}

	for _, paramName := range expectedParams {
		t.Run("params_contains_"+paramName, func(t *testing.T) {
			_, ok := params[paramName]
			assert.True(t, ok, "params should contain %s", paramName)
		})
		t.Run("defaultParams_contains_"+paramName, func(t *testing.T) {
			_, ok := defaultParams[paramName]
			assert.True(t, ok, "defaultParams should contain %s", paramName)
		})
	}

	for _, paramName := range reservedParams {
		t.Run("params_excludes_reserved_"+paramName, func(t *testing.T) {
			_, ok := params[paramName]
			assert.False(t, ok, "params should not contain reserved param %s", paramName)
		})
		t.Run("defaultParams_excludes_reserved_"+paramName, func(t *testing.T) {
			_, ok := defaultParams[paramName]
			assert.False(t, ok, "defaultParams should not contain reserved param %s", paramName)
		})
	}

	// verify reserved params are tracked
	assert.True(t, reservedInTemplate["ClusterName"], "reservedInTemplate should contain ClusterName")
	assert.True(t, reservedInTemplate["NuonInstallID"], "reservedInTemplate should contain NuonInstallID")
	assert.True(t, reservedInTemplate["NuonOrgID"], "reservedInTemplate should contain NuonOrgID")
	assert.True(t, reservedInTemplate["NuonAppID"], "reservedInTemplate should contain NuonAppID")

	require.Contains(t, defaultParams, "VpcCIDR")
	assert.Equal(t, "String", defaultParams["VpcCIDR"].Type)
	assert.Equal(t, "10.128.0.0/16", defaultParams["VpcCIDR"].Default)
	require.NotNil(t, defaultParams["VpcCIDR"].Description)
	assert.Equal(t, "Please enter the IP range (CIDR notation) for this VPC.", *defaultParams["VpcCIDR"].Description)
}

const mockRunnerASGTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Runner ASG template
Parameters:
  SubnetId:
    Description: The subnet on which the app will run within the selected VPC.
    Type: AWS::EC2::Subnet::Id
  RunnerEgressGroupId:
    Description: The security group for the runner instance that allows outbound traffic.
    Type: AWS::EC2::SecurityGroup::Id
  InstallId:
    Type: String
    Description: The install ID
  RunnerId:
    Type: String
    Description: The runner ID
  RunnerApiToken:
    Type: String
    Description: API token for the runner
  RunnerApiUrl:
    Type: String
    Description: API URL for the runner
    Default: https://runner.nuon.co
  RunnerInitScriptUrl:
    Type: String
    Description: URL for the init script that is added to the use data for the Runner ASG VM instances.
    Default: https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh
  InstanceType:
    Type: String
    Description: EC2 instance type for the runner
    Default: t3a.medium
  RootVolumeSize:
    Type: Number
    Description: Size of the root EBS volume in GB
    Default: 30
Resources:
  RunnerAutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      MaxSize: "1"
      MinSize: "1"
`

func TestExtractNestedStackParameters_RunnerASG(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(mockRunnerASGTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	params, defaultParams, reservedInTemplate, _, err := tpl.extractNestedStackParameters(server.URL + "/stack.yaml")
	require.NoError(t, err)

	assert.Empty(t, reservedInTemplate, "runner ASG template should have no reserved params")

	expectedParams := []string{
		"SubnetId",
		"RunnerEgressGroupId",
		"InstallId",
		"RunnerId",
		"RunnerApiToken",
		"RunnerApiUrl",
		"RunnerInitScriptUrl",
		"InstanceType",
		"RootVolumeSize",
	}

	for _, paramName := range expectedParams {
		t.Run("params_contains_"+paramName, func(t *testing.T) {
			_, ok := params[paramName]
			assert.True(t, ok, "params should contain %s", paramName)
		})
		t.Run("defaultParams_contains_"+paramName, func(t *testing.T) {
			_, ok := defaultParams[paramName]
			assert.True(t, ok, "defaultParams should contain %s", paramName)
		})
	}

	require.Contains(t, defaultParams, "InstanceType")
	assert.Equal(t, "String", defaultParams["InstanceType"].Type)
	assert.Equal(t, "t3a.medium", defaultParams["InstanceType"].Default)

	require.Contains(t, defaultParams, "RootVolumeSize")
	assert.Equal(t, "Number", defaultParams["RootVolumeSize"].Type)
	assert.Equal(t, 30, defaultParams["RootVolumeSize"].Default)

	require.Contains(t, defaultParams, "SubnetId")
	assert.Equal(t, "AWS::EC2::Subnet::Id", defaultParams["SubnetId"].Type)
	assert.Nil(t, defaultParams["SubnetId"].Default)
}

const mockBYOVPCTemplateYAML = `
AWSTemplateFormatVersion: '2010-09-09'
Description: Pass-through CloudFormation stack for BYO-VPC. Accepts existing VPC and subnet IDs, validates they belong together, and outputs them for use by Nuon.
Parameters:
  VpcID:
    Description: The ID of the existing VPC (e.g., vpc-xxxxxxxxx).
    Type: AWS::EC2::VPC::Id
  PublicSubnetIDs:
    Description: Comma-separated list of existing public subnet IDs.
    Type: String
  PrivateSubnetIDs:
    Description: Comma-separated list of existing private subnet IDs.
    Type: String
  RunnerSubnetID:
    Description: Subnet ID for the Nuon runner. Must have outbound internet access (e.g., via NAT Gateway).
    Type: AWS::EC2::Subnet::Id
  NuonInstallID:
    Description: The Nuon Install ID.
    Type: String
  NuonOrgID:
    Description: The Nuon Org ID. Used in tags.
    Type: String
  NuonAppID:
    Description: The Nuon App ID. Used in tags.
    Type: String
Resources:
  SubnetValidatorRole:
    Type: AWS::IAM::Role
`

func TestExtractNestedStackParameters_BYOVPC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(mockBYOVPCTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	params, defaultParams, reservedInTemplate, _, err := tpl.extractNestedStackParameters(server.URL + "/stack.yaml")
	require.NoError(t, err)

	// BYOVPC template has NuonInstallID, NuonOrgID, NuonAppID but no ClusterName
	assert.False(t, reservedInTemplate["ClusterName"], "BYOVPC template should not have ClusterName")
	assert.True(t, reservedInTemplate["NuonInstallID"], "BYOVPC template should have NuonInstallID")

	expectedParams := []string{
		"VpcID",
		"PublicSubnetIDs",
		"PrivateSubnetIDs",
		"RunnerSubnetID",
	}

	reservedParams := []string{"ClusterName", "NuonInstallID", "NuonOrgID", "NuonAppID"}

	for _, paramName := range expectedParams {
		t.Run("params_contains_"+paramName, func(t *testing.T) {
			_, ok := params[paramName]
			assert.True(t, ok, "params should contain %s", paramName)
		})
		t.Run("defaultParams_contains_"+paramName, func(t *testing.T) {
			_, ok := defaultParams[paramName]
			assert.True(t, ok, "defaultParams should contain %s", paramName)
		})
	}

	for _, paramName := range reservedParams {
		t.Run("params_excludes_reserved_"+paramName, func(t *testing.T) {
			_, ok := params[paramName]
			assert.False(t, ok, "params should not contain reserved param %s", paramName)
		})
		t.Run("defaultParams_excludes_reserved_"+paramName, func(t *testing.T) {
			_, ok := defaultParams[paramName]
			assert.False(t, ok, "defaultParams should not contain reserved param %s", paramName)
		})
	}

	// Verify VpcID parameter details
	require.Contains(t, defaultParams, "VpcID")
	assert.Equal(t, "AWS::EC2::VPC::Id", defaultParams["VpcID"].Type)
	assert.Nil(t, defaultParams["VpcID"].Default)
	require.NotNil(t, defaultParams["VpcID"].Description)
	assert.Equal(t, "The ID of the existing VPC (e.g., vpc-xxxxxxxxx).", *defaultParams["VpcID"].Description)

	// Verify RunnerSubnetID parameter details
	require.Contains(t, defaultParams, "RunnerSubnetID")
	assert.Equal(t, "AWS::EC2::Subnet::Id", defaultParams["RunnerSubnetID"].Type)
	assert.Nil(t, defaultParams["RunnerSubnetID"].Default)

	// Verify String type parameters
	require.Contains(t, defaultParams, "PublicSubnetIDs")
	assert.Equal(t, "String", defaultParams["PublicSubnetIDs"].Type)

	require.Contains(t, defaultParams, "PrivateSubnetIDs")
	assert.Equal(t, "String", defaultParams["PrivateSubnetIDs"].Type)
}

func TestGetVPCNestedStack_OmitsClusterNameWhenNotInTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(mockBYOVPCTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL: server.URL + "/stack.yaml",
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	stack, _, err := tpl.getVPCNestedStack(inp, tb)
	require.NoError(t, err)

	assert.NotContains(t, stack.Parameters, "ClusterName", "ClusterName should not be in stack parameters when template does not define it")
	assert.Equal(t, inp.Install.ID, stack.Parameters["NuonInstallID"])
	assert.Equal(t, inp.Install.AppID, stack.Parameters["NuonAppID"])
	assert.Equal(t, inp.Install.OrgID, stack.Parameters["NuonOrgID"])
}

func TestGetVPCNestedStack_IncludesClusterNameWhenInTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write([]byte(mockVPCTemplateYAML))
	}))
	defer server.Close()

	tpl := &Templates{}
	inp := &stacks.TemplateInput{
		Install: &app.Install{
			ID:    "test-install-id",
			AppID: "test-app-id",
			OrgID: "test-org-id",
		},
		AppCfg: &app.AppConfig{
			StackConfig: app.AppStackConfig{
				VPCNestedTemplateURL: server.URL + "/stack.yaml",
			},
		},
	}
	tb := tagBuilder{installID: inp.Install.ID}

	stack, _, err := tpl.getVPCNestedStack(inp, tb)
	require.NoError(t, err)

	assert.Contains(t, stack.Parameters, "ClusterName", "ClusterName should be in stack parameters when template defines it")
	assert.Equal(t, inp.Install.ID, stack.Parameters["ClusterName"])
	assert.Equal(t, inp.Install.ID, stack.Parameters["NuonInstallID"])
}

func TestExtractNestedStackParameters_InvalidURL(t *testing.T) {
	tpl := &Templates{}
	_, _, _, _, err := tpl.extractNestedStackParameters("http://invalid.localhost.test/template.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch template")
}

func TestExtractNestedStackParameters_InvalidTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not valid yaml: [[["))
	}))
	defer server.Close()

	tpl := &Templates{}
	_, _, _, _, err := tpl.extractNestedStackParameters(server.URL + "/stack.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch template")
}
