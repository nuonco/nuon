# Copy guidelines

Voice, tone and writing patterns for every piece of user-facing text in
`client/lite/`. Read this before writing a label, heading, button, empty state,
error, toast or help line.

This is Lite's own guide. The production dashboard has its own
(`services/dashboard-ui/COPY_STYLE.md`) which shares most of this voice, but the
component APIs and several patterns differ — follow this one for anything in
`client/lite/`.

Visual guidance is in [STYLES.md](./STYLES.md), pattern guidance in
[DESIGN.md](./DESIGN.md), engineering conventions in [DEV.md](./DEV.md).

---

# Foundations

## Voice

**Direct, confident and calm.** Lite speaks like a competent teammate — not a
marketer, not a robot, not a support script.

- **Plain-spoken** — normal words. "Remove", not "eliminate". "Start", not
  "initiate".
- **Technical when appropriate** — users are engineers. Don't dumb down domain
  terms (deploy, provision, teardown, drift scan) and don't explain what a
  webhook is.
- **Brief** — every word earns its place. If cutting a word loses no meaning,
  cut it.
- **Neutral** — don't celebrate, don't over-apologise, don't exclaim. State what
  happened and what to do next.

## Tone by context

| Context | Tone | Example |
|---|---|---|
| Neutral actions | Matter-of-fact | "Create webhook" |
| Destructive actions | Calm but serious | "Deprovisioning {name} will remove all resources from the cloud account." |
| Errors | Honest, no blame | "Deploy failed" / "This is usually temporary. Try again." |
| Empty states | Helpful, forward-looking | "No workflows yet. Activity will appear here once the runner starts processing jobs." |
| Success | Understated | "Plan approved" |
| Warnings | Clear, factual | "Force unlocking a workspace that is actively in use may cause state corruption." |

## Capitalization

**Sentence case everywhere.** Capitalize the first word and proper nouns only —
headings, buttons, labels, tabs, empty states, tooltips, toasts, table headers,
nav items.

```
"Create your org"          not  "Create Your Org"
"No webhooks configured"   not  "No Webhooks Configured"
"API tokens"               not  "Api Tokens"
```

Exceptions are proper nouns (AWS, Nuon, Terraform, GitHub, Slack) and acronyms
(API, CLI, VCS, URL, OIDC).

## What this guide governs

This guide owns **copy we write**. Two other kinds of string appear in the UI
and are not ours to style:

- **API vocabulary** — statuses, types, operations. `Status` humanizes these
  automatically (kebab-case → sentence case). **Never hand-write or re-case a
  status label**; pass the status and let the component format it.
- **Identifiers** — resource names, IDs, regions, Kubernetes kinds, commit SHAs.
  These render verbatim, usually in mono. Never re-case or truncate them in copy.

## Pronouns and possessives

**Use "your" only for account-level possessions** — "your org", "your team",
"your account". For resources, use the entity name or "this {thing}".

```
"Removing {email} will revoke their access to your org."      ← correct
"Deprovisioning {installName} will remove all resources."     ← not "your install"
"This webhook will stop receiving events."                     ← not "your webhook"
```

Users often manage resources they don't personally own — "your install" is wrong
when an admin manages a customer's deployment. The entity name is always
unambiguous.

Onboarding is the exception: "Create your first app" is fine, because the user is
always acting on their own behalf there.

Never gendered language. Use they/them if a pronoun is needed.

## Counts and pluralization

Show a count when the number matters, and keep singular/plural logic next to the
number:

```tsx
`${count} component${count === 1 ? '' : 's'} deployed`
```

"3 component deployed" reads as a bug. Don't skip it.

**Lite has no multi-select** (see DESIGN.md), so there is no "N selected" copy.
Whole-set actions name their scope instead:

```
"Deploy all components"
"Deploying all components. This may take a few minutes."
```

---

# UI patterns

## Buttons and actions

### Verb + object

```
"Build component"      "Create webhook"       "Run action"
"Deploy build"         "Invite team member"   "Remove user"
"Cancel workflow"      "Deprovision install"  "Approve all"
```

Never vague labels — no "Submit", "Confirm", "OK", or "Continue" when a specific
verb exists. Never "please".

### Loading

`Button` takes a `loading` prop that renders the spinner, but the **label still
changes** — switch to the gerund:

