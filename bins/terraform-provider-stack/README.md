# terraform-provider-stack

The `stack` Terraform provider. It lets an install-stacks Terraform module read
its Nuon-rendered configuration from the control plane instead of receiving it
as generated tfvars.

The provider exposes a single data source:

- **`stack_config` data source** — read-only fetch of a stack's rendered config
  (runner details, permissions, roles, install inputs, secrets) keyed by
  `phone_home_id`. Intended for use *inside* an install-stacks module (e.g.
  `nuonco/install-stacks//gcp`) so it reads config from the API rather than
  receiving it as generated tfvars. Provisions nothing.

It calls the stack SDK's read-only `FetchConfig` (`sdks/stack`), which hits the
public, side-effect-free `GET /v1/stack-runs/{phone_home_id}/config` endpoint —
the per-stack-version `phone_home_id` in the URL is the secret.

## Layout

```
main.go                          provider entry point (providerserver.Serve)
internal/provider/
  provider.go                    provider schema + api_url; registers the data source
  stack_data_source.go           stack_config data source: schema + read
  stack_data_source_model.go     data source model + config flattener
  *_test.go                      schema validation + flatten unit tests
examples/
  data-source-gcp/main.tf        stack_config data source example (GCP)
docs/
  data-source.html               architecture/walkthrough for the data source
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

See `docs/data-source.html` for the architecture diagram, schema tables, and a
step-by-step walkthrough.
