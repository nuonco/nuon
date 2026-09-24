# Flows

Who uses Lite, what they are trying to get done, and what "done" looks like.

Read this before building or changing a multi-step flow. It decides *whether* a
flow should exist and how it should feel; the mechanics of building one are in
[DESIGN.md](./DESIGN.md) §4 "Task flows".

## This file is deliberately short

It holds **eight flows** — the ones that have actually been decided. It is not
an inventory of everything the product does, and the gaps are not oversights.

**Do not add a flow here by inferring one** from a route, a production
`e2e/flows/` doc, or an API endpoint. A flow gets written down when someone has
decided who it serves and what must not break, which is a product decision, not
something to be reconstructed from the code. An invented flow is worse than a
missing one, because it looks decided.

If you need a flow that is not here, that is a conversation — see DESIGN.md
"Breaking these rules".

## Two layers, deliberately separate

| Layer | Where | Changes |
|---|---|---|
| **Intent** — who, why, what done means | this file | rarely |
| **Steps** — selectors and assertions | `e2e/flows-lite/*.flow.md` | every time the UI moves |

They are apart because a flow's intent is stable while its selectors churn. One
file would mean the volatile half constantly rewriting the stable half.

The join is the **Done** and **Must not break** lines: each is written so it
translates into an `expect:` line in the step doc. If you cannot turn a Done
line into an assertion, it is not specific enough yet.

Step docs follow the format in `e2e/flows/README.md` — same DSL as the
production suite, separate directory so the two never collide.

## Personas

**1. Platform engineer** *(primary)* — works at the software vendor. Configures
apps, branches and deployment plans; owns the config repo. Lives in app and
branch surfaces. Technical, reads Terraform plans closely, is the person an
approval is asking.

**2. Operator** *(primary)* — also at the vendor, often the same person wearing
a different hat. Runs and supports customer installs. Lives in install
surfaces.

The two split by **surface, not seniority** — apps and branches are
configuration, installs are operation. When a flow feels like it serves both,
check which surface it lives on.

**3. Customer** *(earmarked — no surfaces today)* — the vendor's end customer,
whose cloud account the app is installed into. They have **no access to the
dashboard or to Lite at present.** A read-only view of their own install — its
state, recent changes and health — is wanted and worth designing toward.

**Do not build customer-facing surfaces yet.** But when designing an install
surface, it is worth asking which parts a customer could safely see, since that
is cheaper to consider now than to retrofit.

## How to read an entry

```markdown
## Flow name

- **Who** — persona
- **Job** — When {situation}, I want to {motivation}, so I can {outcome}.
- **Trigger** — what brings them here
- **Path** — the decisions, at UX altitude. Not selectors.
- **Done** — the observable end state
- **Must not break** — what holds, whatever else changes
- **Shape** — wizard / modal / inline / page (see DESIGN.md §4)
- **Status** — built / partial / not built, in Lite
- **Spec** — the step doc, or none
```

**Job** is a job story, not "as a user I want" — the situation is the
load-bearing part, because it says when the flow gets reached.

---

# Setup

The three setup flows share one mechanic: **the wizard is the whole create
flow, and the real pages are gated behind it.** Collect inputs, create the
record, watch it come up — and only then leave setup. Identity lives in the
query (`?appId=`, `?installId=`) so the user stays on one path throughout, and
opening `/apps/:appId` or `/installs/:installId` too early redirects back into
setup. Server state is the source of truth for which step you are on; stored
progress is only a hint. Details in `.planning/lite/plans/setup-wizards.md`.

## 1. First run

- **Who** — platform engineer
- **Job** — When I first sign in, I want to get to the point of configuring an
  app, so I can start shipping.
- **Trigger** — first sign-in, or accepting an invite
- **Path** — sign in → **the org is created automatically** → get oriented →
  start creating an app
- **Done** — the user is in an org and on their way into app setup
- **Must not break** — the user is never asked to name an org before they have
  done anything; a self-signup user never sees an empty org picker; an invited
  user lands in the org they were invited to
- **Shape** — wizard in `FocusShell` (no org context to return to)
- **Status** — scaffold only (`/onboarding` renders a placeholder)
- **Spec** — none

**Decided:** Lite onboarding **auto-creates the org** — there is no org-naming
form in the flow. Renaming happens later from org settings, once the user has
context for what to call it. The remaining shape is open and gets settled in
the onboarding plan.

## 2. Set up an app

- **Who** — platform engineer
- **Job** — When we want to ship our product into customer clouds, I want to
  describe it once as an app, so installs can be created from it.
