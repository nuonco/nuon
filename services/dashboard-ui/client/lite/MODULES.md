# Modules

How Lite is packaged per customer. A **module** is a slice of the dashboard a
software vendor might want on its own, and the **modules page** is the internal
panel where Nuon staff decide which slices an org's control plane ships.

Read this before adding a page or a settings section, since each one belongs to
a module, and before changing what the sidebar shows.

## Why

Most Nuon customers run a BYOC control plane that Nuon installs for them, and
several have asked for a control plane that does one thing. Modules make that a
configuration decision rather than a fork: the same build ships everywhere, and
each deployment or org turns off what it does not need.

## The registry

`utils/modules.ts` is the single source of truth. Each entry declares:

| Field | What it decides |
|---|---|
| `id` | The module id, and the suffix of its feature flag |
| `name`, `description`, `icon` | How the modules page presents it |
| `area` | Which group it renders under: product, access, or integrations |
| `readiness` | Whether Lite has built it, partly built it, or only scaffolded it |
| `nav` | A sidebar item, with its path, shortcut, and group |
| `settings` | A settings section, with its path and label |

The current modules are Apps, Installs, Team, and the six settings sections:
Connections, Webhooks, Triggers, API tokens, Service accounts, and OIDC
federation. The dashboard index and the onboarding flow are not modules, because
the first is the shell's home and the second runs before an org exists.

**Presets** bundle modules for the requests that come up repeatedly: full
platform, core, install operations, and app delivery. A selection that matches
no preset reads as custom.

## How a module is hidden

Each module has an org feature flag named `disable-module-<id>`, declared in
`services/ctl-api/internal/app/org.go` and listed by `ModuleFeatures()`. The
polarity is deliberate: the flag is **off by default**, so every module ships
for every existing org, and turning the flag on hides the module. That also
lets a deployment pin modules off through ctl-api's existing
`forced_enabled_features` config, which is how a bespoke control plane trims
itself for every org it hosts.

Two layers, resolved in `enabledModules()`:

1. **Deployment** pins through `FORCED_ENABLED_FEATURES` on ctl-api. A pinned
   flag reads as on for every org and cannot be toggled off. The modules page
   shows these as pinned and renders their switches read only.
2. **Org** stores the flag on `org.features`. Nuon staff change it from the
   modules page, which calls the admin API through the BFF's `/admin` proxy.

The flags are admin managed. Org users cannot flip them through the public
features endpoint, even with `user-managed-features` on.

## What a hidden module removes

- Its sidebar item and keyboard shortcut, from `orgNavigation` in
  `pages/OrgLayout.tsx`.
- Its settings section from the settings sub navigation, and the Settings
  sidebar item once no section is left.
- Its routes, through a `ModuleLayout` layout route in `routes.tsx` that
  renders the org not-found page instead of the module's pages.
- Cross-module links that would dead-end. `InstallsTable` takes `appLinks` for
  this and renders app names as text when Apps is hidden.

What it does **not** remove: API access, CLI access, or data. Hiding a module is
a presentation decision for that org's dashboard.

## The modules page

`/:orgId/modules`, reached from the user menu as **Manage modules**. It renders
only for Nuon staff, by the same `@nuon.co` check the production dashboard uses,
and shows the org not-found page to anyone else.

`ModuleManager` is the organism. Its container reads the enabled set from
`ModulesProvider`, the flag catalog from the admin features list, and writes
patches with `adminUpdateOrgFeatures`. The presentational half renders three
overview cards, the preset select, one `ModuleCard` per module grouped by area,
and a deployment config block with the `FORCED_ENABLED_FEATURES` value that
pins the current selection.

## Adding a module

1. Add the flag constant to `org.go`, then to `GetFeatures`, `DefaultFeatures`
   (off), `GetFeatureDescriptions`, `adminOnlyFeatures`, and `ModuleFeatures`.
   `org_modules_test.go` checks all of that.
2. Add the registry entry in `utils/modules.ts`, including `nav` or
   `settings` when it has a surface of its own.
3. Wrap its routes in `ModuleLayout` in `routes.tsx` and extend
   `routes.test.tsx`.
4. Gate any links into it from other modules, as `InstallsTable` does.

## Known gaps

- Module state lives on the org, so an org's first cold load renders the full
  navigation until the org resolves, then trims it. Later loads read the query
  cache.
- When Connections is hidden, the Settings sidebar item links to the first
  visible section and only highlights on that section.
- The API and CLI are not gated. That is a separate, larger decision.
