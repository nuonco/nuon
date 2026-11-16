# bins

Binaries that are compiled and run as executables. Services that run as part of the control plane live in
[services](../services).

| Directory | What it is |
| --- | --- |
| [cli](cli) | The public `nuon` command line tool. Install it with `brew install nuonco/tap/nuon`. |
| [runner](runner) | Executes deploys and jobs inside the customer's cloud account. Runs in Kubernetes containers and cloud VMs. |
| [lsp](lsp) | Language server for Nuon TOML configuration files. |
| [telemetry-agent](telemetry-agent) | Forwards OTLP logs, metrics, and traces from an install to its telemetry relay. |
| [telemetry-relay](telemetry-relay) | OpenTelemetry Collector distribution that forwards install telemetry to an OTLP/HTTP backend. |

Each directory has its own `README.md` or `AGENTS.md` with build and usage details.
