# Dashboard UI Copy Style Guide

This guide defines the voice, tone, and writing patterns for all user-facing text in the dashboard UI. It is derived from the existing codebase and should be followed by all contributors — human and AI.

---

# Foundations

Rules that apply to all UI copy, regardless of surface.

## Voice

**Direct, confident, and calm.** Nuon's UI speaks like a competent teammate — not a marketer, not a robot, not a customer support script. The voice is:

- **Plain-spoken** — Use normal words. Say "remove" not "eliminate". Say "start" not "initiate".
- **Technical when appropriate** — Users are engineers. Don't dumb down domain terms (deploy, provision, teardown, drift scan). Don't explain what a webhook is.
- **Brief** — Every word should earn its place. If you can cut a word without losing meaning, cut it.
- **Neutral** — Don't celebrate ("Great job!"), don't apologize excessively ("We're so sorry!"), don't use exclamation marks. State what happened and what to do next.

## Tone by context

| Context | Tone | Example |
|---------|------|---------|
| Neutral actions (create, edit) | Matter-of-fact | "Create webhook" |
| Destructive actions (delete, deprovision) | Calm but serious | "This action will remove the install and cannot be undone." |
| Errors | Honest, no blame | heading: "Loading failed" / description: "This is usually temporary. Try refreshing the page." |
| Empty states | Helpful, forward-looking | "No workflows found. Activity will appear here once the runner starts processing jobs." |
| Success confirmations | Understated | "Plan approved" / "Building sandbox" |
| Warnings | Clear, factual | "Force unlocking a workspace that is actively in use may cause state corruption." |

## Capitalization

**Sentence case everywhere.** Capitalize the first word and proper nouns only. This applies to all UI text: headings, buttons, labels, tabs, empty states, tooltips, toasts.

```
"Create your org"          not  "Create Your Org"
"Connect a cloud account"  not  "Connect A Cloud Account"
"No webhooks configured"   not  "No Webhooks Configured"
```

**Exceptions:** Proper nouns (AWS, Nuon, Terraform, GitHub, Slack) and acronyms (API, CLI, VCS, URL).

## Pronouns & possessives

**Use "your" sparingly — only for account-level possessions.** "Your org", "your team", "your account" are fine. For resources (installs, components, webhooks, runners), use the entity name or "this [thing]".

```
"Removing {email} will revoke their access to your org."     ← "your org" is correct
"Deprovisioning {installName} will remove all resources."    ← not "your install"
"This webhook will stop receiving events."                    ← not "your webhook"
```

**Why:** Users often manage resources they don't personally own. "Your install" is awkward when an admin is managing a customer's deployment. The entity name is always unambiguous.

**Exception:** Onboarding copy and first-run experiences can use "your" more freely ("Create your first app") because the user is always acting on their own behalf.

## Pluralization & counts

### Use counts when the number matters

When the user took a bulk action or needs to know a quantity, show the count:

```
"3 workflows canceled"
"Pruned 12 old tokens"
"1 workflow canceled"              ← not "1 workflows canceled"
```

### Use "all" when the user selected everything

```
"All selected workflows were canceled."
"All plans approved."
```

### Pluralization pattern

Use a ternary for simple cases. Keep the singular/plural logic next to the number:

```tsx
`${count} workflow${count === 1 ? '' : 's'} canceled`
```

For more complex copy, write both variants:
```tsx
count === 1
  ? `Pruned 1 old token.`
  : `Pruned ${count} old tokens.`
```

**Don't** skip pluralization — "3 workflow canceled" reads as a bug.

---

# UI patterns

Reference for specific surfaces. Find the section that matches what you're building.

## Buttons & actions

### Pattern: verb + object

Buttons describe what they do. Lead with an action verb, follow with the object.

```
"Build component"       "Create webhook"        "Run action"
"Deploy build"          "Invite team member"     "Remove user"
"Cancel workflow"       "Shutdown process"       "Sync secrets"
"Deprovision install"   "Teardown component"     "Approve all"
```

**Don't** use vague labels like "Submit", "Confirm", "OK", or "Continue" when a specific verb is available.

**Don't** add "please" or other pleasantries to button labels.

### Loading states

When a button action is in progress, switch to the gerund form (verb + "-ing"):