```
"Build component"  →  "Building component"
"Remove user"      →  "Removing user"
"Create"           →  "Creating..."
"Save"             →  "Saving..."
```

Short generic verbs take the gerund plus an ellipsis; specific verb + object
phrases don't need one.

### Danger

`variant="danger"` names the destructive action plainly. Don't soften it.

```
"Delete webhook"        not  "Remove this webhook"
"Deprovision install"   not  "Are you sure?"
```

### Disabled reasons

A disabled action explains itself through `Button`'s own `tooltip` prop — never a
hand-wrapped `Tooltip`, never a `title` attribute.

**Pattern: "Cannot {action} — {reason}"**

```
"Cannot deploy — no successful builds yet"
"Cannot teardown — deploy in progress"
"Cannot remove — you are the only admin"
```

One line. The reason names the blocker or suggests the fix. Never
"This action is currently unavailable" — say *why*.

No tooltip is needed when the reason is already obvious: the label is mid-async
("Saving…"), a field shows its own validation error, or it's a pagination
control at a boundary.

## Modals

### Modal or just do it?

- **No modal** — the action is instantly reversible or trivially low-risk.
  Toggling a preference, copying a value, switching table view. Do it and show a
  toast, or say nothing at all.
- **Modal** — irreversible, slow to undo, or touches infrastructure.

When unsure, use the modal. It is less disruptive than losing infrastructure.

### Severity tiers

**Tier 1 — simple confirm.** Reversible or low impact: cancelling a workflow,
skipping a step, removing a channel subscription. Consequence sentence only.

```
heading: "Cancel workflow?"
body:    "Canceling this workflow will stop all in-progress steps. You will need to trigger a new workflow."
primaryAction: "Cancel workflow" (variant="danger")
```

**Tier 2 — warning callout.** Significant but recoverable: reprovisioning an
install, shutting down a runner, removing a VCS connection. Consequence plus a
bolded **Warning:** line naming the specific risk.

**Tier 3 — type to confirm.** Irreversible or high blast radius: deprovisioning
an install, forgetting an install, removing a user. Consequence, warning, then
"To verify, type {name} below." The primary action stays disabled until the input
matches.

> **Not built yet.** Lite has no `Banner` and no type-to-confirm input, so tiers
> 2 and 3 have no implementation. Write the copy to these shapes, and raise the
> missing components rather than improvising a one-off.

### Headings

- **Destructive confirmations** — a question: "Delete webhook?",
  "Deprovision install?", "Remove team member?"
- **Constructive actions** — verb + object, no question mark: "Create webhook",
  "Edit install", "Invite team member"

### Body copy

**Lead with the consequence, never "Are you sure?"** The heading, theme and icon
already signal that this is a confirmation; restating the question wastes a line.

```
"This webhook will stop receiving workflow lifecycle events."
"Removing this user will revoke their access immediately."
"Once a workflow is canceled, it cannot be restarted."
```

One sentence. Don't hedge with "might" or "could" when the outcome is certain.

Callout labels, in order of severity:

- **Warning:** — data loss, state corruption, irreversibility
- **Important:** — required follow-up or changed behaviour
- **Note:** — helpful context, no risk

For multiple effects, use a lead-in plus bullets:

```
"This will create a workflow that attempts to:"
• "Teardown each install component according to the dependency order"
• "Teardown the install sandbox"
```

## Toasts

Every toast is a **heading** plus a **description**. The heading says what
happened in generic terms; the description carries the specifics — which entity,
what to expect, how long.

| Situation | Heading tense | Theme |
|---|---|---|
| Async job started | Present progressive | `info` |
| Instant completion | Past tense | `success` |
| Failure | "{thing} failed" | `error` |

```tsx
<Toast heading="Deploying component" theme="info">
  <Text>Deploying {component.name} to {install.name}. This may take a few minutes.</Text>
</Toast>

<Toast heading="Plan approved" theme="success">
  <Text>Approved changes for {component.name} on {install.name}.</Text>
</Toast>

<Toast heading="Build failed" theme="error">
  <Text>{err?.error || 'Unable to start the build.'}</Text>
</Toast>
```

**Rules:**

- **Heading is a plain string** — no JSX, no badges, no markup. `IToast` types it
  as `string` for this reason.
