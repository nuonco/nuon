Returns a normalized, chronological deployment feed for an install.

Each record represents one install-owned workflow that caused a real change: provisioning, reprovisioning, component deploys, input updates, stack reprovisioning, sandbox reprovisioning, and install-config updates. Action runs, runbook runs, and policy checks are returned by the activity feed. Plan-only and preview records are excluded.

Records include a `type`, workflow `status`, `title`, `summary`, workflow and app branch references, affected resources, and change groups. Component and image details are included when applicable.

Supports pagination via `page`/`offset`/`limit`/`has_more`, and filtering by `type`, `status`, `resource`, `search`, `created_at_gte`, and `created_at_lte`.

Use the deployment summaries endpoint for lightweight progress lists with lifecycle filtering and cursor pagination.
