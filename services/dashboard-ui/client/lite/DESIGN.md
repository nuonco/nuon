# Design guidelines

How Lite is organised and how to decide what to build. This document holds the
two frameworks everything in the app is built on:

1. **Five UX patterns** — every user-facing problem is one of five kinds, and
   each kind has an agreed set of treatments.
2. **Atomic design** — every component sits at one of five tiers, which decides
   what it may know about and what it may compose.

For *what things look like* — themes, tokens, type, spacing, motion — read
[STYLES.md](./STYLES.md). Engineering conventions are in [DEV.md](./DEV.md),
copy rules in [COPY.md](./COPY.md), and per-component rules in each component's
`Overview` story.

## The premise

Lite exists to improve the dashboard's usability **by removing surface area.**
Every rule here serves that. When the choice is between more expressiveness and
one fewer thing to learn, take the second.

### The lens: current state, and when it changed

Users come to this product with two questions, and essentially only two:

1. **What is the state of my system right now?**
2. **What changed, and when?**

Everything else is in service of those. This is the lens to hold against any
proposed screen: if it does not answer one of them, it probably should not be
built.

**The consequence is a preference order for how we show things.** A system's
state is a *shape* — components, resources and their dependencies — so the
truest view of it is a **graph**. Change is a *sequence*, so the truest view of
it is a **timeline**. A table of resources is neither: it is the data model
transcribed, and it makes the reader reconstruct both the shape and the history
in their head.

So: **prefer topology graphs and change timelines over lists and tables of
resources.** Reach for a table only when the user's real task is comparing
records field by field.

Lite does not fully reflect this yet — the foundational work came first — so
expect to find tables where a graph or timeline belongs. That is a known gap,
not the target.

### What this rules out

The organising question is **"what exists, and what happened?"** — not "what is
in the data model". A concept in the API does not earn a page, a list or a nav
item by existing. **Do not create a list for every noun.** Most nouns are better
shown as a node in a graph, an event in a timeline, or a field on the one
resource the user actually cares about.

## How to use these two frameworks

Before building anything, answer two questions:

1. **Which UX pattern is this?** Navigation, discovery, resource management, task
   flow, or system guidance. That tells you which treatments already exist and
   which decisions are already made.
2. **Which atomic tier is this?** Atom, molecule, organism, template, or page.
   That tells you what it may depend on and where it lives.

The point of both frameworks is that **new work should mostly be assembly, not
invention.** A feature needing a new pattern *and* a new organism is a signal to
look again, not a licence to build both.

## Breaking these rules

**The default answer is no.** Almost everything that feels like it needs a new
pattern, a new tier, a new treatment or a bespoke component is an existing one
you have not found yet, or two existing ones composed. Assume that first.

If you are an agent and you conclude a rule here has to be broken:

1. **Stop. Do not build it.** Not as a draft, not "to show the option".
2. **Say plainly that you want to break a documented rule**, and name the rule.
   Do not present it as a neutral design choice or bury it in a summary — the
   person reading has to know a rule is on the table before they can weigh it.
3. **Show the work that got you there.** Which existing patterns, treatments and
   components you evaluated, and the specific reason each one fails. "It didn't
   feel right" is not a reason. If you cannot name what you ruled out, you have
   not looked hard enough to be asking.
4. **Offer the closest compliant alternative**, including what it costs. There is
   almost always one, and it is often fine.
5. **Wait for an explicit decision.** Silence, "sounds good", or a reply about
   something else is not approval.

This exists so a teammate is never handed a rule change disguised as an
implementation detail. **Escape hatches should be rare.** An agent reaching for
one repeatedly in a session is misreading these documents, not finding genuine
gaps — reread them before asking again.

If a rule here turns out to be genuinely wrong, the outcome is a change to
*this document*, agreed first, not a one-off exception in one component.

---

# Atomic design

The tiers are not a filing system, they are a **dependency and reuse
contract**. Knowing the tier of a thing tells you what it is allowed to know.

