# Telemetry agent

A standalone edge telemetry agent for one Nuon install. It supervises an
OpenTelemetry Collector that receives OTLP logs, metrics, and traces from
existing application or infrastructure collectors and forwards them to the
install's configured telemetry relay. It does not scrape Kubernetes, read node
log files, or require Kubernetes API permissions.

The controller uses an install-scoped opaque Nuon API token to poll the runner
API's `GET /v1/installs/{install_id}/telemetry/collector-settings` every 15 seconds,
including while telemetry is disabled. When enabled, it exchanges that credential
for a short-lived relay JWT at
`POST /v1/installs/{install_id}/telemetry/access-token?relay_endpoint=...`.
JWTs are renewed with jitter at 60–70% of their lifetime and stored atomically in
a protected file. The bootstrap credential is never sent to the relay.

## Bootstrap and deployment

Use the **public API** as an org administrator to prepare the identity:

1. `POST /v1/installs/{install_id}/telemetry/collector` with no body creates or
   reconciles the managed account. Read `account_id` from the response. This
   operation neither enables telemetry nor issues a credential.
2. `POST /v1/service-accounts/{account_id}/tokens` with
   `{"name":"telemetry agent","duration":"8760h","invalidate":false}`
   creates an opaque credential. Save the response's `token` to a protected file;
   do not put it in source control, Helm values, or command-line arguments.
3. Create the deployment namespace if needed, then create a Secret from that
   file. Remove the local credential file when it is no longer needed:

   ```sh
   kubectl -n telemetry create secret generic nuon-telemetry-agent-bootstrap \
     --from-file=token=/protected/path/bootstrap-token
   ```

The managed account's permissions cover only that install's telemetry resource.
Do not substitute an org-wide runner or administrator token.

Build the image from the repository root and publish it to a registry accessible
by the customer cluster:

```sh
docker build -f bins/telemetry-agent/Dockerfile \
  -t registry.example.com/nuon-telemetry-agent:0.1.0 .
```

Deploy using the [telemetry-agent Helm chart](https://github.com/nuonco/charts/tree/main/charts/telemetry-agent).
Its documentation covers required values, Secret mounting, producer endpoints,
and persistent storage. `apiURL` is the **runner API**, not the public API used
for bootstrap. Enable forwarding through the install's Nuon telemetry settings;
creating the account or deploying the chart does not enable it.

**Ingress is unauthenticated cluster-local OTLP.** Restrict it with cluster
network policy to trusted producers for this install. Do not expose the Service
publicly or use this gateway to aggregate unrelated installs. Outbound relay
connections always require HTTPS and certificate verification. Restrict health
port 13133 to cluster operations as well.

## Settings, credentials, and durability

- Disablement stops the OTLP child process and JWT renewal, removes the JWT
  directory, and continues settings polling. `/livez` and `/readyz` remain healthy
  for an intentionally disabled controller. Producers should retry during
  disabled periods; the gateway does not accept or buffer new input then.
- Successful name or label updates restart the child and may briefly interrupt
  ingestion. The current `nuon.org.name`, `nuon.app.name`, `nuon.install.name`, and
  `nuon.install.labels.*` resource attributes are applied without modifying other
  producer attributes. Labels must not contain secrets. Queued records keep
  their original metadata.
- Failed settings fetches retain the active configuration. Metadata-only start
  failures can restore the previous configuration. Destination changes stop the
  old child and renewal before requesting credentials for the new destination;
  a failed switch never resumes the old destination.
- Logs, metrics, and traces use fsynced persistent exporter queues, bounded to
  1 GiB per signal, with retries for transient failures. Queue overflow rejects
  input, so producers must honor OTLP errors and retry. Permanent relay rejection
  can discard records; delivery is not exactly-once. Queued records resume with
  the current destination after disablement or endpoint changes.
- The default `emptyDir` survives container/child restarts, **not Pod replacement**.
  Set the chart's `persistence.existingClaim` to a dedicated writable PVC for durability
  across Pod replacement. Budget space for all three queues, storage overhead,
  and compaction; queue byte limits do not cap filesystem usage. The chart uses
  one replica and `Recreate` to avoid concurrent access to queue files. Protect
  the volume as sensitive telemetry storage.
- Rotate the bootstrap token before expiry by issuing another account token,
  updating the existing Secret, and then revoking the old token after the
  projected update is visible. `invalidate:true` revokes existing account tokens
  before issuing the replacement and can cause a temporary interruption.
  Each authenticated request reopens the file, so projected updates are picked
  up without restarting. Invalid or missing files fail closed. Revocation blocks
  new JWT issuance; already-issued relay JWTs remain valid until expiry.

`/readyz` returns 503 before settings load or when an enabled child cannot start;
`/livez` tracks the controller, not relay availability. Relay outages must not
cause Kubernetes to restart the pod and discard an `emptyDir` queue.

## Local development and verification

```sh
NUON_GEN_TARGETS=runner go generate ./services/ctl-api
go -C sdks/nuon-runner-go generate .
go -C bins/telemetry-agent run go.opentelemetry.io/collector/cmd/builder@v0.150.0 \
  --config build-config.yaml --skip-compilation
go build -C bins/telemetry-agent/otelcol-build -o /tmp/nuon-otelcol
go build -o /tmp/nuon-telemetry-agent ./bins/telemetry-agent

/tmp/nuon-telemetry-agent --api-url=http://localhost:8083 --allow-insecure-api \
  --install-id=inl_example --token-file=/protected/path/bootstrap-token \
  --collector-binary=/tmp/nuon-otelcol --data-dir=/tmp/nuon-agent-state \
  --otlp-grpc-address=127.0.0.1:14317 --otlp-http-address=127.0.0.1:14318 \
  --health-address=127.0.0.1:14333 --collector-health-address=127.0.0.1:14334
```

`--allow-insecure-api` is for a trusted development API only; it never permits
an HTTP relay. TLS/proxy environment variables are allowlisted for the child;
parent bootstrap and cloud credentials are not passed through. `SSL_CERT_FILE`
can supply an additional relay CA bundle without replacing the system trust
pool. On Linux, the pool also includes roots from `SSL_CERT_DIR`.

The integration test runs the actual controller and pinned Collector against a
local simulated control plane and HTTPS OTLP receivers on isolated loopback ports.
It does not contact a cluster or cloud account. The end-to-end test requires
Linux to exercise combined file/directory CA trust; use the Docker target on macOS:

```sh
NUON_TEST_TELEMETRY_AGENT=/tmp/nuon-telemetry-agent \
NUON_TEST_OTELCOL=/tmp/nuon-otelcol \
  go test -race -count=1 -v ./bins/telemetry-agent

docker build --target test -f bins/telemetry-agent/Dockerfile .
```
