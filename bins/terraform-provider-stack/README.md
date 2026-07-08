# terraform-provider-stack

The `stack` Terraform provider. It lets an install-stacks Terraform module read
its Nuon-rendered configuration from the control plane instead of receiving it
as generated tfvars.

The provider exposes two surfaces:

- **`stack_config` data source** — read-only fetch of a stack's rendered config
  (runner details, permissions, roles, install inputs, secrets) keyed by
  `phone_home_id`. Intended for use *inside* an install-stacks module (e.g.
  `nuonco/install-stacks//gcp`) so it reads config from the API rather than
  receiving it as generated tfvars. Provisions nothing.
- **`stack_phone_home` resource** — reports the result of a run back to the
  control plane, so the module reports run status through the provider instead
  of building the phone-home HTTP request itself. The resource lifecycle drives
  the phone-home `request_type` (Create/Update/Delete); the reported outputs are
  passed as an opaque `jsonencode({...})` payload.

The data source calls the stack SDK's read-only `FetchConfig` (`sdks/stack`),
which hits the public, side-effect-free `GET /v1/stack-runs/{phone_home_id}/config`
endpoint. The resource calls the SDK's `PhoneHome`, which POSTs to the public
`/v1/installs/{install_id}/phone-home/{phone_home_id}` endpoint. In both cases
the per-stack-version `phone_home_id` in the URL is the secret.

## Layout

```
main.go                          provider entry point (providerserver.Serve)
internal/provider/
  provider.go                    provider schema + api_url; registers the data source + resource
  stack_data_source.go           stack_config data source: schema + read
  stack_data_source_model.go     data source model + config flattener
  phone_home_resource.go         stack_phone_home resource: schema + lifecycle
  phone_home_resource_model.go   resource model
  *_test.go                      schema validation + flatten unit tests
examples/
  data-source-gcp/main.tf        stack_config + stack_phone_home example (GCP)
docs/
  data-source.html               architecture/walkthrough for the data source + resource
```

## Provider configuration

```hcl
provider "stack" {
  api_url = "https://api.nuon.co" # optional; base URL up to but excluding /v1
}
```

The config endpoint lives on Nuon's runner API surface. In production `api_url`
is `https://api.nuon.co`; for local development point it at the local runner API
(`http://localhost:8083`).

## Development

Build and install for local testing:

```bash
go build -o "$(go env GOPATH)/bin/terraform-provider-stack" .
```

Point Terraform at the local build with a dev override (`~/.terraformrc` or a
file referenced by `TF_CLI_CONFIG_FILE`):

```hcl
provider_installation {
  dev_overrides { "nuonco/stack" = "/Users/you/go/bin" }
  direct {}
}
```

With a dev override set, skip `terraform init` — run `terraform plan`/`apply`
directly.

Run the tests:

```bash
go test ./...
```

See `docs/data-source.html` for the architecture diagrams, schema tables, and
step-by-step walkthroughs of both the data source and the resource.