| Tier | Knows about | Never knows about |
|---|---|---|
| **Atom** | Its own props. Tokens. | Data shapes, routes, other components |
| **Molecule** | Atoms. One small domain concept. | Fetching, routes, layout |
| **Organism** | Molecules, atoms. A real API resource. | Where on the page it sits |
| **Template** | Layout and slots. | Which resource is being shown |
| **Page** | Routes, providers, which organisms to compose. | How anything renders |

**A component imports from its own tier or below, never above.** An atom
importing a molecule means the atom is really a molecule, or the shared part
needs extracting downward.

## What each tier is for

**Atoms** are the vocabulary — `Text`, `Button`, `Icon`, `Badge`, `Status`,
`Card`, `Input`, `Link`, `Tooltip`. They are generic: an atom that knows what an
install is has been built at the wrong tier. Atoms own their own visual
behaviour completely, which is why `Button` owns its tooltip and its loading
spinner rather than having callers wrap it.

**Molecules** are the smallest units that mean something in *this* product —
`ID`, `Time`, `Duration`, `Cron`, `CloudRegion`, `ComponentType`,
`CommitSummary`, `OverviewCard`, `Pagination`. A molecule turns a value into its
canonical presentation. This is the tier that does the most work for
consistency: because `Time` exists, no two places format a timestamp
differently.

**Organisms** are a resource, rendered — `InstallsTable`, `AppConfigDiff`,
`BranchSwitcher`, `Timeline`, `Toast`, `Modal`. This is the tier that may fetch,
and it fetches for *itself* via a `*Container`. An organism does not know what
else is on the page.

**Templates** are structure with holes in it — `DashboardShell`, `FocusShell`,
`PageTransition`. A template knows about sidebars, headers, status bars and slots.
It never knows which resource is in the slot, which is what makes it storyable
and what makes every page share one shell.

**Pages** are what the router mounts. A page's whole job is: declare its title
and breadcrumbs, mount the providers it needs, and compose organisms into a
template. **Pages do not fetch.** A page that holds a dozen queries is the
failure mode this whole structure exists to prevent.

## Deciding the tier of something new

Work down, not up:

1. Can an existing component do this with different props? Then do that. A
   second component that differs only in padding is the most common mistake.
2. Is it generic and value-free? → **atom**.
3. Does it present one domain value in its canonical form? → **molecule**.
4. Does it represent an API resource, or need to fetch? → **organism**.
5. Is it layout with slots? → **template**.
6. Is it a route? → **page**.

## Composition over configuration

Prefer combining existing pieces to adding props to one piece.

- `Card` composes four independent axes — padding, opacity, blur, shadow —
  rather than offering `variant="toast" | "popover" | "panel"`. Adding a named
  variant per use site is how a primitive rots.
- `Status` has four *variants* (`chip`, `inline`, `dot`, `icon`) because those
  are genuinely different shapes of the same idea, not four different ideas.
- A `Badge` and a `Status` look similar and are not the same thing. Metadata is a
  badge; state is a status. Keeping them separate is worth the resemblance.

The signal that you are configuring instead of composing: a boolean prop that
changes what the component *is* rather than how it looks.

---

# The five UX patterns

Every user-facing problem in the app is one of these five. Each section says
what it covers, what Lite has chosen, and how to pick between the options.

## 1. Navigation

*How users move between areas, understand hierarchy, and backtrack.*

**Two levels of navigation today:**

1. **Main sidebar** — the org's top-level areas. Dashboard, Apps, Installs,
   Team, Settings.
2. **Resource context nav** — within one resource: install pages, app branch
   pages, org settings.

For moving between *areas*, two levels is the rule. Something asking for a third
level of area navigation is telling you the information architecture is wrong.

### Open: navigation inside a run detail page

A run — a workflow, a deployment, a build, an action run — has several
independent concerns (summary, logs, plan, trace) and will likely need its own
level of navigation within the page. **This is not yet decided.**

It is probably not a third level of the hierarchy above but a different
mechanism: navigation *within one resource* rather than *between areas*. The
production dashboard solves it with a routed tab nav, which is the obvious
starting point but not a decision Lite has made.

