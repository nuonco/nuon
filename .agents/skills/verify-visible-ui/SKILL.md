---
name: verify-visible-ui
description: |
  Render user-visible dashboard, readme, and runbook changes in the real UI and screenshot them before claiming they work or opening a PR. Always save a before and an after Amit can share with the team. Use a GIF when the change is motion or a click sequence. Use when editing dashboard-ui pages, Markdown, readme HTML, runbook tables, nuon-status, nuon-badge, nuon-time, nuon-panel, or any layout a person will see.
---

# Verify visible UI

A string check or a Go template render is not verification. The shared table component and the status badge change layout in ways the source does not show.

Also read the skill that matches the surface, before writing:

- Dashboard page or component: `frontend-builder`, then `services/dashboard-ui/DESIGN.md`.
- A Figma frame exists: `figma-design-to-code`. Do not invent a layout beside it.
- The user wants the flow clicked in the app: `source-command-dashboard-ui-e2e`.

## Before adding a surface

Search the dashboard and the readme for a view of the same data. Extend that view. A second table of the same rows is a failed design.

Reuse `Status`, `Badge`, and `Table`. Do not restyle them with one-off CSS to make a column fit.

## Render

1. If the source is a Go template, render it with `pkg/render` `RenderV2` against sample data. Include a runner with two process types, a count that contains the digits 8 and 9, and an empty version.
2. Put that HTML in a throwaway Ladle story beside `client/components/common/Markdown/Markdown.stories.tsx`. Name the file `*.tmp.stories.tsx`. Do not wrap `MemoryRouter` or `SurfacesProvider`. `.ladle/components.tsx` already does. A second router throws and the story is blank.
3. From `services/dashboard-ui`, run `bun run dev:ladle`. Read the story id from `http://localhost:61000/meta.json`. It looks like `ui--markdown--<export-name>`.
4. Screenshot with Playwright from `services/dashboard-ui` (it has `@playwright/test`). Viewport 1200px wide. Add `class="dark"` on `documentElement` so it matches the dashboard. Save a full-page png and a clip of the region that changed.
5. Measure in the page, and fail the change if any of these are true:
   - A box that should not scroll has `scrollWidth` greater than `clientWidth`. The shared `Table` sets `min-w-[800px]`, so a 2-column `<table>` inside a half-width column scrolls and clips. Use a grid of divs for a short count list.
   - A status line is taller than one line. `nuon-status variant="badge"` is `display:flex` and breaks out of `white-space:nowrap`. Put the label, badge, and uptime in one `display:flex; flex-wrap:nowrap` row.
   - A table that should sort has no `cursor-pointer` on its header, or a click does not reorder the rows. `htmlTableToExtracted` skips any table with `rowspan` or `colspan` greater than 1.
   - The HTML still contains a literal `####` heading. A markdown heading inside a raw HTML block is not converted. Use `<h4>`.
   - A count was built with zero-padded `printf` and sprig `int`. That parse is octal, so a count containing 8 or 9 becomes 0.
6. Delete the temp story and the screenshot script. Stop the Ladle process.

## Show the picture

Read the png so it is attached to the reply. A markdown link to a local path does not show up in chat. If the user says they cannot see it, `open` the png.

Do not open a PR, and do not say the layout is fixed, until that screenshot exists and the measurements pass.

## Before and after, for the team

Every visible change gets a before and an after. Show both in the reply. Do not push until Amit has seen them.

Save labeled side-by-side PNGs he can drop into Slack. Put them in a folder on the Desktop named for the task, one file per change, titled with what changed. Label the halves BEFORE and AFTER.

The before is the real prior state the comparison is about, not whichever commit happens to be HEAD. If the interesting change landed earlier, render that older source. On the runners page, the before he wanted was the page from before the version column and the two count tables, not the commit that already had those tables. Ask when it is ambiguous.

Use fixture data and fictional names. Do not put a real customer, org, install, or domain in an image meant for the team.

If the change is motion, a sequence of clicks, or anything a still cannot show, record a short GIF and save it next to the stills. Start Playwright with `recordVideo`, exercise the flow, close the context so the webm is flushed, then convert it with ffmpeg. Read the GIF so it attaches in chat. If ffmpeg is missing, say so and keep the stills.

Delete the temp story and the screenshot script. Stop the Ladle process.
