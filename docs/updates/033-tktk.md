---
title: '033 - April 2026 Updates'
description: 'April 2026 platform, CLI, dashboard, and docs updates.'
---

_April 16, 2026_

<div className="badge badge--primary">Since v0.19.872</div>

## Azure Cloud Support (Harsh Thakur)

Added Azure management mode authentication and ACR (Azure Container Registry) support for org runners. This extends
Nuon's multi-cloud story to include Azure-native container registry and runner management capabilities.

<!-- TODO (Harsh Thakur): Confirm scope — is this GA or preview? Add any setup/config notes. -->

## GCP Improvements (Jordan Acosta, Amit Meena)

- GCP runner management mode authentication. _(Amit Meena)_
- GCP runner template added. _(Amit Meena)_
- V2 provision signal now supports GCP. _(Jordan Acosta)_
- Fake GCP stack outputs in sandbox mode for local development. _(Jordan Acosta)_

<!-- TODO (Jordan Acosta / Amit Meena): Summarize GCP auth flow and what this unblocks for users. -->

## Lifecycle Hooks & Webhook Delivery (Harsh Thakur)

Introduced lifecycle hooks for signal events and a webhook delivery system that fires on signal lifecycle transitions.
Ad hoc action outputs are now captured and surfaced.

<!-- TODO (Harsh Thakur): Describe supported hook events, webhook payload shape, and how to configure. -->

## Policy Config API & Reporting (Prem Saraswat)

Added a `GET /v1/apps/{app_id}/policy_config_id` endpoint for retrieving the active policy configuration, with
corresponding dashboard-ui integration. "Policy Evaluations" has been renamed to "Policy Reports" across the UI and
docs. Sandbox policies now correctly scope to the sandbox install only.

<!-- TODO (Prem Saraswat): Brief description of use case. -->

## Runner Process Management & Queue Infrastructure (Jon Morehouse)

Introduced a runner process management system and migrated the orgs namespace to use queues with parallel job support.
Orgs are now required to use queues, and queue lifecycle handling has been hardened with improvements to ready-state
gating, non-retry on deleted queues, and sleep/cleanup behavior. An admin panel for queue management was also added.

<!-- TODO (Jon Morehouse): Expand with user-facing summary of queue migration impact and runner process details -->

## Workflow Steps as Signals (Jon Morehouse)

Workflow steps now run as signals, changing the internal execution model. Install step transitions were also reworked.

<!-- TODO (Jon Morehouse): Explain user-facing impact — faster workflows? better observability? -->

## VCS Connection Worker (Jon Morehouse)

Added a dedicated VCS connection worker for processing version control system events, along with a vcs-worker update.

<!-- TODO (Jon Morehouse): What does this enable? Background repo sync? Webhook processing? -->

## CLI: `nuon installs stacks` Command

The CLI now supports a few new commands:

For listing and retrieving install stacks. This command is useful for scripting workflows and

```bash
nuon installs stacks
# e.g. to get the most recent AWS quick link URL
nuon installs stacks latest --json | jq '.quick_link_url'
```

`nuon installs actions` has been augmented with sub-commands to get action outputs:

```bash
# for use in CI
nuon installs actions outputs --action-workflow-id --json
```

`nuon installs components` has been augmented with sub-commands to get component outputs:

```bash
# for use in CI
nuon installs components outputs --component-id --json
```

## Dashboard UI Overhaul (Nat Hamilton, Stephen)

Major dashboard redesign covering onboarding, layout, and deploy diff views:

