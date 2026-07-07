# `nuon` Terraform Provider — `nuon_stack` Resource

## Goal

Ship a `nuon` Terraform provider with a `nuon_stack` resource that creates, updates,
and destroys Nuon install stacks by driving the stack SDK (`sdks/stack`) **locally** on
the machine running `terraform apply`.

## Execution model — local-exec (firm)

The resource runs provisioning itself: it fetches the `nuonco/install-stacks` module
(an HTTPS tarball, not `git clone`) and runs the Terraform CLI via the SDK's
`internal/terraform` provisioner, on the apply host.

This is a hard requirement. BYOC customers will not let Nuon's control plane reach into
their cloud, so the customer must run provisioning themselves. The alternative — a thin,
API-backed resource that triggers Nuon's server-side provision workflow (Spacelift-style)
— is a **non-starter** and must not be reconsidered.

### SDK → Terraform CRUD mapping

| TF op   | SDK call                                                        | Returns     |
| ------- | -------------------------------------------------------------- | ----------- |
| Create  | `FromURL(ctx, URLOptions{Kind: KindProvision, ...})` → `Provision()`   | `*Outputs`  |
| Read    | `New(ctx, Options{InstallID, AWSRegion, ...})` → `Status()`           | `*Outputs`  |
| Update  | `FromURL(ctx, URLOptions{Kind: KindReprovision, ...})` → `Reprovision()` | `*Outputs` |
| Delete  | `FromURL(ctx, URLOptions{Kind: KindDeprovision, ...})` → `Deprovision()` | `error`   |
| Import  | parse `install_id:region` → Read                              |             |

`FromURL` calls the ctl-api create-run endpoint and hydrates a `Config` carrying
`install_id`, `cloud`, AWS `region`, log-stream wiring, and the input/secret schema. The
phone-home ID embedded in the URL is the per-stack-version secret; no API token is needed.
It is stable per stack version, so re-running `FromURL` each apply re-resolves the same
config.

## State — remote, in the target account (decided)

The nested Terraform state is stored **remotely in the customer's target cloud account**:
S3 for AWS, GCS for GCP, keyed by install ID
(default `nuon/<install_id>/terraform.tfstate`).

Consequences that make this the clean choice:

- **No state round-trip** through the resource's own TF state.
- **Ephemeral workdir** — a temp dir per operation. Every op re-`init`s against the remote
  backend and Terraform pulls current state, so ephemeral CI runners "just work."
- **Locking for free** — S3 native lockfile (TF ≥ 1.10) or a DynamoDB table; GCS has native
  locking. Concurrent applies are safe.

The customer supplies an **existing** bucket (matches the dashboard `backend.tf` snippets).
Auto-creating the bucket is deferred.

## Schema

```hcl
provider "nuon" {
  api_url = "https://api.nuon.co"   # optional; base up to but excluding /v1
}

resource "nuon_stack" "example" {
  phone_home_id = "<phone-home-id>"   # required — identifies the stack version

  terraform_version = "1.9.5"         # optional — hc-install pin
  module_ref        = "main"          # optional — install-stacks tarball ref/URL override

  # Exactly one cloud block (Max: 1 each), matching the target cloud.
  aws {
    region         = "us-west-2"      # optional — overrides API-resolved region
    account_id     = "..."            # optional — validation only (compared to resolved output)
    state_bucket   = "my-tf-state"    # required — existing S3 bucket in the target account
    state_key      = "nuon/..."       # optional — defaults to nuon/<install_id>/terraform.tfstate
    dynamodb_table = "tf-locks"       # optional — lock table (or rely on S3 native lockfile)
  }

  gcp {
    project_id             = "..."     # required
    region                 = "..."     # required
    state_bucket           = "..."     # required — existing GCS bucket
    state_prefix           = "nuon/..."# optional — defaults to nuon/<install_id>
    runner_machine_type    = "..."     # optional
    has_gke_node_pool      = true      # optional
    gke_node_pool_sa_email = "..."     # optional
  }

  inputs = {                           # optional — customer install-input VALUES
    domain     = "example.com"
    node_count = "3"
  }

  secrets = {                          # optional — sensitive, write-only
    db_password = var.db_password
  }

  # ── computed ──
  install_id = (computed)              # also the resource ID
  outputs    = (computed)              # flattened *Outputs
}
```

### Schema rules

- **Cloud** is inferred from which block is present; validate exactly one, matching the
  run's resolved cloud. No separate `cloud` attribute.
- **Method** is implicitly `terraform`; not exposed.
- **`aws` block is optional** (API resolves region/account) except `state_bucket`, which is
  required for remote state. **`gcp` block is required** for GCP (project/region are not
  known server-side) and also requires `state_bucket`.
