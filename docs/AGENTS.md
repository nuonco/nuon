# Docs

`docs/` is the Mintlify site published at https://docs.nuon.co. Navigation and site config live in
`docs.json`. Preview locally with `./dev.sh` (see [docs/README.md](README.md)); it serves at
http://localhost:3333.

The conventions below apply to any page you edit — don't retrofit pages you aren't otherwise
touching.

## Feature names

Capitalize feature names: **App Branches, Install Groups, Install Configs, Previews, Actions,
Runbooks, Policies, Operation Roles, Break Glass, Triggers, Install Telemetry**. Singular forms
capitalize the same way (an App Branch, an Install Group).

Core platform nouns stay lowercase: app, install, component, sandbox, runner, workflow, org.

Exception: when quoting a dashboard UI label, keep the UI's exact casing in bold — the dashboard
uses sentence case, e.g. **Change app branch**.

The rule covers headings, link text, and `<Card>` titles too: "Configure App Branches", not
"Configure app branches".

CLI commands, flags, and TOML keys stay in code formatting with their literal casing (`nuon
branches sync`, `--app-branch-id`, `app_branch`).

## Linking features

Link the first mention of a feature in the page body to its concept page (`/concepts/...`). Don't
link later mentions, headings, or the title. A page about one feature links only to *other*
features on first mention — not to itself. An explicit cross-reference ("See
[Previews](/guides/app-branches#previews) for ...") can link again.

Features without their own concept page link to the page or section that defines them:

| Feature | Link |
| --- | --- |
| Install Groups | `/concepts/app-branches#install-groups` |
| Previews | `/guides/app-branches#previews` |
| Install Configs | `/guides/install-configs` |

## Naming config files

Name TOML config files by type: "branch config file", "install config file" — never just "file"
or "branch file". The file type stays lowercase; link its first mention to the feature's page
(e.g. [install config file](/guides/install-configs)). The app config is a directory, so it stays
"app config" (no "file").

## Tabs for multiple methods

When a task can be done more than one way, put each method in a Mintlify `<Tabs>` block, using
only the methods that apply, in this order:

- `Config file`
- `CLI`
- `Dashboard`

Mintlify syncs tabs with matching titles across a page, so keeping titles consistent lets a reader
pick a method once and have it stick for the rest of the page.

```mdx
<Tabs>
  <Tab title="Config file">
    Set `app_branch` in the install config file, then sync it:

    ```sh
    nuon installs sync -d installs/ --app-id <app-id> --confirm
    ```
  </Tab>
  <Tab title="CLI">
    ```sh
    nuon installs create --app-id <app-id> --name acme --region us-east-1 --app-branch-id <branch-id>
    ```
  </Tab>
  <Tab title="Dashboard">
    Choose the App Branch in the create install flow.
  </Tab>
</Tabs>
```

## Search

Mintlify ranks title matches above heading matches above body matches; `boost` only reorders
results within a match group, it doesn't move a page into a higher group.

- Put the term users will actually search for in `title`.
- Use `keywords` for related terms — they raise relevance but aren't title matches.
- Keep `boost` in the existing range: concept pages up to ~4, changelog pages below 1. Prefer
  setting boost on the nav group in `docs.json` over per-page `boost` when the whole group should
  move together.

## Before opening a PR

```bash
cd docs && bunx mint@<MINT_VERSION from dev.sh> broken-links
```

This does not validate `#anchors` — check those by hand against the target page's headings.