- **Description is a `<Text>` child** — entity names and context live here.
- **Never "successfully".** The success theme says it.
- **Never API error text in the heading.** Headings stay scannable.
- Add "This may take a few minutes." for builds, deploys and provisions. Skip it
  for fast operations.
- An action inside a toast uses `actionLabel` and follows verb + object.
- **Never a toast for something the user must act on** — that is a modal or a
  banner.

## Errors

### One pattern: "{thing} failed"

Use it everywhere — toast headings, error headings, inline errors. It is the
shortest form and scans fastest.

```
"Build failed"      "Deploy failed"      "Branch creation failed"
```

"Unable to {action}" is for **description** text only, never a heading:

```
heading:     "Deploy failed"
description: "Unable to deploy {component.name} to {install.name}. This is usually temporary."
```

Don't mix in "Failed to {action}", "Could not", "Couldn't", or "We were unable
to". No "Oops!", "Uh oh!", or humour.

### Form errors

Submission failures go in `FormErrorBanner` inside the form — never a toast. A
toast disappears while the broken field stays broken.

Validation errors state what is wrong, with no "please":

```
"Branch name is required"
"Repository is required when using VCS"
"Email doesn't match"
```

### CompositeError text is not ours

When the API returns a `composite_error`, **its text is the error message.** Do
not paraphrase it, do not truncate its sections, and do not wrap it in a heading
of your own invention — the API already sent `message`, `type` and section
headings. The only copy you write around it is a fallback for when it is absent.

## Empty and failed collections

In Lite, `Table` and `Timeline` take a single `emptyState`, and it carries
**both** the empty and the failed message. Which one you pass depends on whether
the request failed.

```tsx
emptyState={error ? 'Installs failed to load' : 'No installs yet'}
```

Showing "No installs yet" when the request failed is a lie. Always branch.

### Empty titles

- **"No {things} yet"** — the user hasn't created any: "No apps yet"
- **"No {things} found"** — filters or search returned nothing: "No workflows found"
- **"No {things} configured"** — settings not set up: "No webhooks configured"

### Empty messages end with a next step

Every empty message says what will make things appear, or what the user can do.
Never a dead end.

```
"Activity will appear here once the runner starts processing jobs."
"Trigger a run to see history here."
"Create a webhook to receive workflow lifecycle events from this org."
```

Not this:

```
"There are no workflows to display."          ← restates the title
"It looks like there's nothing here!"          ← filler, and an exclamation mark
```

### Failed messages

Use the error pattern, and say it is retryable if it is:

```
"Installs failed to load"
"Workflows failed to load. This is usually temporary."
```

## Drafts

The resume prompt names both choices as verbs, and states the draft's age:

```
heading:         "Resume draft"
body:            "You have unsaved changes from {age}. Resume your draft or start fresh?"
primaryAction:   "Resume draft"
secondaryAction: "Start fresh"
```

Draft age renders through `Time` with `format="relative"` — never a hand-written
"2 hours ago".

## Approvals

Approval copy is **specific to the approval type**. Generic "Approval required"
is not acceptable — the user needs to know what kind of change they are blessing.

```
"Terraform plan requires review"
  "This Terraform plan is ready for your review. Inspect the proposed infrastructure changes before applying them."

"Kubernetes manifest requires review"
  "This Kubernetes manifest contains pending configuration changes. Review these updates before applying them to your cluster."

"New installs require review"
  "The config repo describes installs that do not exist yet. Review the proposed installs before they are created."
```

Pattern: heading is "{what} requires review"; body says what it contains and what
reviewing it achieves. Decision buttons are "Approve", "Deny", "Approve all",
and the resulting toasts are "Plan approved" / "Approval failed".

## Page headings and descriptions

The heading names the resource type. The description is one sentence starting
with a verb.

```
title:       "Installs"
description: "View and manage deployments of your app into customer cloud accounts."
```

Never start with "This page…" or "Here you can…".

## Page titles

`usePageTitle` sets the browser tab title, and the provider appends `| Nuon`.
At most two segments, most specific first.

- **{specific}** — sentence-case section name, or the entity's own name on a
  detail page. For a tab, fold the parent in: "Deploy logs".
- **{owning entity}** — the install or app name. **The org name is never a
  segment.**

```
Components | acme-app | Nuon
Deploy logs | acme-install | Nuon
Webhooks | Nuon                     ← org-level, no owner segment
```