- **`api_url` on the provider, `phone_home_id` on the resource.** The URL is built
  internally: `fmt.Sprintf("%s/v1/stack-runs/%s", api_url, phone_home_id)`. `parseURL`
  tolerates a reverse-proxy prefix, so an `api_url` that already includes one is fine.

### Inputs & secrets

The input/secret **schema** (which keys exist, which are required, the auto-generate list,
descriptions) comes from ctl-api in the createRun `Config`. The customer supplies **values**
via `inputs` / `secrets`.

- `inputs` → `map(string)`, overlays `cfg.InstallInputs`.
- `secrets` → `map(string)`, **`Sensitive: true`**, overlays `cfg.Secrets[k].Value`.
- **Error on unknown keys** — any `inputs`/`secrets` key not declared in the app config
  fails at plan time (pre-checked against `PreparedConfig()`; SDK `validateRequiredValues`
  is the apply-time backstop).
- **Secrets are write-only** — config changes trigger reprovision, but Read never
  reconciles or clobbers them; prior state is preserved.
- **Auto-generated secrets** (`cfg.AutoGenerateSecrets`) are never required from the
  customer, and a user-supplied value for one is rejected.
- **Required gating** — `cfg.RequiredInputs` + `SecretInput.Required` drive which keys must
  be present; report all missing at once.
- Secret values live in Terraform state (unavoidable for a declarative resource); mark
  sensitive and document secure state storage.

## SDK-prep milestone (`sdks/stack/`, must land before provider CRUD)

1. **Remote backend on `tf.Init`** — extend `prepare()` / `Options` to accept backend
   config (bucket, key/prefix, region, lock table) and pass via `tfexec.BackendConfig(...)`
   or a written `backend.tf`. `prepare()` currently inits with no backend
   (`provisioner.go:129`).
2. **Ephemeral / injectable workdir** — replace hardcoded `workDirFor`/`binCacheDir`
   (`provisioner.go:153,165`) with a temp or injected dir; state is remote, so no
   persistence needed.
3. **Offline / pre-installed Terraform binary** — honor an explicit exec path /
   `TerraformVersion` pin so hc-install's `releases.hashicorp.com` fetch (`engine.go`) isn't
   mandatory in locked-down environments.
4. **Declarative inputs/secrets overlay** — add `Options.InstallInputs map[string]string`
   and `Options.Secrets map[string]string` (and on `URLOptions`), overlaid onto the hydrated
   `cfg` before `validateRequiredValues`. Replaces the current "mutate `PreparedConfig()` in
   place" TUI-only path.
5. **Confirm `Status`/`Deprovision`** re-init against the same backend and read/destroy
   correctly from remote state (`provisioner.go:66,82`).
6. **`slog` → `tflog` bridge** so `terraform apply` streams SDK progress.

## Provider package

`bins/terraform-provider-nuon/` (mirrors `bins/stack-cli/`), built on
`terraform-plugin-framework` (+ `terraform-plugin-go` added to the root `go.mod`).

```
bins/terraform-provider-nuon/
├── main.go                     # providerserver.Serve → registry.terraform.io/nuonco/nuon
├── internal/provider/
│   ├── provider.go             # api_url config, Resources()
│   ├── stack_resource.go       # Schema + Create/Read/Update/Delete/ImportState
│   ├── stack_resource_model.go # TF model ⇄ stack.Options / *Outputs (flatten/expand)
│   └── stack_resource_test.go  # TF_ACC acceptance tests (gated, like the e2e suite)
├── examples/                   # main.tf usage samples
├── docs/                       # generated by tfplugindocs
├── Dockerfile
└── main_test.go
```

## Operations / credentials / egress (document)

The apply host needs:

- **Egress** to `github.com` (module tarball), `releases.hashicorp.com` (terraform binary,
  unless overridden per prep #3), the Nuon `api_url`, and the target cloud APIs.
- **Ambient cloud credentials** (env / instance profile / ADC) — the provider inherits
  process env; it must not strip it. Those creds need permission for **both** the state
  bucket (S3/GCS read/write + lock) **and** the resources the module provisions.

Document required IAM and airgapped limitations.

## Release

GoReleaser (multi-arch, GPG-signed, `terraform-registry-manifest.json`) → Terraform
Registry, modeled on the stack-cli pipeline. Add `bins/terraform-provider-nuon` to
`ci-triggers.yml`. Generate registry docs with `tfplugindocs`.

## Sequencing

1. **SDK-prep** (#1–#6 above) — backend + ephemeral workdir first.
2. Provider skeleton + `nuon_stack` Create/Read/Delete for **AWS** with remote state.
3. Add Update + import + `gcp` block + `aws.account_id` validation +
   `inputs`/`secrets` + `terraform_version`/`module_ref`.
4. Release pipeline + `tfplugindocs` + examples.
```
