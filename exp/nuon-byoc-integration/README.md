# nuon-byoc-integration (experimental agent skill)

An agent skill that integrates a vendor application with the Nuon BYOC
`ctl-api`, so the vendor's customers can create and manage Nuon installs from the
vendor-owned UI — without exposing Nuon credentials to the browser.

It generates a **server-side proxy** (any language/framework) that holds the Nuon
API token and forwards to `ctl-api`, plus **frontend components** that call the
new proxy endpoints. First-class targets are Node/Express (server) and React
(client); the design is language-agnostic and extends to any stack.

## Layout

```
SKILL.md                         entry point: trigger + procedure
references/                      language-neutral protocol (source of truth)
  ctl-api.md                     endpoints, auth headers, error mapping, BYOC url
  service-account-token.md       dedicated service-account token, least privilege
  architecture-and-security.md   3-tier proxy model + security invariants
  lifecycle.md                   async create → status polling
  contracts/
    install-create.md            POST /v1/apps/{app_id}/installs contract
    install-inputs-schema.md     app input schema → dynamic form + whitelist
adapters/                       per-stack idioms (additive)
  node-express.md
  react.md
  _template.md                   copy to add a new language/framework
assets/                         copyable reference-impl snippets
  server/  nuonClient.js, installsRouter.js
  client/  installsApi.js, CreateInstallForm.jsx, InstallStatus.jsx
```

## Status

Experimental / first pass. Scope so far: **create an install** (+ dynamic input
form + async status). Extend to deploy / update-inputs / teardown by adding
files under `references/contracts/`.

To promote to a real skill, move this directory to a skills location
(`.claude/skills/`) once validated.
