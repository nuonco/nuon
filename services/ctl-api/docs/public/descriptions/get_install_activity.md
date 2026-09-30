Returns a normalized, chronological activity feed for an install.

Each record is something that ran against the install rather than a change to it: an action run, a runbook run, or an install-scoped policy check. Policy checks recorded only against a component build, with no install, are omitted.

Records include a `type` (`action_run`, `runbook_run`, or `policy_check`), a `status` taken from that source, a human-readable `title` and `summary`, and a type-specific payload (`action`, `runbook`, or `policy`). Action and runbook records include a `workflow` reference when the run has one.

Supports pagination via `page`/`offset`/`limit`/`has_more`, and filtering by `type`, `status`, `search`, `created_at_gte`, and `created_at_lte`. `status` matches each source's own status string. Action and runbook runs use values such as `queued`, `in-progress`, `finished`, and `error`. Policy checks use `success`, `warning`, and `error`.