A segment that hasn't loaded is omitted — never rendered as "undefined".

## Forms

**Labels** — sentence case, specific enough to work without context. Mark
optional fields; required is the default and goes unmarked.

```
"Branch name"                not  "Name"
"Signing secret (optional)"  not  "Secret"
```

**Placeholders** — example values, not instructions. They disappear on focus.

```
placeholder="production"                not  "Enter branch name here"
placeholder="Search by name or ID..."   ← search fields are the exception
```

**Help text** — one sentence below the input, in `caption`. Explains a constraint
the label can't.

```
"Must be an absolute http or https URL."
"The secret cannot be retrieved later. Edit the webhook to rotate it."
```

Don't repeat the label, and don't start with "This field…" or "Enter the…".

## Contextual help

Short copy near a toggle, setting or unfamiliar concept — enough to decide
whether to act, not documentation.

**Pattern: what it does, then what changes.**

```
"Config sync pulls settings from the install config file on every deploy."
"When auto approve is enabled, all changes will be applied without manual review."
```

- One to two sentences. More than that, link to docs.
- Start with what, not why.
- Use the UI's own terms — if the toggle says "Config sync", so does the help.
- Never "allows you to" or "enables you to". State what happens.
- Don't over-explain standards (webhooks, git branches, API keys). Give more
  context for Nuon-specific concepts (install configs, sandboxes, teardown).

## Status text

Status strings arrive kebab-case from the API and `Status` converts them to
sentence case. **Never format one by hand.**

For composite or custom display, join a status to its qualifier with an em dash:

```
"Awaiting approval"      "Failed — awaiting retry"
"Auto-approved"          "Not attempted"
```

## Links

Copy depends on the link's class — see [STYLES.md](./STYLES.md) and
[DESIGN.md](./DESIGN.md) for the visual side.

- **Entity link** — the resource's own name is the link text. No verb, no icon.
- **View link** — standalone, "View {resource}": "View plan", "View logs",
  "View run". "View details" only when there is no better noun. Never "See",
  "Open" or "Go to".
- **External link** — set `external`; the new-tab icon renders itself. Never type
  an icon or "(opens in new tab)" into the text.

Never "click here" or "click the button".

## Navigation, tabs and table headers

Sentence case, one or two words.

```
Nav / tabs:      "Summary"  "Logs"  "Trace"  "Components"  "Actions"  "Settings"
Table headers:   "App name"  "Status"  "Created"  "Role"  "Version"  "Type"
```

## Lite-only surfaces

These have no production equivalent, so their copy is defined here.

**Theme switcher** — the four preference values read "Light", "Dark",
"High contrast", "System". "System" means follow the OS; don't call it "Auto".

**Table view toggle** — "Table" and "Cards". Not "List"/"Grid".

**User preferences panel** — heading "Preferences". Each control gets a
sentence-case label and, where the effect isn't obvious, one `caption` line of
help. Preferences apply immediately, so **no save button and no "saved" toast.**

**Status bar (modeline)** — the densest surface in the app. Values only, no
labels, no sentences. Version renders as `v1.4.2`, connection loss as
`disconnected` in lower case, matching the modeline's mono `label` type.

---

# Reference

## Word list

| Use | Not |
|---|---|
| install | deployment, instance |
| deprovision | delete (for infrastructure) |
| teardown | destroy, remove (for components) |
| remove | delete (for users, subscriptions) |
| delete | remove (for webhooks, data) |
| sandbox | dev environment, staging |
| runner | agent, executor |
| workflow | pipeline, process |
| build | compile, package |
| deploy | ship, push, release |
| drift scan | drift check, drift detection |
| component | service, module |
| reprovision | recreate, rebuild |
| action | task, job |
| org | organization |
| branch | app branch (when context is clear) |
| cannot | can not |

## Never

- **No exclamation marks.** Ever.
- **No emoji.** Anywhere.
- **No "please"** in buttons or errors.
- **No "successfully"** — if it rendered, it succeeded.
- **No title case** — not in headings, not in buttons, not anywhere.
- **No "click here"** or "click the button".
- **No gendered language** — they/them if needed.
- **No "Are you sure?"** as modal body copy — lead with the consequence.
- **No hand-formatted statuses, timestamps or durations** — `Status`, `Time` and
  `Duration` own those.
- **Use the Oxford comma** for three or more items.
