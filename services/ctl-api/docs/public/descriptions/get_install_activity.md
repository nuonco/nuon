Returns a normalized, chronological activity feed for an install.

Each record is a top-level operation that ran against the install rather than a change to it: a standalone action run, a runbook run, or an install-scoped policy check. Action runs executed as steps within another workflow are omitted. Policy checks recorded only against a component build, with no install, are also omitted.

Records include a `type` (`action_run`, `runbook_run`, or `policy_check`), a `status` taken from that source, a human-readable `title` and `summary`, and a type-specific payload (`action`, `runbook`, or `policy`). Action and runbook records include a `workflow` reference when the run has one.

For efficient pagination, request the first page with `limit`, then pass the returned `next_cursor` as `cursor` to fetch the next page. Keep filters unchanged while paging; restart without a cursor when changing filters. Records sort by timestamp, type, and ID descending. `has_more` indicates another page is available. Legacy `page`/`offset` pagination remains supported, but cannot be combined with a cursor and reads preceding records on deep pages.

Supports filtering by `type`, `status`, `search`, `created_at_gte`, and `created_at_lte`. `status` matches each source's own stored status string. Action and runbook runs use values such as `queued`, `in-progress`, `finished`, and `error`. Policy checks use `success`, `warning`, and `error`.
