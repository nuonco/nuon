# Telemetry agent

Forwards OTLP logs, metrics, and traces from existing producers to a Nuon install's
configured telemetry relay. It does not scrape Kubernetes, read node logs, or
require Kubernetes API access.

## Bootstrap and deployment

The agent requires:

- A reachable HTTPS Nuon **runner API** and an install ID.
- An opaque API token for the install's managed telemetry service account.
  Prepare the account and issue its token through the Nuon public API. Do not
  substitute an org-wide runner or administrator credential.
- A container image accessible to the cluster.

Store the bootstrap token in a protected file or Kubernetes Secret, never in
source control, Helm values, or command-line arguments. Build the image from the
repository root:

```sh
docker build --target final -f bins/telemetry-agent/Dockerfile \
  --build-arg AGENT_SOURCE=bins/telemetry-agent \
  --build-arg BUILD_IMAGE=golang:1.26-alpine \
  --build-arg RUNTIME_IMAGE=alpine:3.21 \
  -t registry.example.com/nuon-telemetry-agent:0.1.0 .
```

Deploy using the [telemetry-agent Helm chart](https://github.com/nuonco/charts/tree/main/charts/telemetry-agent).
The chart documentation covers Secret creation, required values, producer
endpoints, and persistent storage. Enable forwarding in the install's Nuon
telemetry settings; creating an account or deploying the chart does not enable it.

## Operational notes

- **OTLP ingress is unauthenticated.** Restrict it to trusted producers for this
  install with network policy. Do not expose it publicly or aggregate unrelated
  installs. Outbound relay connections require HTTPS with certificate verification.
- Disabled agents remain healthy and continue checking settings, but do not accept
  new telemetry. Producers must retry during disabled periods and brief
  interruptions when configuration changes.
- Disk-backed buffers are bounded to 1 GiB per signal. Budget additional space for
  storage overhead. Use a dedicated PVC to retain data across Pod replacement;
  the chart's default `emptyDir` does not provide that durability. Producers must
  retry requests rejected when buffers are full. Permanent relay rejection can
  discard accepted data; delivery is not exactly-once.
- Rotate the bootstrap token before expiry. Issue a replacement, update the Secret,
  then revoke the old token after the projected update is visible. Rotation does
  not require an agent restart. Already-issued relay credentials remain valid
  until expiry after bootstrap revocation.
- Liveness checks controller availability, not relay availability. Do not restart
  the agent merely because the relay is unreachable.

`SSL_CERT_FILE` supplies additional relay CA trust without replacing system roots;
Linux also supports `SSL_CERT_DIR`. `--allow-insecure-api` is for trusted local
development only and never permits an HTTP relay.