- **Trigger** — a new product, or a new deployable surface of one
- **Path** — name the app → **connect the GitHub account** → connect the branch
  → create the deployment plan
- **Done** — the app exists with at least one branch, and the branch has a
  deployment plan
- **Must not break** — a half-created app never blocks a retry; a user who
  already has a GitHub connection is not made to make another
- **Shape** — wizard in `DashboardShell`
- **Status** — page stub (`/:orgId/apps/setup`), wizard not built
- **Spec** — none

**Decided:** the GitHub connection happens **inside this flow**, not as a
prerequisite the user has to go find. It is also reachable from
`settings/connections`, so both entry points share one surface rather than
forking.

The flow runs **all the way through app branch creation, including the
deployment plan** — it does not stop at "app created". Whether the deployment
plan step can be skipped is open; assume required until the plan says
otherwise.

## 3. Set up an install

- **Who** — operator
- **Job** — When we take on a new customer, I want to deploy our app into their
  cloud account, so they can start using it.
- **Trigger** — the vendor decides to stand up a new customer or environment.
  **The customer is not involved** — customer interaction happens outside Nuon,
  and only the vendor can create an install.
- **Path** — select the app config / branch to create from → enter the required
  install info (name, region, inputs) → optionally assign it to an app branch
  deployment plan install group → **watch the install provision**
- **Done** — the install is provisioned, and only then does the operator reach
  its real pages
- **Must not break** — the operator stays inside the setup flow watching the
  provision step until it finishes; the install's normal pages are not
  reachable before it is provisioned; a half-created install is never left
  behind; provision progress is always visible, never a spinner with no end
- **Shape** — wizard in `DashboardShell`
- **Status** — page stub (`/:orgId/installs/setup`), wizard not built
- **Spec** — none

**Decided:** this flow is **hand-held to the end.** The install's product pages
are gated until provisioned; opening one during provision redirects back to
`/installs/setup?installId=`. The install group assignment is **optional** and
its exact shape is open.

---

# Understanding state and change

These two are the same pair of questions from DESIGN.md "The lens", and they
apply to **both installs and apps**. They stay separate flows because they want
different shapes — a graph versus a timeline — and merging them loses one.

## 4. Check status

- **Who** — operator (install) or platform engineer (app)
- **Job** — When something looks wrong, I want to see the state of this install
  or app at a glance, so I can tell what is broken and where.
- **Trigger** — a support request, an alert, or a routine check
- **Path** — open the install or app → **read the graph** → spot the component
  or runner in a bad state → **click the node to open its details panel** →
  **click through to the deploy or workflow that caused it**
- **Done** — the user knows whether it is healthy, and if not, has reached the
  run that broke it
- **Must not break** — a failed component is never hidden behind a healthy
  top-level status; a bad node is identifiable without reading every label; the
  graph stays on screen while a node's panel is open; every failing node leads
  somewhere — a bad state with no run to open is a dead end
- **Shape** — page holding the graph. Node → **panel**. Run → **page**. The
  canonical Discovery chain (DESIGN.md §2)
- **Status** — partial. Overview cards are built; **the graph is not**, so the
  drill-down starts from cards rather than topology today
- **Spec** — none

## 5. Check when changes happened

- **Who** — operator (install) or platform engineer (app)
- **Job** — When I need to know when a change landed, I want to see the history
  of changes to this install or app config, so I can tell what happened and in
  what order.
- **Trigger** — "when was this deployed?", "did the config change go through?",
  "what has happened this week?"
- **Path** — open the activity view → scan the timeline → open any run for
  detail
- **Done** — the user can see changes in time order and open any of them
- **Must not break** — the timeline is ordered and complete, with no silently
  dropped events; a failed run is visually distinct from a successful one
  without reading the label; the time of a change is always visible, not only
  on hover
- **Shape** — page with a timeline; runs open as pages
- **Status** — built (`/:orgId/installs/:installId/activity`,
  `/:orgId/apps/:appId/branches/:branchId/activity`)
- **Spec** — none

---

# Day 2 operations

Running things against an install that already exists. Distinct from the
blessed path — these do not change what the app *is*, they carry out an
operation against a running install.

Both flows have the same spine: **trigger → watch the workflow → read logs and
outputs.** Both therefore need the run detail page and the logs organism, and
both hit DESIGN.md's open question about navigation inside a run detail page.

## 6. Run an action on an install

- **Who** — operator
- **Job** — When a customer needs a specific operation run, I want to run the
  action against their install and watch it, so I can tell them it is done.
