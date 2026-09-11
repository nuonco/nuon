# Runner telemetry export

The vendor Collector receives OTLP logs, metrics, and traces on ports 4317 (gRPC)
and 4318 (HTTP) and forwards them to the configured telemetry relay using a runner
telemetry token. It runs only for install runners with vendor telemetry enabled.
The separate audit-export Collector is configured independently and is not affected
by the resource enrichment described below.

## Install resource attributes

The runner-settings API supplies current org, app, and install names and all resolved
install labels. The vendor Collector adds them to every incoming resource:

| Attribute | Example |
| --- | --- |
| `nuon.org.name` | `acme` |
| `nuon.app.name` | `payments` |
| `nuon.install.name` | `production-eu` |
| `nuon.install.labels.<label-key>` | `nuon.install.labels.tier=enterprise` |

Label keys and string values are preserved literally, including punctuation. Labels
include materialized app defaults and resolved template values, not template text.
All install labels are exported: **do not store secrets in labels**. Frequently
changing values can increase metric cardinality and storage costs.

The Collector overwrites the three name attributes and replaces the
`nuon.install.labels.*` resource attributes with the current install labels. It
preserves other resource attributes, including `service.*` and cloud metadata,
and does not modify log-record, span, or metric-datapoint attributes. Names and
labels are descriptive metadata, not authenticated identity; the relay enforces
the JWT-derived identity IDs separately. Backends may require explicit resource
attribute promotion to make these values available as metric labels. Avoid
indiscriminately promoting all resource attributes.

Settings are polled every 15 seconds. Name or label changes restart only the vendor
Collector and may briefly interrupt ingestion; unchanged settings do not restart
it. Failed settings fetches retain the active configuration, and failed replacement
starts attempt to restore the previous configuration. Deleted labels disappear
from newly processed resources after a successful refresh. Already queued telemetry
retains its original attributes.

Deploy API support before updated runners. Older runners ignore the added settings
field; an updated runner receiving settings without it retains passthrough behavior.
This enrichment does not enable additional telemetry sources or alter the relay's
environment pipeline.
