# Production UI patterns

What to copy when building a dashboard page. Visual rules live in [DESIGN.md](../../DESIGN.md). Copy rules live in [COPY_STYLE.md](../../COPY_STYLE.md). Read the story before using a component.

Onboarding, Lite, and `playground/` are not precedents. Do not copy them into a dashboard page.

## Page shells

| Need | Use | Story | Shipped page |
| --- | --- | --- | --- |
| Org-level list | `ListPage variant="page"` | `layout/ListPage.stories.tsx` | `views/org/Installs.tsx` |
| List inside an app or install layout | `ListPage` (default `variant="section"`) | same | `views/install/Deployments.tsx` |
| Section that is not a list | `PageSection` + `SectionHeader` | `layout/SectionHeader.stories.tsx` | `views/settings/General.tsx` |
| A run (deploy, sandbox run, action run) | `DetailPage` + `DetailHeader` + routed `TabNav` | `layout/DetailPage.stories.tsx` | `views/install/DeployLayout.tsx` |
| One configured resource | `DetailPage` + `DetailHeader` + `HistoryPanelButton` | `layout/HistoryPanelButton.stories.tsx` | `views/install/Sandbox.tsx` |

Avoid: a hand-built `PageHeader` / `HeadingGroup` / `PageHeadingGroup` in a view, a history column beside the page, unrouted `Tabs` as page structure, `PageLayout` inside a layout `Outlet`.

## Lists

| The items | Use | Story | Shipped page |
| --- | --- | --- | --- |
| Exist, and you compare them by column | `Table` | `common/Table.stories.tsx` | `views/org/Installs.tsx` (`InstallsTable`) |
| Happened, sorted by time, each row has a status | `Timeline` | `common/Timeline.stories.tsx` | `components/workflows/WorkflowTimeline/` |
| Content is the point, nothing to compare | `Cards`, and only with a written exception | — | Announcements, `BranchCards`, `ActiveWorkflows` are the exceptions. Do not add another. |

Drill-down is one of: its own page, a panel opened from the identity cell (`panelTriggerClass`, shape of `InstallResourceDetailPanel`), or an expand of a few lines. A diff, nested collection, tabs, or row actions are not an expand.

## Labeled data, status, identity

| The string | Use | Story |
| --- | --- | --- |
| A field we named (`Status`, `Created`) | `LabeledValue` / `LabeledStatus` | `common/LabeledValue.stories.tsx` |
| A key/value map from the API or the user | `KeyValueList` | `common/KeyValueList.stories.tsx` |
| Records with columns we named | `PropertyGrid`, until it needs search, sort, or row actions — then `Table` | `common/PropertyGrid.stories.tsx` |
| A lifecycle that changes on its own | `Status` | `common/Status.stories.tsx` |
| A static classification | `Badge` | `common/Badge.stories.tsx` |
| A name, id, path, tag, or other copy-paste string | `ID`, `Code`, `Hash`, `ClickToCopy`, `Text family="mono"` | `common/ID.stories.tsx` |
| A timestamp or duration | `Time`, `Duration` | `common/Time.stories.tsx` |

Vocabulary from our API goes through `humanize()`. Do not re-case it at the call site.

## Actions

| Need | Use | Story | Shipped example |
| --- | --- | --- | --- |
| A button | `Button` | `common/Button.stories.tsx` | — |
| A disabled reason | `tooltipProps` on `Button` | same | — |
| A confirm, then a mutation | `Modal` via `useSurfaces().addModal` | `surfaces/Modal.stories.tsx` | `components/approvals/ApprovePlan.tsx` |
| A form | TanStack Form + Zod + `common/form/` | `api-tokens/CreateApiToken/CreateApiToken.stories.tsx` | `components/api-tokens/CreateApiToken/` |
| Navigation by resource name | `Link` | `common/Link.stories.tsx` | — |
| An icon | `Icon` | `common/Icon.stories.tsx` | — |
| Loading a known layout | The real component with `loading` (`Text`, `LabeledValue`, `ID`, `Time`, `Badge`, `Status`, `Code`) or `isLoading` (`Table`, `Timeline`) | `common/Text.stories.tsx`, `common/Table.stories.tsx` | — |
| Loading an unknown shape | `Loading variant="large"` | `common/Loading.stories.tsx` | Plan diffs |

