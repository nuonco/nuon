# Permissions Model: Concept Mapping

The permissions/IAM model was originally designed for AWS and the naming reflects that heritage. It now serves AWS, GCP, and Azure. This document maps the AWS-centric names to the cloud-neutral concepts they actually represent.

## Concept Mapping

| Current Name (AWS-centric) | Cloud-Neutral Concept | AWS | GCP | Azure |
|---|---|---|---|---|
| `AppAWSIAMRole` / `AppAWSIAMRoleConfig` | **Permission Identity** | IAM Role | Service Account | Managed Identity |
| `AppAWSIAMPolicy` / `AppAWSIAMPolicyConfig` | **Permission Grant** | Inline or managed IAM policy | Permission list or predefined role | Role assignment |
| `AWSIAMRoleType` | **Role Purpose** | Same across all clouds |

## Policy Fields by Cloud

The `AppAWSIAMPolicy` struct contains fields for multiple clouds in a single type:

| Field | Cloud | Purpose |
|---|---|---|
| `ManagedPolicyName` | AWS | ARN or name of an AWS managed policy (e.g. `AmazonS3FullAccess`) |
| `Contents` | AWS | Inline JSON policy document |
| `GCPPermissions` | GCP | List of individual GCP permission strings |
| `GCPPredefinedRole` | GCP | Google-managed role bundle (e.g. `roles/editor`) — GCP equivalent of AWS managed policies |

## Role Type Constants

| Constant | Value | Meaning |
|---|---|---|
| `AWSIAMRoleTypeRunnerProvision` | `runner_provision` | Used during initial install setup |
| `AWSIAMRoleTypeRunnerDeprovision` | `runner_deprovision` | Used when tearing down an install |
| `AWSIAMRoleTypeRunnerMaintenance` | `runner_maintenance` | Used for updates and day-to-day operations |
| `AWSIAMRoleTypeBreakGlass` | `breakglass` | Elevated access for vendor break-glass scenarios |
| `AWSIAMRoleTypeCustom` | `custom` | Vendor-defined roles for specific app operations |

## Cloud Dispatch

The `CloudPlatform` field on `AppAWSIAMRole` / `AppAWSIAMRoleConfig` disambiguates which cloud the role targets. Downstream consumers branch on this:

- **CloudFormation renderer** (`stacks/cloudformation/`) — only processes AWS roles
- **GCP renderer** (`stacks/gcp/`) — filters on `cloud_platform = "gcp"`
- **Operation role selector** (`operation-roles/selector.go`) — dispatches to `getAWSRoleMap`, `getGCPSAMap`, or `getAzureRoleMap` based on stack outputs

## DB Table Names

The GORM models derive table names from struct names. These are **not safe to rename** without a migration or explicit `TableName()` override:

- `app_aws_iam_role_configs` ← `AppAWSIAMRoleConfig`
- `app_aws_iam_policy_configs` ← `AppAWSIAMPolicyConfig`


