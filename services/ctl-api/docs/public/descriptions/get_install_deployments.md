Returns a lightweight, chronological deployment feed for an install.

Each record represents one install-owned workflow that caused a real change: provisioning, reprovisioning, component deploys, input updates, stack reprovisioning, sandbox reprovisioning, and install-config updates. Action runs, runbook runs, and policy checks are returned by the activity feed. Plan-only and preview records are excluded.

Records include a `type`, workflow `status`, `title`, `activity`, `finished`, and slim workflow `steps` for progress and resource outcomes. Use the single deployment endpoint for app branch, image, affected resource, and change details.

Supports pagination via `page`/`offset`/`limit`/`has_more`, and filtering by `type`, `status`, `resource`, `search`, `created_at_gte`, and `created_at_lte`.

`state=active` returns deployments whose workflow status is pending, queued, in progress, retrying, awaiting approval, approved, or failed pending retry. `state=finished` returns every other status. When `state=active`, `total` counts all matching active deployments, ignoring `limit` and `cursor`.

`sort=attention` orders deployments awaiting approval first, failed pending retry second, then all others. Each group is ordered newest first. The default order is newest first.

When `has_more` is true, `next_cursor` is an opaque cursor for the next page. Pass it back as `cursor` with the same `state` and `sort`. A cursor cannot be combined with a non-zero `page` or `offset`. An invalid `state`, `sort`, or `cursor` returns 400.
