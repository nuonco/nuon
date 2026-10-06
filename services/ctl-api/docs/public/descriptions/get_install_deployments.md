Returns a lightweight, chronological deployment feed for an install.

Each record represents one install-owned workflow that caused a real change: provisioning, reprovisioning, component deploys, input updates, stack reprovisioning, sandbox reprovisioning, and install-config updates. Action runs, runbook runs, and policy checks are returned by the activity feed. Plan-only and preview records are excluded.

Records include a `type`, workflow `status`, `title`, `activity`, `finished`, and slim workflow `steps` for progress and resource outcomes. Use the single deployment endpoint for app branch, image, affected resource, and change details.

Supports pagination via `page`/`offset`/`limit`/`has_more`, and filtering by `type`, `status`, `resource`, `search`, `created_at_gte`, and `created_at_lte`.
