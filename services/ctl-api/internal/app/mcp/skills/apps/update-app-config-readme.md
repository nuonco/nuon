---
name: update-app-config-readme
description: Write or update the README for a Nuon app or runbook (the `readme` field in metadata.toml / README.md). Use when asked to write, update, or improve an app README, runbook README, or install overview doc.
domain: apps
---

# Writing app and runbook READMEs

A README is a Go template string (`text/template` + [sprig](https://masterminds.github.io/sprig/) funcs) rendered
server-side per install. It must **never fail to render**. Missing, stale, or not-yet-populated data is normal —
render a clear placeholder or a banner telling the operator what to run, not a template error.

Reference docs:
- [Using READMEs](https://docs.nuon.co/guides/using-readmes) — static formatting, tables, callouts, and the full
  `<nuon-*>` display-component reference
- [Programmable READMEs](https://docs.nuon.co/guides/programmable-readmes) — template variables and live
  data components (`<nuon-config-graph>`, `<nuon-run-runbook>`, etc.)

## Required structure

Use the documented `<nuon-*>` components (see [Using READMEs](https://docs.nuon.co/guides/using-readmes)), not raw
ad-hoc HTML, so content matches dashboard styling:

- `<nuon-banner theme="warn|info|error|...">` — call out missing setup or data not yet populated
- `<nuon-status status="...">` / `<nuon-label-badge label="key:value">` / `<nuon-badge theme="...">` — status and
  metadata, never raw text like "Status: active"
- `<nuon-tabs>` / `<nuon-tab name="...">` — split a long README into sections (Application / Infrastructure / etc.)
- `<nuon-card>` / `<nuon-panel heading="..." trigger="View">` / `<nuon-modal>` — group related content, keep detail
  panels out of the main flow
- `<nuon-time time="..." format="relative">` — never print raw timestamps
- `<nuon-run-runbook name="...">`, `<nuon-component-card name="...">`, `<nuon-action-card name="...">`,
  `<nuon-config-graph>` — live per-install components (see
  [Programmable READMEs](https://docs.nuon.co/guides/programmable-readmes)); only render at install level (they
  render as inline code at app level, before an install exists)
- Markdown tables for tabular data — they render sortable with `<nuon-table-search column="...">` for filtering

## Defensive rendering — never let a nil or missing field break the render

The template engine errors on `missingkey=error`, and indexing a nil map panics. Guard every path into
`.nuon.*` data before using it:

**Never index a possibly-nil map directly.** Wrap the source in `default` first:

```gotemplate
{{ index (default dict .nuon.actions.workflows) "some_action" }}
```

not:

```gotemplate
{{ index .nuon.actions.workflows "some_action" }}  {{/* panics if .workflows is nil */}}
```

**Use `dig` to walk a path with a default at every step**, instead of chained field access that errors on the first
missing key:

```gotemplate
{{ $version := dig "outputs" "version" "unknown" $action }}
```

**`default` any dict/list before ranging or digging into it:**

```gotemplate
{{ $items := default dict (dig "outputs" dict $action) }}
{{ range $items }}...{{ end }}
```

**`coalesce` when a value can come from more than one source, first non-empty wins:**

```gotemplate
{{ coalesce (dig "version" nil $cluster) (dig "platform_version" nil $cluster) "—" }}
```

**`with ... else` for a single optional value — always give it a placeholder**, so the README shows a clean
em-dash instead of blank space:

```gotemplate
{{ with dig "updated_at" "" $data }}<nuon-time time="{{ . }}" format="relative"></nuon-time>{{ else }}—{{ end }}
```

**Check `kindIs "invalid"` before formatting a `dig` result with no default value**, to tell "genuinely absent"
apart from a real zero:

```gotemplate
{{ $v := dig "cpu_pct" nil $r }}{{ if kindIs "invalid" $v }}—{{ else }}{{ $v }}%{{ end }}
```

**Gate an entire section on whether its data is ready, not just whether the map key exists.** If a section depends
on an action or runbook that hasn't run yet, say so instead of rendering an empty table:

```gotemplate
{{ if and .populated (eq .status "finished") (gt (len $rows) 0) }}
<table>...</table>
{{ else if and .populated (eq .status "finished") }}
<nuon-banner theme="info">No results returned by the last run.</nuon-banner>
{{ else }}
<nuon-banner theme="warn">Waiting on {{ $actionName }}. Run it to populate this section.</nuon-banner>
{{ end }}
```

**Map raw status/enum values through a lookup dict with a `default` fallback**, so a status value you didn't
anticipate still renders (as neutral), not a template error or broken badge:

```gotemplate
{{ $theme := dict "active" "success" "error" "error" "pending" "warn" }}
<nuon-label-badge theme="{{ dig (lower $status) "neutral" $theme }}" label="{{ $status }}"></nuon-label-badge>
```

## Comments

Only comment the *why*, and keep it to one line. Use `{{/* ... */}}` for a constraint the template can't express
otherwise — e.g. why a list is hardcoded because the data source can't be read at render time. Don't comment what a
line does; the components and variable names already say that.

## Example

Bad — errors if `.workflows` is nil, blank cell if `updated_at` is absent, raw status text:

```gotemplate
{{ $wf := index .nuon.actions.workflows "sync_status" }}
<tr><td>{{ $wf.status }}</td><td>{{ $wf.updated_at }}</td></tr>
```

Good — guarded, placeholders, real components:

```gotemplate
{{ $wf := default dict (index (default dict .nuon.actions.workflows) "sync_status") }}
{{ $status := dig "status" "" $wf }}
<tr>
  <td>{{ if $status }}<nuon-status status="{{ $status }}" variant="badge"></nuon-status>{{ else }}—{{ end }}</td>
  <td>{{ with dig "updated_at" "" $wf }}<nuon-time time="{{ . }}" format="relative"></nuon-time>{{ else }}—{{ end }}</td>
</tr>
```
