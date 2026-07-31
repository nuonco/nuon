---
name: feature-flag-promote
description: Flip an existing org feature flag's default from false to true and emit the admin bulk-enable request body for existing orgs
---

You are promoting an org feature flag: its default flips from `false` to `true` so all newly created orgs get it,
and you produce the admin API request that flips it on for all existing orgs. The flag itself and its gating code
stay in place — removal happens later via `/feature-flag-cleanup` once the promotion has been stable in prod.

The flag name may be passed as an argument: $ARGUMENTS

## Step 1: Resolve and validate the flag

1. If no flag name was given, list the flags currently defaulted to `false` in the `defaultFeatures` map in
   `Org.BeforeCreate` (`services/ctl-api/internal/app/org.go`) and ask the user which one to promote, in plain
   prose (never AskUserQuestion).
2. Verify the flag exists as an `OrgFeature` constant and appears in `GetFeatures()`. If it doesn't exist, stop
   and tell the user — this skill never creates flags.
3. Verify its current default is `false`. If it is already `true`, say so and skip to Step 3 (the bulk-enable
   body may still be what the user wants).
4. Check for special cases and surface them before editing:
   - Is the flag in `adminOnlyFeatures` in `org.go`? Promotion is still valid, but note that users cannot toggle
     it themselves.
   - Grep `services/ctl-api/internal/pkg/features/` and `Org.BeforeCreate` for logic naming the flag
     (mutual-exclusion rules, config-driven overrides). If any exists, describe it and confirm with the user
     before proceeding.

## Step 2: Flip the default

In the `defaultFeatures` map in `Org.BeforeCreate`:

1. Move the flag's entry from the `// Disabled by default` section to the `// Enabled by default` section and set
   it to `true`. Do not touch any other entry, the constants block, `GetFeatures()`, or `GetFeatureDescriptions()`.
2. Format and verify:
   ```bash
   gofmt -w ./services/ctl-api/internal/app/ && goimports -w ./services/ctl-api/internal/app/
   go build ./services/ctl-api/...
   ```

## Step 3: Emit the admin bulk-enable request body

Output the request the user runs against prod so existing orgs match the new default (the default-map change only
affects orgs created after it deploys):

```
PATCH /v1/orgs/admin-toggle-feature
{
  "features": {
    "<flag-name>": true
  }
}
```

Handler: `services/ctl-api/internal/app/orgs/service/admin_toggle_orgs_feature.go` — patches the given flags onto
every org via jsonb merge, preserving each org's other flags. There is also a per-org variant at
`PATCH /v1/orgs/{org_id}/admin-features` if the user wants a staged rollout instead; offer it when the flag gates
something risky.

Unlike cleanup, ordering is forgiving here: the flag still exists in `GetFeatures()`, so the PATCH is valid both
before and after the default-flip deploys. Recommend running it at (or shortly after) deploy time so existing and
new orgs converge.

## Step 4: Remind about the cleanup horizon

Tell the user: once this promotion has shipped in a tagged release and survived ~2 weeks, the flag becomes a
candidate in `/feature-flag-cleanup`, which removes the flag and its gating code entirely. If this flag should
never be cleaned up (permanent kill switch), suggest adding it to the keep-list in
`.claude/commands/feature-flag-cleanup.md` now.