```
"Build component"    →  "Building component"
"Deploy build"       →  "Deploying build"
"Remove user"        →  "Removing user"
"Cancel workflow"    →  "Canceling workflow"
"Shutdown process"   →  "Shutting down process"
```

For short generic actions, use the gerund + ellipsis:

```
"Create"    →  "Creating..."
"Save"      →  "Saving..."
"Delete"    →  "Deleting..."
"Invite"    →  "Inviting..."
```

### Danger buttons

Use `variant="danger"` for destructive actions. The label should name the destructive action plainly — don't soften it.

```
"Delete webhook"          not  "Remove this webhook"
"Deprovision install"     not  "Are you sure?"
"Remove user"             not  "Revoke access"
```

### Disabled states

When a button or action is disabled, **always explain why** via a tooltip. The user should never have to guess.

**Pattern: "Cannot [action] — [reason]"**

```
"Cannot deploy — no successful builds yet"
"Cannot teardown — deploy in progress"
"Cannot remove — you are the only admin"
"Cannot run — action requires at least one step"
```

Keep it to one line. The reason should suggest what to do, or at least name the blocker.

**Don't** just disable a button silently. **Don't** use "This action is currently unavailable" — say *why*.

## Modals

### When to use a modal vs just do it

Not every action needs a confirmation modal. Use this rule:

- **No modal (just do it + toast):** The action is instantly reversible or trivially low-risk. Toggling a UI setting, reordering a list, copying a value. Just perform the action and show a toast.
- **Modal:** The action is irreversible, takes more than a few seconds to undo, or affects infrastructure. Everything below.

If you're unsure, err toward a modal — it's less disruptive than losing infrastructure.

### Destructive action severity tiers

Every destructive action falls into one of three tiers. Pick the right tier based on blast radius and reversibility.

#### Tier 1 — Simple confirm

**When:** Reversible or low-impact. Canceling a workflow, skipping a step, removing a channel subscription.

**Modal structure:** Consequence statement only. No warning banner, no type-to-confirm.

```
heading: "Cancel workflow?"
body: "Canceling this workflow will stop all in-progress steps. You will need to trigger a new workflow."
button: "Cancel workflow" (variant="danger")
```

#### Tier 2 — Warning banner

**When:** Significant but recoverable. Reprovisioning an install, shutting down a runner, disabling config sync, removing a VCS connection.

**Modal structure:** Consequence statement + `<strong>Warning:</strong>` banner explaining the specific risk.

```
heading: "Shutdown runner process?"
body: "Shutting down this runner will restart the process."
warning: "Causes all jobs to queue while the process restarts."
button: "Shutdown process" (variant="danger")
```

#### Tier 3 — Type to confirm

**When:** Irreversible or high-blast-radius. Deprovisioning an install, forgetting an install, deleting all data, removing a user.

**Modal structure:** Consequence statement + warning banner + "To verify, type [entity name] below." The button stays disabled until the input matches.

```
heading: "Deprovision install?"
body: "Deprovisioning {installName} will remove all resources from the cloud account."
warning: "This action cannot be undone."
verify: "To verify, type {installName} below."
button: "Deprovision install" (variant="danger", disabled until input matches)
```

### Modal headings

**Confirmation modals (destructive)** — Use a question ending with `?`:

```
"Delete webhook?"
"Remove team member?"
"Cancel build?"
"Shutdown runner process?"
"Deprovision install?"
```

**Action modals (constructive)** — Use a verb + object statement (no question mark):

```
"Create webhook"
"Edit install"
"Invite team member"
"Subscribe a channel"
```

**Icons & themes:**

| Action type | Theme | Icon pattern |
|------------|-------|-------------|
| Constructive (create, edit, connect) | `info` | Domain-specific icon (WebhooksLogoIcon, SlackLogoIcon) |
| Cautionary (shutdown, reprovision) | `warn` | PowerIcon, ArrowURightUpIcon |
| Destructive (delete, deprovision, remove) | `error` | WarningIcon or TrashIcon |

### Modal body copy

Confirmation modals lead with the consequence — not "Are you sure?". The modal heading, danger theme, and warning icon already signal that this is a confirmation. Restating the question wastes the user's time. Get to the point: what will happen.