- **Onboarding redesign**: New onboarding dashboard with light/dark mode support, configure-inputs step design.
  _(Stephen, #792, #812, #815, #847)_
- **Active workflows view**: Redesigned active workflows page. _(Stephen, #845)_
- **Diff viewer improvements**: Overhauled Terraform, Helm, and Kubernetes diff views with better line styles and
  `to_replace` support. _(Nat Hamilton, #878, #879, #881, #882, #886)_
- **Markdown card support**: Component cards, stack cards, sandbox cards, runner cards, and surface components now
  render markdown content. _(Nat Hamilton, #853, #856, #861, #869, #876)_
- **Stack version banner**: Banner shown when a newer stack version is available. _(Nat Hamilton, #802, #803)_
- **Pulumi component type**: Added Pulumi icon and component type to the UI. _(Nat Hamilton, #798)_
- **Container pattern refactor**: Migrated UI architecture to a container pattern with Ladle stories. _(Nat Hamilton,
  #823, #838)_
- **Flexbox common components**: New shared flexbox layout components. _(Nat Hamilton, #864)_
- **Admin links**: Admin link section in user dropdown and additional admin links. _(Nat Hamilton, #808, #826)_
- **Misc fixes**: Approval banner on failed steps, input modal, subnav layout, dashboard padding fixes. _(Nat Hamilton,
  Stephen)_

<!-- TODO (Nat Hamilton / Stephen): Pick 2-3 screenshots to include. Confirm any items that should be called out more prominently. -->

## Documentation (matt-zach-s)

- New guide: **Runner Kill Switch** — explains how customers can pause or stop the Nuon Runner from their AWS account by
  scaling the ASG to zero or deleting it. _(matt-zach-s, #824)_
- Added Plausible analytics integration to docs site. _(matt-zach-s, #821)_

<!-- TODO (matt-zach-s): Anything else to highlight? -->

## Changed

- Workflow steps now execute as signals rather than inline. _(Jon Morehouse, #872)_
- Orgs are now required to use queues with parallel job support. _(Jon Morehouse, #862)_
- Install step transitions reworked. _(robisso, #855)_
- Dual writes of all status fields ensured for consistency. _(Jon Morehouse, #817)_
- Sleep functionality removed from handlers. _(Jon Morehouse, #850)_
- Org cleanup and queue sleep behavior improved. _(Jon Morehouse, #804)_
- Install migration ordering fixed during app sync. _(Somesh Koli, #857)_
- Sandbox flag removed from install input. _(robisso, #844)_
- Onboarding sandbox default set to false. _(robisso, #796)_
- AWS-specific cloud-provider fix applied. _(Amit Meena, #810)_
- Runner metadata now includes `owner_name`. _(Jordan Acosta, #809)_
- Skip deploying components on sandbox reprovisions. _(Harsh Thakur, #806)_
- Codegen fix for GithubClient breaking. _(Erick Yellott, #827)_
- "Policy Evaluations" renamed to "Policy Reports" in UI and docs. _(Prem Saraswat, #883)_
- Platform-agnostic terminology in deprovision install modal. _(Jordan Acosta, #833)_
- SDK cleanup. _(Amit Meena, #835)_

## Added

- Azure management mode & auth. _(Harsh Thakur, #811)_
- ACR & Azure org runner support. _(Harsh Thakur, #851)_
- GCP runner management mode authentication. _(Amit Meena, #793)_
- GCP runner template. _(Amit Meena, #794)_
- Lifecycle hooks system. _(Harsh Thakur, #791)_
- Webhook delivery for signal lifecycle events. _(Harsh Thakur, #822)_
- Ad hoc action outputs. _(Harsh Thakur, #801)_
- Runner process management. _(Jon Morehouse, #789)_
- Queue admin panel. _(Jon Morehouse, #867)_
- VCS connection worker. _(Jon Morehouse, #818)_
- `GET /v1/apps/{app_id}/policy_config_id` endpoint. _(Prem Saraswat, #880)_
- Fake GCP stack outputs in sandbox mode. _(Jordan Acosta, #871)_
- CLI `nuon installs stacks` command. _(fidiego, #854)_
- Runner shutdown documentation. _(matt-zach-s, #824)_
- Plausible analytics for docs. _(matt-zach-s, #821)_
- Pulumi icon and component type in UI. _(Nat Hamilton, #798)_
- Markdown support for component, stack, sandbox, runner, and surface cards. _(Nat Hamilton, #853, #856, #861, #869,
  #876)_
- Stack version banner. _(Nat Hamilton, #802)_

## Bug Fixes

- Fixed soft delete for orgs. _(robisso, #887)_
- Fixed onboarding missing region. _(robisso, #842)_
- Fixed queue update handlers not awaiting ready state before processing. _(Jon Morehouse, #846)_
- Fixed non-retry behavior when queue is deleted. _(Jon Morehouse, #865)_
- Fixed org deprovisioning and runner process hooks. _(Jon Morehouse, #800)_
- Fixed V2 provision signal for GCP. _(Jordan Acosta, #859)_
- Fixed codegen breaking on GithubClient. _(Erick Yellott, #827)_
- Reverted IAM role assumption for management ECR auth. _(Harsh Thakur, #814)_
- Fixed sandbox policies scoping only to the sandbox install. _(Jordan Acosta, #848)_
- Fixed runner type. _(Somesh Koli, #877)_
- Fixed Terraform diff `to_replace` rendering. _(Nat Hamilton, #886)_
- Fixed approval banner showing on failed steps. _(Nat Hamilton, #870)_
- Fixed current input modal. _(Nat Hamilton, #866)_
- Fixed dashboard padding. _(Nat Hamilton, #885)_
- Fixed subnav layout issues. _(Nat Hamilton, #840)_
- Fixed onboarding v2 skip and org switcher height. _(Nat Hamilton, #837)_
- Fixed install creation from app page. _(Jordan Acosta, #830)_
- Fixed guard for missing roles in install action run header. _(Nat Hamilton, #797)_
