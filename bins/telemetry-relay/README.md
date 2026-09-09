# Nuon BYOC Telemetry Relay

An OpenTelemetry Collector distribution that forwards logs, metrics, and traces from Nuon install runners to an
OTLP/HTTP backend. It verifies runner JWTs and replaces caller-supplied `nuon.*` attributes with verified identity.

The relay requires a Nuon control plane or a compatible token issuer; arbitrary JWTs are not supported.
It has no persistent queue or retry loop. Senders provide buffering, and delivery is not lossless.

## Configuration

The supplied [Collector configuration](config.yaml) uses these environment variables:

| Variable | Purpose |
|---|---|
| `NUON_TELEMETRY_ISSUER` | Expected JWT issuer |
| `NUON_TELEMETRY_JWKS_URL` | Issuer's public signing-key URL |
| `NUON_TELEMETRY_JWKS_ALLOW_INSECURE` | Permit HTTP issuer and JWKS URLs; defaults to `false` |
| `VENDOR_OTLP_ENDPOINT` | Backend OTLP/HTTP base URL |
| `VENDOR_OTLP_AUTHORIZATION` | Backend `Authorization` header value |

Issuer and JWKS URLs require HTTPS by default. Enable HTTP only over a trusted network; exact JWT issuer matching
and HTTPS certificate verification remain enforced. Keep private signing keys on the issuer.

The default configuration accepts authenticated OTLP/HTTP on port 4318 and serves health checks on port 13133.
Terminate TLS in front of the relay and keep health checks private.