Structure (not all parts required — use only what's needed, based on the severity tier):

1. **Consequence** — One sentence explaining what will happen. Lead with this.
2. **Warning** (tier 2+) — Bolded prefix: `<strong>Warning:</strong>` followed by the specific risk.
3. **Verification** (tier 3 only) — "To verify, type [entity name] below."

```tsx
// Tier 1: Cancel workflow (heading: "Cancel workflow?")
"Canceling this workflow will stop all in-progress steps. You will need to trigger a new workflow."

// Tier 2: Reprovision (heading: "Reprovision install?")
"Reprovisioning {installName} will recreate all resources and redeploy all components."
<strong>Warning:</strong> "This will cause downtime while resources are recreated."

// Tier 3: Forget install (heading: "Forget {installName}?")
"This will permanently remove {installName} from the dashboard."
<strong>Warning:</strong> "This should only be used when an install was broken in an
unusual way and needs to be manually removed."
"To verify, type {installName} below."
```

**Don't** open with "Are you sure you want to...?" — the heading already asks the question. Start with what happens.

> **Migration note:** Many existing modals still use "Are you sure you want to..." phrasing. When touching these components, update them to the consequence-first pattern. Don't rewrite them all at once.

#### Consequence statements

State the outcome plainly. One sentence. Don't hedge with "might" or "could" when the outcome is certain.

```
"This webhook will stop receiving workflow lifecycle events."
"Removing this user will revoke their access immediately."
"Lifecycle events will stop posting to this channel."
"Once a workflow is canceled, it cannot be restarted."
```

#### Warning labels

Use bold labels for callouts. Three levels:

- `<strong>Warning:</strong>` — Risk of data loss, state corruption, or irreversibility
- `<strong>Important:</strong>` — Required follow-up action or changed behavior
- `<strong>Note:</strong>` — Helpful context, no risk

#### Bulleted impacts

When an action has multiple effects, use a lead-in sentence followed by bullets:

```
"This will create a workflow that attempts to:"
• "Teardown each install component according to the dependency order"
• "Teardown the install sandbox"
```

## Toasts

Every toast has a **heading** and a **description** (children). The heading says *what happened* in generic terms. The description adds *specific context* — which entity, what to expect, how long it might take.

### Async actions (kicked off a background job)

Use `theme="info"`. Heading in present progressive (the job is now running). Description names the specific entities and sets expectations.

```tsx
<Toast heading="Deploying component" theme="info">
  <Text>Deploying {component.name} to {install.name}. This may take a few minutes.</Text>
</Toast>

<Toast heading="Building component" theme="info">
  <Text>Building {component.name}. This may take a few minutes.</Text>
</Toast>

<Toast heading="Reprovisioning install" theme="info">
  <Text>Reprovisioning {install.name}. All components will be redeployed.</Text>
</Toast>

<Toast heading="Shutting down runner" theme="info">
  <Text>Shutting down {process.name}. Jobs will queue until restart.</Text>
</Toast>

<Toast heading="Scanning for drift" theme="info">
  <Text>Scanning {component.name} on {install.name} for drift.</Text>
</Toast>
```

### Instant completions (action finished immediately)

Use `theme="success"`. Heading in past tense. Description confirms what was affected.

```tsx
<Toast heading="Plan approved" theme="success">
  <Text>Approved changes for {component.name} on {install.name}.</Text>
</Toast>

<Toast heading="Step skipped" theme="success">
  <Text>{step.name} was skipped. The workflow will continue.</Text>
</Toast>

<Toast heading="Workflow canceled" theme="success">
  <Text>Canceled the {workflow.type} workflow.</Text>
</Toast>

<Toast heading="Connection removed" theme="success">
  <Text>GitHub connection {name} has been removed.</Text>
</Toast>

<Toast heading="Branch created" theme="success">
  <Text>Created app branch: {branch.name}.</Text>
</Toast>
```

### Error toasts

Use `theme="error"`. Heading: always **"[thing] failed"**. Description: the API error message or a fallback explanation.

```tsx
<Toast heading="Build failed" theme="error">
  <Text>{err?.error || 'Unable to start the build.'}</Text>
</Toast>

<Toast heading="Step retry failed" theme="error">
  <Text>{err?.error || 'The step was not queued for retry. Try again later.'}</Text>
</Toast>

<Toast heading="Workflow cancellation failed" theme="error">
  <Text>{err?.error || 'An unknown error occurred.'}</Text>
</Toast>
```

### Toast heading patterns

| Action type | Heading tense | Examples |
|------------|---------------|---------|
| Async job started | Present progressive | "Deploying component", "Building sandbox", "Shutting down runner" |
| Instant completion | Past tense | "Plan approved", "Branch created", "Workspace unlinked" |
| Error | "[thing] failed" | "Build failed", "Connection failed", "Workflow cancellation failed" |

### Rules

- **Heading is always a plain string** — no JSX, no Badge components, no inline markup.
- **Description is always a `<Text>` child** — this is where entity names and context go.
- **Don't say "successfully"** — the success theme already communicates that.
- **Don't put API error details in the heading** — keep headings scannable, put error messages in the description.
- **Duration hints** — add "This may take a few minutes." for builds, deploys, and provisions. Skip it for fast operations.

## Error messages

### One pattern: "[thing] failed"

Use **"[thing] failed"** as the primary error framing everywhere — toast headings, error state headings, inline errors. This is the shortest form and scans fastest.

```
"Build failed"
"Connection removal failed"
"Deploy failed"
"Branch creation failed"
```

For longer error state descriptions (not headings), you can expand with "Unable to [action]" as a follow-up sentence:

```
heading: "Deploy failed"
description: "Unable to deploy {component.name} to {install.name}. This is usually temporary."

heading: "Connection failed"
description: "Unable to connect to the API. This is usually temporary."
```

**Don't** mix patterns. Avoid these alternatives:
- "Failed to [action]" — wordier than "[thing] failed" and less scannable
- "Unable to [action]" as a heading — save this for description text only
- "Could not" / "Couldn't" / "We were unable to" — hedging language

**Don't** say "Oops!", "Uh oh!", or use humor in error states.

### Error descriptions

Follow the heading with a calm, factual explanation. One sentence. If the API returned an error message, show it. If not, use a generic fallback.

```
"Unable to start the build. Check the logs for details."
"This is usually temporary. Try refreshing the page."
"{err.error}"  ← API error message, shown as-is
```

### Validation errors

State what's wrong. No "please".

```
"Branch name is required"
"Repository is required when using VCS"
"Email doesn't match"
```

## Empty states

Empty states have two parts: a title and a message. **Every empty state message must end with a next step** — tell the user what will make things appear or what they can do. Never leave the user staring at a dead end.

### Title pattern: "No [things] yet" or "No [things] found"

- Use **"yet"** when the user hasn't created the resource: `"No apps yet"`, `"No runs yet"`
- Use **"found"** when filters/search returned nothing: `"No matching reports"`, `"No workflows found"`
- Use **"configured"** for settings that haven't been set up: `"No webhooks configured"`, `"No policies configured"`

### Message pattern: explain what will make things appear

Always end with a forward-looking statement — either what the user can do, or what will trigger content to show up.

```
"Activity will appear here once the runner starts processing jobs."
"This action has not been run yet. Trigger a run to see history here."
"Evaluations will appear here once a deploy or sandbox run triggers a policy check."
"Logs will appear here as soon as the runner starts streaming them."
"Create a webhook to receive workflow lifecycle events from this org."
```

**Don't** write passive descriptions of emptiness:
```
"No runner processes are currently active or offline."     ← dead end, no next step
"There are no workflows to display."                       ← just restates the title
"It looks like there's nothing here!"                      ← filler
```

Instead, rewrite with a next step:
```
"No active processes. Processes will appear here when a runner connects."
"No workflows to display. Workflows run automatically when you deploy or teardown components."
```

## Page headings & descriptions

Page headings name the resource type. Descriptions are one sentence explaining what the user can do here.

```
title: "Installs"
description: "View and manage deployments of your app into customer cloud accounts."

title: "Team"
description: "Manage your team members and permissions."

title: "Workflows"
description: "View past and active workflows for this install."
```

**Don't** start descriptions with "This page..." or "Here you can...". Start with a verb.

## Form labels & help text

### Labels

Sentence case. Be specific enough that the label works without context.

```
"Branch name"                  not  "Name"
"Email address"                not  "Email"
"Signing secret (optional)"   not  "Secret"
"Path filter (optional)"      not  "Filter"
```

Mark optional fields with "(optional)" in the label. Required is the default — don't mark it.

### Placeholder text

Use example values, not instructions. Placeholders disappear on focus — they're hints, not labels.

```
placeholder="production"               not  "Enter branch name here"
placeholder="^(src/|config/).*"        not  "Enter regex pattern"
placeholder="Search by name or ID..."  ← search fields are the exception
```

### Help text (subtext)

One sentence below the input. Explains constraints or gives context the label can't.

```
"Must be an absolute http or https URL."
"The secret cannot be retrieved later. Edit the webhook to rotate it."
"Regex pattern to filter which file changes trigger workflow runs."
"Path to the application config (use "." for root)."
```

**Don't** repeat the label in the help text. Don't start with "This field..." or "Enter the...".

## Contextual help & feature explanations

Short explanatory copy that appears near a feature toggle, setting, or unfamiliar concept. This isn't help docs — it's a sentence or two that helps the user decide whether to act.

### Pattern: what it does + what changes

Lead with what the feature does, then state the concrete effect of enabling/disabling it.

```
"Config sync pulls settings from the install config file on every deploy."
"When auto approve is enabled, all changes will be applied without manual review."
"Drift detection compares the actual state of your infrastructure against the expected state."
```

### Rules

- **One to two sentences max.** If it needs more, link to docs.
- **Start with what, not why.** Users can decide "why" themselves — just tell them what the thing does.
- **Use the same terms as the UI.** If the toggle says "Config sync", the explanation says "Config sync", not "configuration synchronization".
- **Don't use "allows you to" or "enables you to"** — just state what happens. ("Drift detection compares..." not "Drift detection allows you to compare...")
- **Don't over-explain to engineers.** If the concept is standard (webhooks, git branches, API keys), a one-liner is enough. If it's Nuon-specific (install configs, sandboxes, component teardown), give slightly more context.

## Status text

Status values come from the API as kebab-case strings (e.g., `not-provisioned`) and are auto-converted to sentence case by the Status component. Don't manually format status strings.

For composite statuses or custom display:
```
"Awaiting approval"
"Failed — awaiting retry"
"Auto-approved (policies)"
"Not attempted"
```

Use an em dash `—` to join status with qualifier, not a hyphen.

## Navigation & tabs

Tab labels and nav items use sentence case, one or two words maximum.

```
"Summary"    "Logs"       "Trace"      "Outputs"
"Plan"       "State"      "Variables"  "Components"
"Actions"    "Roles"      "Policies"   "Workflows"
```

## Table column headers

Sentence case. Keep them short — ideally one or two words.

```
"App name"    "Status"     "Created"    "Uptime"
"Email"       "Role"       "Channel"    "Version"
"URL"         "Type"       "ID"         "Started"
```

---

# Reference

## Word list

Preferred terms — use these consistently instead of alternatives.

| Use | Don't use |
|-----|-----------|
| install | deployment, instance |
| deprovision | delete (for infrastructure) |
| teardown | destroy, remove (for components) |
| remove | delete (for users, subscriptions) |
| delete | remove (for webhooks, data) |
| sandbox | dev environment, staging |
| runner | agent, executor |
| workflow | pipeline, process (for orchestration) |
| build | compile, package |
| deploy | ship, push, release |
| drift scan | drift check, drift detection |
| component | service, module |
| reprovision | recreate, rebuild (for infrastructure) |
| action | task, job (for user-triggered operations) |
| org | organization (in UI copy — use the short form) |
| cannot | can not (one word, always) |

## Things to never do

- **No exclamation marks** in UI copy. Ever.
- **No emoji** in UI copy.
- **No "please"** in buttons or error messages. ("Remove user" not "Please remove user")
- **No "successfully"** — if the toast shows, it succeeded. ("Plan approved" not "Plan successfully approved")
- **No title case** — not in headings, not in buttons, not anywhere.
- **No "click here"** or "click the button" — the UI should be self-evident.
- **No gendered language** — use "they/them" if pronouns are needed (rare in UI copy).
- **No Oxford comma debates** — use the Oxford comma when listing three or more items.
