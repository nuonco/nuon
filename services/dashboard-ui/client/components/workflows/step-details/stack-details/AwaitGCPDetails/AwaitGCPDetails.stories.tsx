export default {
  title: 'Workflows/StepDetails/AwaitGCPDetails',
}

import { AwaitGCPDetails, AwaitGCPDetailsSkeleton } from './AwaitGCPDetails'

const mockStack = {
  versions: [
    {
      contents: JSON.stringify({
        inputs_tfvars:
          'nuon_install_id = "install-1"\ninstall_inputs = {\n  "cluster_name" = ""\n}\n',
        secrets_tfvars:
          'auto_generate_secrets = ["db_password", ]\nsecrets = {\n  "stripe_key" = {\n    description = "Your Stripe API key"\n    required    = true\n    value       = ""\n  }\n}\n',
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
  terraform_version = "1.9.0"
  autodeploy        = true
}

resource "spacelift_mounted_file" "inputs" {
  stack_id      = spacelift_stack.nuon.id
  relative_path = "source/gcp/inputs.auto.tfvars"
  write_only    = false
  content       = "bnVvbl9pbnN0YWxsX2lkID0gImluc3RhbGwtMSIK"
}

resource "spacelift_mounted_file" "secrets" {
  stack_id      = spacelift_stack.nuon.id
  relative_path = "source/gcp/secrets.auto.tfvars"
  write_only    = true
  content       = "YXV0b19nZW5lcmF0ZV9zZWNyZXRzID0gW10K"
}
`,
        spacelift_blueprint_yaml: `inputs: []
stack:
  name: nuon-install-1
  space: root
  vcs:
    branch: main
    repository: install-stacks
    project_root: gcp
    provider: GITHUB
  vendor:
    terraform:
      manage_state: true
      version: "1.9.0"
  environment:
    mounted_files:
      - path: source/gcp/inputs.auto.tfvars
        write_only: false
        content: bnVvbl9pbnN0YWxsX2lkID0gImluc3RhbGwtMSIK
      - path: source/gcp/secrets.auto.tfvars
        write_only: true
        content: YXV0b19nZW5lcmF0ZV9zZWNyZXRzID0gW10K
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
    <AwaitGCPDetails stack={mockStack} step={mockStep} installId="install-1" />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-4">
    <AwaitGCPDetailsSkeleton />
  </div>
)
