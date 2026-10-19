---
name: nuon
description: Use when building, shipping, or troubleshooting a Nuon app. Nuon deploys a vendor's software (Terraform, Helm charts, Kubernetes manifests, container images) into their customers' own AWS, Azure, or GCP accounts.
---

# Nuon

Nuon is a BYOC (bring your own cloud) platform for software vendors. It packages Terraform, Helm, Kubernetes manifests,
and container images into an app, then installs and operates that app in customer cloud accounts through a runner in
each account.

## Where to look

| You need | Use |
| --- | --- |
| How Nuon works: concepts, config reference, CLI, guides | These docs (search here, or read pages directly) |
| Agent-only troubleshooting detail | The Markdown version of a page (see below) |
| Live state in the user's org: apps, installs, workflows, logs, approvals | The Nuon MCP at `https://mcp.nuon.co/mcp`. Setup: https://docs.nuon.co/guides/agents/setup |

## Agent-only content

Some pages carry troubleshooting detail written for agents, such as tables of why a change did not apply. It is hidden
on the web page. To read it, fetch the page's Markdown version by adding `.md` to its URL (for example
`https://docs.nuon.co/guides/app-branches.md`), or read `https://docs.nuon.co/llms-full.txt`. Check it before reading
Nuon source code to explain product behavior.

## Shipping app config

App config ships through app branches. `nuon apps sync` and `nuon sync` are deprecated: do not use or suggest them,
even if an older page or example does.

1. Add a branch file, for example `branches/main.toml`, with `name` and one repo block (`[connected_repo]` for a repo
   connected through the Nuon GitHub App, or `[public_repo]`). The repo block needs `repo`, `directory` (the app config
   path inside the repo, `"."` for the repo root), and `branch`.
2. Validate the config directory: `nuon apps validate [dir]`.
3. Register the branch once: `nuon branches sync --file branches/main.toml`.
4. Ship changes by committing and pushing to the tracked git branch. Each push starts a branch run. To start one
   without a push: `nuon branches trigger --branch-id <name>`.

Details: https://docs.nuon.co/guides/app-branches

## Key commands

| Task | Command |
| --- | --- |
| Sign in | `nuon auth login` |
| Create an app | `nuon apps create --name <name>` |
| Select the current app | `nuon apps select` |
| Validate app config | `nuon apps validate [dir]` |
| Register or update app branches | `nuon branches sync --file <file-or-dir>` |
| Start a branch run | `nuon branches trigger --branch-id <name>` |
| Watch branch runs | `nuon branches runs` |
| Agent orientation | `nuon agents help` |

## Resources

- Page index: https://docs.nuon.co/llms.txt
- Concepts: https://docs.nuon.co/concepts/overview
- Config reference: https://docs.nuon.co/config-ref/index
- App branches: https://docs.nuon.co/concepts/app-branches
- Agents and the Nuon MCP: https://docs.nuon.co/guides/agents/overview
