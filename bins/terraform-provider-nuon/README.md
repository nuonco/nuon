# terraform-provider-nuon

The `nuon` Terraform provider. It integrates Terraform with Nuon install stacks
using the stack SDK (`sdks/stack`) — provisioning stacks locally (BYOC) and
reading their configuration from the Nuon control plane.

The provider exposes one resource and one data source:

- **`nuon_stack` resource** — provisions/updates/destroys an install stack from
  Terraform. On `apply` it drives the stack SDK to fetch config from the control
  plane, run the `nuonco/install-stacks` Terraform module against the customer's
  cloud account, and stream logs back to Nuon. State lives in a remote backend in
  the target account (S3/GCS).
- **`nuon_stack` data source** — read-only fetch of a stack's rendered config
  (runner details, permissions, roles, install inputs) keyed by `phone_home_id`.
  Intended for use *inside* an install-stacks module so it reads config from the
  API instead of receiving it as generated tfvars. Provisions nothing.

## Layout

```
main.go                          provider entry point (providerserver.Serve)
internal/provider/
  provider.go                    provider schema + api_url; registers resource & data source
  stack_resource.go              nuon_stack resource: schema + CRUD/import
  stack_resource_model.go        resource state model + outputs flattener
  stack_data_source.go           nuon_stack data source: schema + read
  stack_data_source_model.go     data source model + config flattener
  *_test.go                      schema validation + flatten unit tests
examples/
  main.tf                        nuon_stack resource example
  data-source-gcp/main.tf        nuon_stack data source example (GCP)
docs/
  resource.html                  architecture/walkthrough for the resource
  data-source.html               architecture/walkthrough for the data source
```

## Provider configuration

```hcl
provider "nuon" {
  api_url = "https://api.nuon.co" # optional; base URL up to but excluding /v1
}
```

The stack-run endpoints live on Nuon's runner API surface. In production
`api_url` is `https://api.nuon.co`; for local development point it at the local
runner API (`http://localhost:8083`).

## Development

Build and install for local testing:

```bash
go build -o "$(go env GOPATH)/bin/terraform-provider-nuon" .
```

Point Terraform at the local build with a dev override (`~/.terraformrc` or a
file referenced by `TF_CLI_CONFIG_FILE`):

```hcl
provider_installation {
  dev_overrides { "nuonco/nuon" = "/Users/you/go/bin" }
  direct {}
}
```

With a dev override set, skip `terraform init` — run `terraform plan`/`apply`
directly.

Run the tests:

```bash
go test ./...
```

See `docs/resource.html` and `docs/data-source.html` for architecture diagrams,
schema tables, and step-by-step walkthroughs.
