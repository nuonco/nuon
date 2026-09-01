# Custom nested stack support for Terraform stack modules

## Problem

The terraform modules do not support custom nested stacks.
In order to make adoption as seamless as possible, we want the modules to be fully backwards-compatible with existing app configs.
This requires adding support for custom nested stacks, for all 3 platforms.

- `aws-cloudformation` — CFN nested stacks
  (`services/ctl-api/internal/pkg/stacks/cloudformation/nested_template_custom.go`)
- `azure-bicep` — ARM linked deployments
  (`services/ctl-api/internal/pkg/stacks/arm/linked_deployment_custom.go`, payload emission in
  `arm/resource_phone_home.go`)
- `gcp-terraform` — curated Terraform modules from github.com/nuonco/install-stacks
  (`services/ctl-api/internal/pkg/stacks/gcp/render.go`)

## Approach

### AWS

When creating a new stack version for an AWS app -- alongside the current, stand-alone CloudFormation stack template -- create a custom-stacks-only template.
This new template will exclude the VPC, Runner, Key Vault, and phone-home resources.
It only provision the nested stacks.
This allows dependencies between stacks to resolve within CloudFormation at deploy-time, exactly as they do today.

### Azure

TBD

### GCP

TBD