Loading is the component drawing itself empty. Do not render `Skeleton`, `TableSkeleton`, or `TimelineSkeleton` from a feature, and do not write a `*Skeleton` component. A hand-built skeleton is a second copy of the layout that goes stale.

Avoid: a raw `<button>`, `<input>`, `<select>`, or `<textarea>` in a feature; `Skeleton` / `TableSkeleton` / `TimelineSkeleton` or any `*Skeleton` outside `common/`; `Tooltip` wrapped around `Button`; a second `variant="primary"` on the same page.

## Choosing a control

These are the places people invent a one-off instead of reaching for what already exists. Each one is a test, not a preference.

**Filtering a list.** One value → `RadioFilterDropdown`. Several values → `CheckboxFilterDropdown`. Both are a funnel-icon dropdown with a built-in clear action. `Select` / `FormSelect` are for form input, not for filtering a list — a `Select` driving a table or timeline filter is a defect, not a style choice. Stories: `common/RadioFilterDropdown.stories.tsx`, `common/CheckboxFilterDropdown.stories.tsx`.

**Size.** Every component with a `size` prop (`Button`, `Badge`, `Avatar`, ...) defaults to its default size. Do not pass `size` without a reason the PR states — a dense table cell, a toolbar the design calls for. "It looked too big" is not a reason. `ToggleButton` sets its own internal size; never pass `size` to it either.

**Button variant.** One question per button: what is it?
| It is | Variant |
| --- | --- |
| An ordinary action | `secondary` (the default — pass nothing) |
| The single most important action on the screen | `primary` — at most one per page, see DESIGN.md §4.8 |
| A destructive action | `danger` |
| UI chrome: opening/closing the sidebar, expanding a diff section or row, the `ShowMore` control, the panel-trigger link on a table row | `ghost` (or `icon` for icon-only chrome) |
| A tab inside `Tabs` / `TabNav` | `tab` — owned by those components, never set by hand |

A `ghost` button doing an ordinary action — a form submit, a create button, a row action — is a defect.

**`ToggleButton` vs `Tabs` vs `TabNav`.**
| Test | Use |
| --- | --- |
| Same content, 2–4 ways to display it in place (wrap/no-wrap, list/graph, unified/split) | `ToggleButton` — `common/ToggleButton.stories.tsx` |
| Different content panels inside one section | `Tabs` — `common/Tabs.stories.tsx` |
| Page structure, each tab has its own URL | Routed `TabNav`, see PATTERNS.md "Page shells" |

Never hand-roll a segmented control with `aria-pressed` buttons — that is one of the three above.

**`Expand` vs `ShowMore`.**
| Test | Use |
| --- | --- |
| A collapsible section under its own heading (a step, a diff section, a group of rows) | `Expand` — `common/Expand.stories.tsx` |
| Overflowing content with no heading — a commit message, a long note | `ShowMore` — wrap the content. It measures its own children, clips at `lines` (default 3) or `maxHeight`, fades the bottom, and renders the toggle only when the content actually overflows. Short content gets no button at all. `common/ShowMore.stories.tsx`. Used for the commit body on the app overview (`RunSourceCard`) |
| The small label + icon on a full-row collapsible header | Inline `Text` + `Icon`, not a shared component — the row itself is the control. DESIGN.md "Collapsible header" has the exact markup to copy |

Do not hand-roll another content clip — `ShowMore` is that control. Revealing more cards or rows is a different action — `BranchCards` uses a `Button`, not `ShowMore`.

## Tokens

Colors, type, radius, and spacing come from Stratus (`styles.css`, `Text`, `Card`). The palette is `primary`, `blue`, `green`, `orange`, `red`, `cool-grey`, `dark-grey`, plus the semantic tokens in DESIGN.md §2.

Avoid: `neutral` / `slate` / `gray` / `zinc` and the rest of the stock Tailwind palette, raw hex, `gap-[13px]` and other off-scale pixel values, `border-2` or a colored border, `space-y-*` / `space-x-*`.
