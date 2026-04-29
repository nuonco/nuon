# Docs Reorganization Design (2026-04-27)

## Problem

The current docs site doesn't flow easy → hard. Tabs go *Documentation → Changelog → Concepts → Guides → Architecture → Knowledge Base*, which is out of order — Changelog is a recurring resource, Concepts is the foundation, Architecture is isolated from Concepts despite being part of "how it works." The "Documentation" tab is a grab-bag of Get Started, Resources, Platform Support, and Support. The "Guides" tab is bastardized — most pages are short how-to recipes, not long-form guides. Some content is duplicated across top-level files (`cli.mdx`, `dashboard.mdx`) and concept pages of the same name. Prospective engineers can't quickly answer *"does Nuon fit my situation?"* — qualifying info (deployment options, supported clouds) is buried.

## Motivation

CEO's pyramid (sketched 2026-04-27): docs should flow **easy → hard** as **Setup → How-to → Guides → KB**, with a "How it works" layer between Setup and How-to. Recurring/marketing content (Changelog, Support, Pricing) belongs in a Resources dropdown, not in the main flow. Knowledge Base stays external (support.nuon.co).

## Frameworks used

Confirmed with author 2026-04-27:

- **Diátaxis** ([diataxis.fr](https://diataxis.fr)) — 4-quadrant doc IA: Tutorials / How-to guides / Reference / Explanation. Maps near-1:1 to the CEO's pyramid: Setup ≈ Tutorials, How-it-works ≈ Explanation, How-tos ≈ How-to guides, Reference is its own. KB is external.
- **writethedocs.org** beginner guide + structure/style topic pages — for tone, audience analysis, plain language.
- **Mintlify docs** — platform-specific patterns: tabs, groups, anchors, redirects.

Plugins evaluated but skipped: `danielrosehill/documentation-plugin` and `matsengrp/plugins` were checked; neither offered enough additional value over the existing toolkit for this reorg.

## Final tab structure

Top nav, left → right (= easy → hard):

1. **Get Started** (Tutorials)
2. **How it Works** (Explanation — was Concepts + Architecture)
3. **How-tos** (How-to guides — renamed from "Guides", contains both quick recipes and end-to-end integrations as sub-groups)
4. **Reference** (Reference — CLI, Config, API, SDKs)
5. **Resources ▾** (dropdown — Changelog, Support, Contact)
6. **Knowledge Base ↗** (external link to support.nuon.co)

The CEO's pyramid distinguishes "How-tos" (quick) from "Guides" (long-form) as separate steps. We collapse this into a single tab named **How-tos** with two internal sub-groups: `Quick recipes` and `End-to-end integrations`. Authors no longer debate "is this a guide or a how-to?" — they pick a sub-group based on length and shape. Boundary heuristic: ≤ ~150 lines and single-task → Quick recipe; ~200+ lines or multi-task → End-to-end. The 150–200 zone is judgment-call; resolve in PR review.

## Sidebar by tab

### Get Started

```
Welcome                              [rewrite intro hero cards]

— Does it fit? —
Deployment Options
  ├─ Nuon Cloud (default)
  ├─ Nuon BYOC
  └─ Self-hosted
Supported Clouds                     [renamed from "Platform Support"]
  ├─ AWS
  ├─ Azure
  └─ GCP

— Get hands-on —
Quickstart                           [install CLI · log in · connect GitHub]
Set up the CLI
Set up the Dashboard
Create your first install
  ├─ on AWS (Kubernetes)             [was app-aws-k8s]
  ├─ on AWS (Lambda)                 [was app-aws-lambda]
  ├─ on Azure                        [NEW — gap today]
  └─ on GCP                          [NEW — gap today]
```

### How it Works

```
Overview
Glossary

— Architecture —
Platform Architecture                [was architecture/platform]
App Deployment Architecture          [was architecture/byoc]
Runner Authentication                [was architecture/runner-auth]
Security                             [was top-level security.mdx]

— Connect Your App —
Apps
Components
Sandboxes
Inputs
Secrets
App Variables
Runners

— Customer Infrastructure —
Installs                             [NEW concept page — currently missing!]
Stacks
Operation Roles

— Continuous Delivery —
Workflows
Actions
Policies

— Customer Experience —
Customer Portal
```

### How-tos

```
— Quick recipes — (≤ ~150 lines, action-focused)

  Connect Your App
    Manage apps
    Manage components
    Configure inputs and secrets
    Configure sandboxes
    Configure component dependencies
    Use variables
    Connect VCS
  Components
    Docker build component
    Container image component
    Helm chart component
    Kubernetes manifest component
    Terraform module component
  Operations
    Run actions
    Configure policies                [quick slice]
    Configure operation roles         [quick slice]
    Configure runner management mode
    Configure team management
    Configure custom domains
    Configure notifications
    Use READMEs
    Use the runner kill switch
  Customer Experience
    Customer portal
    Vendor's customers
  Integrations
    GitHub Actions
    AI & Apps

— End-to-end integrations — (~200+ lines, multi-step)

  Set up BYOC end-to-end              [was guides/byoc]
  Self-host the control plane
    ├─ Self-hosting overview
    ├─ AWS
    ├─ Azure
    └─ GCP
  App init walkthrough
  Build custom nested stacks
  Configure policies in depth
  Use external image policies
  Use Helm charts in depth
  Use Kubernetes manifests in depth
  Use Terraform with the CLI
  Build a CLI extension
  Integrate with the control plane
```

### Reference

```
CLI
  ├─ Install & upgrade                [was top-level cli.mdx]
  ├─ Authentication & login
  ├─ TUI / preview features
  └─ Extensions reference

Dashboard
  └─ Tour                             [was top-level dashboard.mdx]

Config Reference (TOML)               [unchanged structure under config-ref/]
  ├─ Index
  ├─ Component types
  ├─ Configuration types
  └─ Other types

API Reference                         [unchanged — generated from OpenAPI]

SDKs                                  [was sdks.mdx]
```

### Resources ▾

```
Changelog                             [was top-level "Changelog" tab]
Support / FAQ                         [was support/support.mdx]
Talk to us                            [external link to nuon.co/contact-sales]
```

### Knowledge Base ↗

External link only: https://support.nuon.co — no change.

## Page disposition table

Disposition codes: **MOVE** (relocate as-is) · **MERGE** (combine with another page) · **SPLIT** (one source → multiple destinations) · **KILL** (delete; content covered elsewhere) · **CREATE** (new page) · **REWRITE** (substantial content rework needed)

### Top-level files

| Source | Disposition | Destination | Notes |
|---|---|---|---|
| `cli.mdx` | SPLIT | Reference/CLI/Install (install bits) + Reference/CLI/Authentication + Reference/CLI/TUI + Reference/CLI/Extensions | Currently 138-line kitchen-sink page |
| `dashboard.mdx` | MOVE+REWRITE | Reference/Dashboard/Tour | Rewrite to be a complete tour, not just screenshots |
| `configuration-files.mdx` | KILL+SPLIT | Sync sections → Get Started/Set up the CLI; reference content → Config Reference index | 368 lines, mostly duplicates `config-ref/` |
| `deployment-options.mdx` | MOVE+REWRITE | Get Started/Deployment Options | Expand into 3 sub-pages (Cloud / BYOC / Self-hosted) |
| `nuon-api.mdx` | KILL | (covered by Reference/API + concepts/api) | 29-line stub redundant with `api-ref/` |
| `sdks.mdx` | MOVE | Reference/SDKs | |
| `security.mdx` | MOVE | How it Works/Architecture/Security | |
| `pricing.mdx` | NO ACTION | (not in nav today; leave for marketing site) | Remove from `/docs/` if confirmed unused |
| `README.md` | NO ACTION | (repo readme, unchanged) | |

### `architecture/`

| Source | Disposition | Destination | Notes |
|---|---|---|---|
| `architecture/byoc.mdx` | MOVE | How it Works/Architecture/App Deployment Architecture | 19 lines — flesh out further |
| `architecture/platform.mdx` | MOVE | How it Works/Architecture/Platform Architecture | 39 lines — flesh out |
| `architecture/runner-auth.mdx` | MOVE | How it Works/Architecture/Runner Authentication | 220 lines, in good shape |

### `concepts/`

| Source | Disposition | Destination | Notes |
|---|---|---|---|
| `concepts/overview.mdx` | MOVE+REWRITE | How it Works/Overview | Update card grid to match new IA |
| `concepts/glossary.mdx` | MOVE | How it Works/Glossary | |
| `concepts/apps.mdx` | MOVE | How it Works/Connect Your App/Apps | |
| `concepts/components.mdx` | MOVE | How it Works/Connect Your App/Components | |
| `concepts/sandboxes.mdx` | MOVE | How it Works/Connect Your App/Sandboxes | |
| `concepts/app-inputs.mdx` | MOVE | How it Works/Connect Your App/Inputs | |
| `concepts/app-secrets.mdx` | MOVE | How it Works/Connect Your App/Secrets | |
| `concepts/app-variables.mdx` | MOVE | How it Works/Connect Your App/App Variables | |
| `concepts/runners.mdx` | MOVE+REWRITE | How it Works/Connect Your App/Runners | 36 lines — expand |
| `concepts/stacks.mdx` | MOVE | How it Works/Customer Infrastructure/Stacks | |
| `concepts/operation-roles.mdx` | MOVE | How it Works/Customer Infrastructure/Operation Roles | |
| `concepts/workflows.mdx` | MOVE | How it Works/Continuous Delivery/Workflows | |
| `concepts/actions.mdx` | MOVE | How it Works/Continuous Delivery/Actions | |
| `concepts/policies.mdx` | MOVE | How it Works/Continuous Delivery/Policies | |
| `concepts/customer-portal.mdx` | MOVE+REWRITE | How it Works/Customer Experience/Customer Portal | 32 lines — expand |
| `concepts/dashboard.mdx` | KILL | (merged into Reference/Dashboard/Tour) | 20 lines, redundant |
| `concepts/api.mdx` | KILL | (covered by Reference/API and api-ref) | 11-line stub |
| `concepts/cli.mdx` | MOVE+REWRITE | How it Works/Customer Infrastructure/Install Methods | Reframe as "ways to create and manage installs" alongside Dashboard, Customer Portal, API |

### `get-started/`

| Source | Disposition | Destination | Notes |
|---|---|---|---|
| `get-started/introduction.mdx` | MOVE+REWRITE | Get Started/Welcome | Update hero cards: Get Started / How it Works / How-tos / KB / Talk to us |
| `get-started/quickstart.mdx` | MOVE | Get Started/Quickstart | Light edit to flow into setup pages |
| `get-started/create-your-first-app.mdx` | KILL | (replaced by new "Create your first install" landing) | Currently a 40-line stub linking to two tutorials |
| `get-started/app-aws-k8s.mdx` | MOVE+RENAME | Get Started/Create your first install/On AWS (Kubernetes) | |
| `get-started/app-aws-lambda.mdx` | MOVE+RENAME | Get Started/Create your first install/On AWS (Lambda) | |

### `platform-support/`

| Source | Disposition | Destination | Notes |
|---|---|---|---|
| `platform-support/introduction.mdx` | MOVE+RENAME | Get Started/Supported Clouds (landing) | |
| `platform-support/aws.mdx` | MOVE | Get Started/Supported Clouds/AWS | |
| `platform-support/azure.mdx` | MOVE | Get Started/Supported Clouds/Azure | |
| `platform-support/gcp.mdx` | MOVE+REWRITE | Get Started/Supported Clouds/GCP | 22 lines — expand |

### `guides/` — Quick recipes

| Source | Lines | Destination |
|---|---|---|
| `guides/managing-apps.mdx` | 73 | Connect Your App/Manage apps |
| `guides/managing-components.mdx` | 38 | Connect Your App/Manage components |
| `guides/configuring-inputs-and-secrets.mdx` | 125 | Connect Your App/Configure inputs and secrets |
| `guides/configuring-sandboxes.mdx` | 117 | Connect Your App/Configure sandboxes |
| `guides/component-dependencies.mdx` | 111 | Connect Your App/Configure component dependencies |
| `guides/using-variables.mdx` | 97 | Connect Your App/Use variables |
| `guides/vcs.mdx` | 85 | Connect Your App/Connect VCS |
| `guides/docker-build-components.mdx` | 96 | Components/Docker build component |
| `guides/container-image-components.mdx` | 161 | Components/Container image component |
| `guides/terraform-components.mdx` | 91 | Components/Terraform module component |
| `guides/actions.mdx` | 193 | Operations/Run actions |
| `guides/configuring-policies.mdx` (split) | (excerpt) | Operations/Configure policies (quick slice) |
| `guides/operation-roles.mdx` (split) | (excerpt) | Operations/Configure operation roles (quick slice) |
| `guides/runner-management-mode.mdx` | 58 | Operations/Configure runner management mode |
| `guides/runner-kill-switch.mdx` | 74 | Operations/Use the runner kill switch |
| `guides/team-management.mdx` | 31 | Operations/Configure team management |
| `guides/custom-domains.mdx` | 106 | Operations/Configure custom domains |
| `guides/notifications.mdx` | 34 | Operations/Configure notifications |
| `guides/using-readmes.mdx` | 148 | Operations/Use READMEs |
| `guides/customer-portal.mdx` | 96 | Customer Experience/Customer portal |
| `guides/vendor-customers.mdx` | 49 | Customer Experience/Vendor's customers |
| `guides/github-actions.mdx` | 68 | Integrations/GitHub Actions |
| `guides/ai-and-apps.mdx` | 34 | Integrations/AI & Apps |

### `guides/` — End-to-end integrations

| Source | Lines | Destination |
|---|---|---|
| `guides/byoc.mdx` | 302 | Set up BYOC end-to-end |
| `guides/self-hosting/index.mdx`, `aws.mdx`, `azure.mdx`, `gcp.mdx` | varied | Self-host the control plane (overview + per-cloud sub-pages) |
| `guides/app-init.mdx` | 379 | App init walkthrough |
| `guides/custom-nested-stacks.mdx` | 375 | Build custom nested stacks |
| `guides/configuring-policies.mdx` (split) | remainder | Configure policies in depth |
| `guides/external-image-policies.mdx` | 703 | Use external image policies |
| `guides/helm-chart-components.mdx` | 298 | Use Helm charts in depth |
| `guides/kubernetes-manifest-components.mdx` | 232 | Use Kubernetes manifests in depth |
| `guides/terraform-cli.mdx` | 179 | Use Terraform with the CLI |
| `guides/cli-extensions.mdx` | 332 | Build a CLI extension |
| `guides/operation-roles.mdx` (split) | remainder | (merge into How it Works/Customer Infrastructure/Operation Roles concept page) |
| `guides/control-plane-integration.mdx` | 239 | Integrate with the control plane |
| `guides/app-install-life-cycle.mdx` | 106 | KILL — content folds into How it Works/Customer Infrastructure/Installs (new concept page) |

### `support/`, `updates/`, `api-ref/`, `config-ref/`

| Source | Disposition | Destination |
|---|---|---|
| `support/support.mdx` | MOVE | Resources/Support |
| `updates/*.mdx` (32 files) | MOVE | Resources/Changelog (keep updates/ folder; just point dropdown at it) |
| `api-ref/**` | NO ACTION | Reference/API Reference (unchanged) |
| `config-ref/**` | NO ACTION | Reference/Config Reference (unchanged structure) |

### `snippets/`, `images/`, `logo/`

No action — unchanged.

## New pages to create

1. **`concepts/installs.mdx`** — the missing core concept page. Currently `/concepts/installs` redirects to `/guides/app-install-life-cycle`. Pull lifecycle content into a real concept page; keep "lifecycle" framing as a sub-section.
2. **`get-started/app-azure.mdx`** — "Create your first install on Azure." Tutorial parity with the AWS pages.
3. **`get-started/app-gcp.mdx`** — "Create your first install on GCP." Tutorial parity.
4. **`get-started/deployment-options-cloud.mdx`** — Nuon Cloud deployment option detail. (Or split inline within `deployment-options.mdx` — to be decided in implementation.)
5. **`get-started/deployment-options-byoc.mdx`** — Nuon BYOC deployment option detail.
6. **`get-started/deployment-options-self-hosted.mdx`** — Self-hosted deployment option detail.
7. **`get-started/setup-cli.mdx`** — Setup-focused page (post-quickstart). Distinct from `Reference/CLI/Install & upgrade` (which is the lookup page).
8. **`get-started/setup-dashboard.mdx`** — Walks the user through first sign-in, org selection, GitHub connection in the UI.
9. **`reference/cli/index.mdx`** — landing for CLI reference (split target of today's `cli.mdx`).
10. **`reference/dashboard/index.mdx`** — landing for the dashboard tour reference.

## Doc content updates (rewrites)

Beyond pure relocation, these pages need substantive rewrites:

- **`get-started/introduction.mdx` (Welcome)** — hero cards currently link to old paths (`/guides/app-install-life-cycle`, `/runner-architecture`). Rewrite to point to the new IA: Get Started → How it Works → How-tos → KB → Talk to us. Also keep the YouTube embed.
- **`get-started/quickstart.mdx`** — light edit to flow into the new `Set up the CLI` and `Set up the Dashboard` pages.
- **`deployment-options.mdx`** (now Get Started/Deployment Options) — the current page is thin and confusing. Rewrite to be a clear comparison: when do you choose Nuon Cloud vs BYOC vs Self-hosted? Include a comparison table.
- **`platform-support/gcp.mdx`** — only 22 lines. Expand to parity with AWS/Azure (see Azure 102 lines as the model).
- **`concepts/runners.mdx`** — only 36 lines. Expand: install vs build runner, lifecycle, observability, security implications.
- **`concepts/customer-portal.mdx`** — 32 lines, expand.
- **`concepts/cli.mdx`** — reframe from "what the CLI does" to "ways to create and manage installs" — list dashboard, CLI, customer portal, API as peers.
- **`concepts/overview.mdx`** — update card grid to match new IA (Connect Your App / Customer Infrastructure / Continuous Delivery / Customer Experience).
- **`dashboard.mdx`** (now Reference/Dashboard/Tour) — the current page is mostly captioned screenshots. Add narrative for what each tab does and why a user would visit it.
- **`security.mdx`** — currently top-level, mixes architecture and feature description. Move to How it Works/Architecture/Security and link out to specific feature pages (policies, operation roles, break-glass) instead of rehashing them.
- **`get-started/create-your-first-app.mdx` → "Create your first install" landing** — rewrite as a branched landing page choosing AWS/Azure/GCP.
- **NEW `concepts/installs.mdx`** — write from scratch; absorb `guides/app-install-life-cycle.mdx` content.
- **`configuration-files.mdx`** — the install/sync workflow part needs to live in Get Started/Set up the CLI; the rest is duplicated reference and should be killed in favor of `config-ref/`.

## Style guidelines (brief)

Conventions to follow when writing or rewriting any page:

1. **Audience-first**: every page opens by stating who the page is for and what they'll be able to do after reading it.
2. **Action verbs in titles**: "Configure custom domains" beats "Custom domains" for how-tos. Reserve noun titles for concept and reference pages.
3. **No FAQs as catch-alls** (per writethedocs). If something keeps coming up, give it its own page.
4. **Plain language**: avoid "simply", "just", animal idioms, and jargon without first definition.
5. **Diátaxis discipline**:
   - Tutorials (Get Started) — *learning*, hand-held, must succeed. No conceptual digressions.
   - How-tos — *task-oriented*, assume competence, no teaching.
   - Concepts (How it Works) — *understanding*, no step-by-step instructions.
   - Reference — *information lookup*, terse, complete.
6. **Links over duplication**. Concepts get explained ONCE in How it Works; how-tos and tutorials link there. The CEO's note ("we need duplicate information, but at different depths") is honored by *summary blurbs* with a "learn more →" link, not by republishing the same content.
7. **Page templates** (suggested defaults — refine in implementation):
   - **Tutorial**: Goal · Prerequisites · Steps · What you built · Next steps
   - **How-to**: Use case · Steps · Verification · Troubleshooting · Related
   - **Concept**: Definition · Why it exists · How it relates to (other concepts) · Configuration · Related
   - **Reference**: Index of fields/commands · Per-item reference · Examples

## Backwards compatibility / redirects

Only the **README links** are load-bearing per CEO ("biggest place is the open source readme"). Email and internal links can be hand-fixed.

**URL strategy.** Mintlify URLs follow file paths (no extension, relative to docs root). Renaming a tab in `docs.json` does *not* change URLs — only moving files does. To keep the filesystem coherent with the new IA, this reorg moves files (e.g. `architecture/*.mdx` → `how-it-works/architecture/*.mdx`); URL changes are therefore unavoidable. Use `git mv` so PR diffs show as renames.

Mintlify redirect strategy:

1. **Per-page `redirects[]`** in `docs.json` for every URL that changes.
2. **Folder-level redirects** for moves (e.g. `/architecture/*` → `/how-it-works/architecture/*`).
3. Keep the existing 5 redirects in `docs.json` working.

Required new redirects (non-exhaustive — full table generated during implementation):

| Old | New |
|---|---|
| `/architecture/byoc` | `/how-it-works/architecture/byoc` |
| `/architecture/platform` | `/how-it-works/architecture/platform` |
| `/architecture/runner-auth` | `/how-it-works/architecture/runner-auth` |
| `/concepts/*` | `/how-it-works/*` |
| `/guides/byoc` | `/how-tos/byoc-end-to-end` |
| `/guides/<short-recipe>` | `/how-tos/<bucket>/<recipe>` |
| `/platform-support/*` | `/get-started/supported-clouds/*` |
| `/configuration-files` | `/get-started/setup-cli` |
| `/cli` | `/reference/cli/install` |
| `/dashboard` | `/reference/dashboard/tour` |
| `/security` | `/how-it-works/architecture/security` |
| `/sdks` | `/reference/sdks` |
| `/support/support` | `/resources/support` |
| `/updates/*` | `/resources/changelog/*` (or keep `updates/` URL, just regroup the tab) |
| `/changelog` | `/resources/changelog` (existing redirect updates) |
| `/runner-architecture` | `/how-it-works/architecture/platform` (existing redirect updates) |
| `/concepts/installs` | `/how-it-works/customer-infrastructure/installs` (was redirecting to a guide; fix to point at new concept page) |
| `/faq` | `/resources/support` (existing redirect updates) |

Also: scan the open source README at `github.com/nuonco/nuon` after the move and update any `/docs.nuon.co/...` links that 404.

## Phased implementation plan

Each phase is independently shippable. Phases 1–3 are the IA reorg; Phase 4 is content updates that can land in parallel after Phase 1.

### Phase 1 — IA skeleton (no content changes)

- Update `docs.json` to the new tab/group structure.
- Move files to new paths. Use `git mv` to preserve history.
- Fix internal links inside MDX files (relative paths break on move).
- Add the redirect table.
- Verify Mintlify build is clean.

**Acceptance**: site builds, all old links redirect, no 404s.

### Phase 2 — New pages (content gaps)

- Create `concepts/installs.mdx` (the missing core concept).
- Create `get-started/setup-cli.mdx` and `setup-dashboard.mdx`.
- Stub `get-started/app-azure.mdx` and `app-gcp.mdx` (full content can land later; ship a "coming soon → see AWS for now" stub if needed to fill the IA).
- Split `cli.mdx` into the 4 reference sub-pages.

**Acceptance**: every node in the new sitemap has a real page (no `404s`, no orphan groups).

### Phase 3 — Kill duplicates

- Delete `configuration-files.mdx`, `nuon-api.mdx`, `concepts/dashboard.mdx`, `concepts/api.mdx`, `get-started/create-your-first-app.mdx`, `guides/app-install-life-cycle.mdx`.
- Confirm redirects send former visitors to the right replacements.
- Update sitemap submission to search engines if applicable.

**Acceptance**: no overlapping content, redirects in place, search rankings unaffected (monitor for 1 week).

### Phase 4 — Content rewrites (parallelizable per-page)

- Rewrites enumerated in *Doc content updates* above.
- One PR per page (or per logical bundle) so reviewers can move fast.
- Each rewrite gets a Diátaxis label in the PR description so the reviewer can check fit.

**Acceptance**: every page on the rewrite list has been touched and reviewed against its template.

### Phase 5 — Style pass

- One reviewer makes a pass through every tutorial, how-to, concept, and reference page applying the style guidelines (audience preamble, action verbs, plain language).
- Build a small house style guide as a Resources page or wiki entry. Keep it < 1 page.

**Acceptance**: every page has been style-passed; a short style guide exists.

## Out of scope / non-goals

- **Visual redesign / Mintlify theme changes** — colors, fonts, layout: not changing.
- **API reference content** (`api-ref/`) — auto-generated, untouched.
- **Marketing site (`website-v2/`, `website/`)** — separate repo concern.
- **The internal wiki (`services/wiki/`)** — different audience, separate concern.
- **Knowledge base content (support.nuon.co)** — external system, not part of this repo.
- **Pricing page** — currently exists in `/docs/` but isn't navigated; assume it's marketing-only, leave alone unless author confirms removal.

## Open questions / followups

These can be resolved in implementation, but flagging for the in-person session:

1. **Configure policies / Configure operation roles** — these guides each have a "quick" and "in-depth" portion. Splitting them is the right call but requires editorial judgment on the cut line.
2. **`guides/control-plane-integration.mdx` (239 lines)** — bucketed as End-to-end. Could also fit as a *Reference* page if it's primarily API surface.
3. **`pricing.mdx`** — kill or keep? Not in nav today.
4. **Whether to keep `concepts/cli.mdx` distinct from `Reference/CLI`** — proposed yes (concept = "ways to create installs"; reference = "how to install the binary"). Validate in person.
5. **Knowledge Base authority boundary** — what content belongs in support.nuon.co (KB) vs How-tos? Current heuristic: troubleshooting and FAQ → KB; positive-path "how to do X" → docs how-to. Worth codifying.

## Author

Drafted 2026-04-27 by Matt Schultheiss with Claude. Built off the CEO's pyramid sketch (2026-04-27) and the team docs call transcript of the same day.
