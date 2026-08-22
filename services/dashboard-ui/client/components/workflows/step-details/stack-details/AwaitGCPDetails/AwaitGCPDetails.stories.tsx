export default {
  title: 'Workflows/StepDetails/AwaitGCPDetails',
}

import { AwaitGCPDetails } from './AwaitGCPDetails'

const mockStack = {
  versions: [
    {
      contents: JSON.stringify({
        inputs_tfvars:
          'nuon_install_id = "install-1"\ninstall_inputs = {\n  "cluster_name" = ""\n}\n',
        provider_tfvars:
          'api_url       = "https://api.nuon.co"\nphone_home_id = "ph-1"\n\ngcp = {\n  project_id = "my-proj"\n  region     = "us-central1"\n}\n\ninstall_inputs = {\n  "cluster_name" = ""\n}\n',
        secrets_tfvars:
          'secrets = {\n  "stripe_key" = {\n    description = "Your Stripe API key"\n    required    = true\n    value       = ""\n  }\n}\n\nauto_generate_secrets = ["db_password", ]\n',
        spacelift_admin_tf: `terraform {
  required_providers {
    spacelift = {
      source = "spacelift-io/spacelift"
    }
  }
}

resource "spacelift_stack" "nuon" {
  name              = "nuon-install-1"
  repository        = "install-stacks"
  branch            = "main"
  project_root      = "gcp"
  terraform_version = "1.5.7"
  autodeploy        = true

  raw_git {
    namespace = "nuonco"
    url       = "https://github.com/nuonco/install-stacks.git"
  }
}

resource "spacelift_mounted_file" "inputs" {
  stack_id      = spacelift_stack.nuon.id
  relative_path = "source/gcp/inputs.auto.tfvars"
  write_only    = false
  content       = filebase64("\${path.module}/inputs.auto.tfvars")
}

resource "spacelift_mounted_file" "secrets" {
  stack_id      = spacelift_stack.nuon.id
  relative_path = "source/gcp/secrets.auto.tfvars"
  write_only    = true
  content       = filebase64("\${path.module}/secrets.auto.tfvars")
}
`,
        spacelift_blueprint_yaml: `inputs:
  - id: gcp_project_id
    name: GCP project ID
    type: short_text
    default: "my-gcp-project"
  - id: gcp_region
    name: GCP region
    type: short_text
    default: "us-central1"
  - id: input_cluster_name
    name: cluster_name
    type: short_text
  - id: secret_stripe_key
    name: stripe_key
    type: secret
    description: Your Stripe API key
stack:
  name: nuon-install-1
  space: root
  vcs:
    branch: main
    provider: RAW_GIT
    repository_url: https://github.com/nuonco/install-stacks.git
    project_root: gcp
  vendor:
    terraform:
      manage_state: true
      version: "1.5.7"
  environment:
    mounted_files:
      - path: source/gcp/inputs.auto.tfvars
        secret: false
        content: |
          nuon_install_id = "install-1"
          gcp_project_id = "\${{ inputs.gcp_project_id }}"
          gcp_region = "\${{ inputs.gcp_region }}"
          install_inputs = {
            "cluster_name" = "\${{ inputs.input_cluster_name }}"
          }
      - path: source/gcp/secrets.auto.tfvars
        secret: true
        content: |
          auto_generate_secrets = []
          secrets = {
            "stripe_key" = {
              description = "Your Stripe API key"
              required    = true
              value       = "\${{ inputs.secret_stripe_key }}"
            }
          }
`,
      }),
    },
  ],
} as any

const mockLegacyStack = {
  versions: [
    {
      contents: JSON.stringify({
        tfvars: `nuon_install_id          = "inla1cquvcch1xh6lww38ikf6k"
nuon_org_id              = "org3d3kug2cr7kgf508st60ozt"
nuon_app_id              = "appnxg0vgp77ar1lkfeq1c1fka"
install_inputs = {
}
auto_generate_secrets = []
secrets = {
  "stripe_key" = {
    description = "Your Stripe API key"
    required    = true
    value       = ""
  }
}
`,
      }),
    },
  ],
} as any

const mockStep = {
  id: 'step-1',
  status: { status: 'active' },
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetails
      stack={mockStack}
      step={mockStep}
      installId="install-1"
      gcpProjectId="my-gcp-project"
      spaceliftEnabled
    />
  </div>
)

export const SpaceliftDisabled = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetails stack={mockStack} step={mockStep} installId="install-1" />
  </div>
)

export const LegacyContents = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetails
      stack={mockLegacyStack}
      step={mockStep}
      installId="install-1"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetails
      stack={mockStack}
      step={mockStep}
      installId="install-1"
      loading
    />
  </div>
)