- **Trigger** — a support request, or routine maintenance
- **Path** — find the action on the install → supply inputs if it takes any →
  trigger the run → **watch the workflow** → read its logs and outputs
- **Done** — the run finished and the operator has seen its logs and outputs
- **Must not break** — an action with required inputs cannot be triggered
  without them; the run is always reachable after triggering, never a
  fire-and-forget with no link; logs and outputs stay available after the run
  finishes, not only while it is live
- **Shape** — modal to collect inputs and trigger, then the run's **page** for
  the workflow, logs and outputs
- **Status** — not built
- **Spec** — none

## 7. Run a runbook on an install

- **Who** — operator
- **Job** — When I need to carry out a documented procedure, I want to run the
  runbook against an install, so its actions happen in order and I can follow
  the documentation while they do.
- **Trigger** — a documented operational procedure — onboarding steps, a
  recovery routine, a migration
- **Path** — open the runbook → **read its readme** → trigger the run → watch
  the workflow step through its actions → read logs and outputs per step
- **Done** — the runbook run finished and the operator can see the result of
  each action in it
- **Must not break** — the readme is readable before *and during* the run, since
  it is the instructions for what is happening; a failure is attributable to the
  specific action that failed, not just to the runbook; each step's logs and
  outputs are reachable individually
- **Shape** — the runbook itself is a **page** (readme plus its actions); a
  modal triggers the run; the run is a **page** with per-step detail
- **Status** — not built
- **Spec** — none

**A runbook is a collection of actions plus a readme.** The readme is
documentation for a human carrying out the procedure, which makes it part of the
flow rather than decoration — it needs the markdown organism, also unbuilt.

---

# Reviewing change

## 8. Approve a change

- **Who** — platform engineer
- **Job** — When a plan is waiting on me, I want to see exactly what will change
  before it applies, so I do not break a customer's infrastructure.
- **Trigger** — a pending approval on an app or install workflow, found without
  knowing where to look
- **Path** — notice something is waiting → open the run → **read the diff** →
  approve, deny, or approve all
- **Done** — the plan is approved or denied, and the run continues or stops
- **Must not break** — the diff is readable before the decision is possible;
  approval copy names the *kind* of change (Terraform plan, Helm chart,
  Kubernetes manifest, install creation), never a generic "approval required";
  the decision is reflected immediately, not after a refetch
- **Shape** — org-wide pending signal → banner on the run → modal to decide.
  All five parts, see DESIGN.md §4 "Approvals"
- **Status** — **diff machinery built, approval chain not.** `Diff`,
  `DiffSection`, `DiffFilter` and the per-engine diffs exist; the pending
  signal, banner and decision modals do not
- **Spec** — none

This is the gate in the blessed path (DESIGN.md "App branches are the blessed
path"). A git push is the trigger, but rollout waits here until a human decides
— which is why this flow is load-bearing rather than informational.

---

# Coverage

Where Lite actually is, so nobody reads this file as a description of the
present:

| Area | State |
|---|---|
| Navigation, shells, breadcrumbs, status bar | built |
| Apps and installs lists (search, filter, pagination, card view) | built |
| Install and branch overviews and activity timelines | built |
| Branch config and plan diffs | built |
| Onboarding, app setup, install setup wizards | page stubs |
| Install and app topology graphs | not built |
| Run detail pages, logs, markdown | not built |
| Actions and runbooks | not built |
| The approval chain around the diffs | not built |
| Everything under org settings | scaffolds |
| Customer-facing surfaces | none, by design |

The gap is concentrated in **templates** — `list page`, `detail page`, `wizard`
and `dashboard` are all unbuilt (see `.planning/lite/NOTES.md`), and most of
these flows need one.

## Writing the step docs

When a flow gets built, add its step doc to `e2e/flows-lite/` and link it from
that flow's **Spec** line. The lite suite is separate from the production one at
every level:

| Suite | Config | testDir | Target |
|---|---|---|---|
| Lite app | `e2e/lite.config.ts` | `e2e/specs-lite/` | :4000 with `nuon_dashboard_lite` on |
| Lite Ladle | `e2e/lite-ladle.config.ts` | `e2e/specs-lite-ladle/` | :62002 |

`global-setup.ts`, `global-teardown.ts`, `env.ts`, `fixtures.ts` and
`helpers.ts` are shared with the production suite — token and org seeding is
identical and should not fork.

**The BFF serves one shell.** `DashboardLite` is a server config flag read at
startup, so the production and lite app suites cannot run against the same
server instance. The lite app suite asserts the served shell is lite as its
first step and fails with a clear message rather than dying on selectors.
