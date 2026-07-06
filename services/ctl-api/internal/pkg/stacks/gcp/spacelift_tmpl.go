package gcp

// spaceliftAdminTfTmpl renders an administrative-stack Terraform config that uses
// the spacelift-io/spacelift provider to create a stack running the public
// install-stacks//gcp module. The inputs/secrets tfvars are read from sibling
// files (delivered alongside this config) rather than embedded, so the customer
// can edit inputs.auto.tfvars and replace secrets.auto.tfvars before applying.
// The secrets file is write-only so its contents aren't exposed after apply.
const spaceliftAdminTfTmpl = `terraform {
  required_providers {
    spacelift = {
      source = "spacelift-io/spacelift"
    }
  }
}

variable "space_id" {
  description = "ID of the Spacelift space to create this stack in. Find it in the Spacelift dashboard under Spaces (the URL slug, e.g. \"legacy\" or a generated ID), or query the spacelift_spaces data source."
  type        = string
}

resource "spacelift_stack" "nuon" {
  name              = "nuon-{{.InstallID}}"
  description       = "Nuon runner install stack for {{.InstallID}}"
  space_id          = var.space_id
  repository        = "install-stacks"
  branch            = "main"
  project_root      = "gcp"
  terraform_version = "{{.TerraformVersion}}"
  autodeploy        = true

  raw_git {
    namespace = "nuonco"
    url       = "https://github.com/nuonco/install-stacks.git"
  }
}

# Attaches Spacelift's native GCP integration; grant the printed service account an IAM role on your GCP project.
resource "spacelift_gcp_service_account" "nuon" {
  stack_id     = spacelift_stack.nuon.id
  token_scopes = ["https://www.googleapis.com/auth/cloud-platform"]
}

resource "spacelift_mounted_file" "inputs" {
  stack_id      = spacelift_stack.nuon.id
  relative_path = "source/gcp/inputs.auto.tfvars"
  write_only    = false
  content       = filebase64("${path.module}/inputs.auto.tfvars")
}

resource "spacelift_mounted_file" "secrets" {
  stack_id      = spacelift_stack.nuon.id
  relative_path = "source/gcp/secrets.auto.tfvars"
  write_only    = true
  content       = filebase64("${path.module}/secrets.auto.tfvars")
}

output "gcp_service_account_email" {
  description = "Grant this service account an IAM role on your target GCP project before triggering the stack's first run."
  value       = spacelift_gcp_service_account.nuon.service_account_email
}
`

// spaceliftBlueprintTmpl renders a Spacelift blueprint that provisions a stack
// running the public install-stacks//gcp module. Customer install inputs and
// secrets are exposed as blueprint inputs and interpolated into the mounted
// tfvars via CEL (`${{ inputs.<id> }}`); the mounted-file content is plaintext,
// which is what blueprints expect (unlike the provider's spacelift_mounted_file).
const spaceliftBlueprintTmpl = `{{- if .Inputs}}
inputs:
{{- range .Inputs}}
  - id: {{.ID}}
    name: {{.Name}}
    type: {{.Type}}
{{- if .Description}}
    description: {{.Description}}
{{- end}}
{{- if .Default}}
    default: "{{.Default}}"
{{- end}}
{{- end}}
{{- else}}
inputs: []
{{- end}}
stack:
  name: nuon-{{.InstallID}}
  description: Nuon runner install stack for {{.InstallID}}
  space: root
  vcs:
    branch: main
    provider: RAW_GIT
    repository_url: https://github.com/nuonco/install-stacks.git
    project_root: gcp
  vendor:
    terraform:
      manage_state: true
      version: "{{.TerraformVersion}}"
  environment:
    mounted_files:
      - path: source/gcp/inputs.auto.tfvars
        secret: false
        content: |
{{.InputsTfvars}}
      - path: source/gcp/secrets.auto.tfvars
        secret: true
        content: |
{{.SecretsTfvars}}
`