**Do not invent this.** If you are building a run detail page and need it, that
is a "Breaking these rules" conversation — stop and raise it, so the answer gets
made once and written here rather than five times in five pages.

**Treatments:**

- **Sidebar `NavLink`** for the top level, each with a `g`-prefixed keyboard
  shortcut.
- **`SubNav`** for resource context.
- **`Breadcrumb`** for location and backtracking. It is the *only* upward
  affordance — no per-page back buttons scattered around.
- **Switchers** (`OrgSwitcherMenu`, `BranchSwitcher`) for moving sideways between
  peers, not for hierarchy.

**Rules:**

- Every navigation path opts into a view transition. Enforced by
  `guardrails/view-transitions.test.tsx`.
- Navigation is a `Link`, never a button styled as one.
- A resource's name is its own link — no "view" verbs, no trailing chevrons.
- **Do not add a list page for every noun in the data model.** A nav item is a
  claim that users organise their work around that concept.

## 2. Discovery

*How users browse, search, filter, sort, scan, and drill into detail.*

Two decisions, in this order: **what shape shows this**, then **panel or page**.

### Decision 1 — what shape?

Follow the preference order from the premise. Ask what the user is actually
asking, not what the endpoint returns:

| The user is asking | Shape |
|---|---|
| "What is my system, and how is it wired?" | **Graph** — topology, nodes and dependencies |
| "Is it healthy right now?" | **Status tiles / overview cards** — a few facts, no scrolling |
| "What changed, and when?" | **Timeline** — events in time order, grouped by day |
| "Which of these N records differs?" | **Table** — and only then |

**A table is the fallback, not the default.** Before adding one, check whether
the same information is a graph of the thing's parts, a timeline of what
happened to it, or a handful of tiles. Reach for a table when the task genuinely
is field-by-field comparison across many records — that is what columns are good
at and nothing else is.

Where a table is right, `Table` is responsive: stacked cards on small screens,
optional card view on desktop. **Search, filter and pagination must work
identically in both layouts.**

### Decision 2 — panel or page?

**Provisional.** Most of these pages are not built yet, so the split below is
the current intent rather than a settled rule. Follow it as the default, but if
a case genuinely does not fit, **raise it instead of quietly diverging** — the
lists are meant to converge on a rule as the pages get built, and a silent
one-off is what stops that happening.

The axis is **whether the thing is a top-level resource in its own right, or a
part of one.**

**Panels — configuration and detail of a part.** You inspect these *in the
context of their parent*, so the parent stays on screen:

- stack config and details
- sandbox config and details
- component config and details
- runner config and details
- policies, and role config and details
- input details

**Pages — top-level resources, and runs.** These are places you go:

- *Top-level resources in the data model* — app branch, install
- *Run details* — workflows, builds, deployments, action runs, runbook runs

The reason runs get a page rather than a panel is that a run has several
independent concerns — summary, logs, plan, trace — and a panel cannot hold
them without becoming a page in a drawer.

**Placing something not listed:** ask whether a user would ever navigate to it
as a destination, or only ever open it while looking at something else. A
destination is a page. Everything else is a panel.

A graph node opening a panel is the canonical case: the graph stays on screen
and stays the user's mental model while they inspect a part of it.

**Rules:**

- **`ListSearch` / `FilterMenu` / `Pagination`** are owned by the collection
  component, never added at page level.
- **Surfaces are addressable.** A panel or modal is reachable by URL via
  `usePanelHref` / `useModalHref`, so a view can be linked and the back button
  behaves.
- List state — search, filters, offset — lives in the URL, never in component
  state. A filtered list that cannot be linked is broken.
- Filters offer *isolate* (show only this) as well as toggle, because that is
  what people actually want when scanning.
- Never a page-level search box competing with the collection's own.

## 3. Resource management

*How users create, edit, delete, and act on resources.*

**Treatments by weight:**

- **Inline** — a toggle or an editable field, for reversible single-value changes.
- **Modal** — a confirmation, or a create/edit form that fits one screen.
- **Full-screen wizard** — creation that spans real decisions. See Task flows.
- **Dropdown menu** — the home for a resource's secondary and destructive
  actions, so rows and headers do not sprout buttons.

