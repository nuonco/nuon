package gcp

// spaceliftAdminTfTmpl renders an administrative-stack Terraform config that uses
// the spacelift-io/spacelift provider to create a stack running the public
// install-stacks//gcp module. The generated inputs/secrets tfvars are mounted
// into the stack workspace as base64-encoded files so no values need to live in
// VCS. The secrets file is write-only so its contents aren't exposed after apply.
const spaceliftAdminTfTmpl = `terraform {
  required_providers {
    spacelift = {
      source = "spacelift-io/spacelift"
    }
  }
}

resource "spacelift_stack" "nuon" {
  name              = "nuon-{{.InstallID}}"
  description       = "Nuon runner install stack for {{.InstallID}}"
  repository        = "install-stacks"
  branch            = "main"
  project_root      = "gcp"
  terraform_version = "{{.TerraformVersion}}"
  autodeploy        = true
}

resource "spacelift_mounted_file" "inputs" {
  stack_id      = spacelift_stack.nuon.id
  relative_path = "source/gcp/inputs.auto.tfvars"
  write_only    = false
  content       = "{{.InputsB64}}"
}

resource "spacelift_mounted_file" "secrets" {
  stack_id      = spacelift_stack.nuon.id
  relative_path = "source/gcp/secrets.auto.tfvars"
  write_only    = true
  content       = "{{.SecretsB64}}"
}
`

// spaceliftBlueprintTmpl renders a Spacelift blueprint that provisions a stack
// running the public install-stacks//gcp module, mounting the generated
// inputs/secrets tfvars as base64-encoded files.
const spaceliftBlueprintTmpl = `inputs: []
stack:
  name: nuon-{{.InstallID}}
  description: Nuon runner install stack for {{.InstallID}}
  space: root
  vcs:
    branch: main
    repository: install-stacks
    project_root: gcp
    provider: GITHUB
  vendor:
    terraform:
      manage_state: true
      version: "{{.TerraformVersion}}"
  environment:
    mounted_files:
      - path: source/gcp/inputs.auto.tfvars
        write_only: false
        content: {{.InputsB64}}
      - path: source/gcp/secrets.auto.tfvars
        write_only: true
        content: {{.SecretsB64}}
`
