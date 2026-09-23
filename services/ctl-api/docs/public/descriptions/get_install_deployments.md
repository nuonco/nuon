Returns a normalized, chronological deployment feed for an install.

Each record represents one install-owned workflow that caused a real change: provisioning, reprovisioning, component deploys, input updates, stack reprovisioning, sandbox reprovisioning, action runs, runbook runs, and install-config updates. Plan-only and preview records are excluded.

Records include a `type`, unified `status`, human-readable `title` and `summary`, an optional `workflow` reference, an optional `app_branch` reference (when the change originated from a branch run), an optional primary `component` reference (for single-component operations), a flat `affected_resources` list of component names, and `change_groups` that group the affected resources by logical category.

Supports pagination via `page`/`offset`/`limit`/`has_more`, and filtering by `type`, `status`, `search`, `created_at_gte`, and `created_at_lte`.