**Rules:**

- **Destructive actions confirm**, and the confirmation names the thing being
  destroyed. `danger` is for the button *inside* the confirmation, not the one
  that opens it.
- **There is no multi-select.** Actions that apply to a whole set — "deploy all
  components" — are a single action on the parent with an ordinary confirmation
  naming what it covers. Do not build checkbox columns, selection state, or a
  bulk-action toolbar; select-then-act is not a pattern this app has.
- Edit reuses create when they hit the same endpoint; separate forms when they
  do not. (See DEV.md and the `form` recipe.)
- Every mutation invalidates the queries that display what it changed.
- A disabled action explains why, via the button's own tooltip.

### Button emphasis

Emphasis is a resource-management decision, not decoration:

- **`secondary` is the default.** Standalone actions that must stay findable
  without surrounding context.
- **`primary` is the one action the page wants.** At most one per view.
- **`ghost` is lowest emphasis.** Inside framed chrome only — toolbars, grouped
  controls, table rows, surface footers.
- **`danger`** for destructive confirmation only.
- Never a ghost button as the only action in a section or page.
- In a toggle group, the selected option is `secondary`, the rest `ghost`.
- A row of related toolbar controls must not become a row of competing
  `secondary` buttons.

Sizes are `md` (default) and `sm`. `sm` is for dense chrome — not for making an
action less important. That is what variants are for.

## 4. Task flows

*How users complete multi-step interactions — setup, approvals, handoffs.*

**The decision: wizard or modal?** A flow earns a **full-screen wizard** when it
spans more than one real decision, or when its result needs watching. Everything
else is a **modal**.

**Full-screen wizard** (`FocusShell`) — the shell strips navigation so the task
is the only thing on screen:

- *Create app and connect a branch* — name the app, connect the repo if needed,
  connect the branch, create the deployment plan.
- *Create install from app* — select app config, enter install info, assign to an
  app branch, then provision. **The first provision gets its own page**, because
  a user's first install is the moment they most need to see progress.

**Modal creation** — one form, one endpoint: API token, service account,
webhook, Slack channel, trust policy.

**Rules:**

- A wizard's steps are named for the user's decision, not the API call.
- Progress is visible and steps are re-enterable; never trap someone mid-flow.
- Errors surface in the form, not as a toast — a toast for a form error is a
  message that disappears while the broken field stays broken.
- The end of a flow lands on the thing that was created.

## 5. System guidance

*How the interface communicates status, validation, errors, confirmations,
undo, empty states and help.*

This is the pattern most often skipped and most responsible for an app feeling
unfinished. Every surface that can be empty, loading, stale, failed or
in-progress needs a decided answer.

**Treatments, by what is being communicated:**

| Situation | Treatment |
|---|---|
| A resource's state | `Status` — `chip` standalone, `dot` in dense rows, `inline` with a label, `icon` when a tooltip carries the label |
| Work in progress | `Status` plus `status-pulse`; a `Spinner` only for indeterminate waits |
| Async work started | `Toast`, `info` theme, present-progressive heading |
| Async work finished | `Toast` on the status transition, via `useStatusToast` |
| Form submission failed | `FormErrorBanner` inside the form. Never a toast |
| A collection is empty | The collection's `emptyState` — "No X yet" plus what will make them appear |
| A collection failed to load | Also `emptyState`, with failure wording: "Installs failed to load" |
| Why an action is unavailable | The button's own `tooltip` |
| A field is invalid | Field-level error, after the field is touched |

**Rules:**

- **Empty and failed are different messages in the same slot.** A collection
  that shows "No installs yet" when the request failed is lying.
- **Never a toast for something the user must act on.** Toasts are for things
  that happened, not decisions to make.
- Status colour is never the only signal — every theme carries its own icon.
- An in-progress state must resolve on its own. A spinner that can spin forever
  is a missing error state.
- Guidance is copy, so it follows [COPY.md](./COPY.md): sentence case, no
  "successfully", no exclamation marks.
