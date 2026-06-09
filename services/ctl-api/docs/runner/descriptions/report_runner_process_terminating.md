Reports that a runner process is terminating because its host VM is shutting down.

Runners send this best-effort beacon when they observe an operating-system
shutdown (e.g. an ACPI power event triggered by a customer stopping or
terminating the VM from their cloud portal). The runner only knows that the VM
is going away; the control plane attributes the cause: if Nuon issued the
shutdown there is an open shutdown record for the process, otherwise the
termination was initiated externally.

The endpoint takes no body and is idempotent. Because it races VM/network
teardown, delivery is not guaranteed — absence of a beacon is handled by the
existing offline-check fallback.
