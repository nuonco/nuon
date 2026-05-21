<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./docs/images/nuon_readme_logo_dark.gif">
  <img src="./docs/images/nuon_readme_logo_light.gif" alt="Nuon logo">
</picture>

</div>

<h1>Nuon</h1>

[![BYOC](https://img.shields.io/badge/BYOC-blue)](https://awesomebyoc.com/)
[![Go report card](https://goreportcard.com/badge/github.com/nuonco/nuon)](https://goreportcard.com/report/github.com/nuonco/nuon)
[![Godoc](https://pkg.go.dev/badge/github.com/nuonco/nuon.svg)](https://pkg.go.dev/github.com/nuonco/nuon)
[![Version](https://img.shields.io/github/v/tag/nuonco/nuon)](https://github.com/nuonco/nuon/tags)
[![BYOC community slack](https://img.shields.io/badge/slack-byoc_community-purple)](https://nuon-byoc.slack.com/join/shared_invite/zt-1q323vw9z-C8ztRP~HfWjZx6AXi50VRA)

Nuon is an open-source platform for software vendors to deploy and operate their software in their customers' cloud
accounts.

This software category is called Bring Your Own Cloud (BYOC).

## What is Nuon?

Historically there have been two ways to deploy any software vendor's product: SaaS (in the vendor's cloud) and
self-hosted (the customer manually installs the vendor's software in their cloud).

But some customers have compliance and regulatory reasons that require some or all of the vendor's software be installed
in the customers' cloud accounts. Customers who self-host would also prefer if the software vendor installed and managed
their software versus managing the complexity themselves.

Now there is an emerging deployment alternative called Bring Your Own Cloud (BYOC) where the software vendor installs
and manages their software - in their customers' cloud accounts.

Nuon is a platform that software vendors can use to quickly and securely offer BYOC to their customers.

## Getting started

The fastest way to get started with Nuon is to [sign up for a free trial](https://app.nuon.co) on Nuon Cloud.

The docs for creating your first app are [here](https://docs.nuon.co/get-started/create-your-first-app). Nuon maintains
a list of [example apps](https://github.com/nuonco/example-app-configs) to test Nuon and install them into AWS. We will
add more examples that install into Azure and Google Cloud.

## Deployment options

Nuon supports three deployment models:

- **Nuon Cloud** — Fully managed, multi-tenant SaaS hosted by Nuon. The fastest way to get started;
  [sign up for a free trial](https://app.nuon.co).
- **Nuon BYOC** — Single-tenant Nuon control plane deployed into your AWS account, managed remotely by Nuon. You own the
  infrastructure and data. See the [BYOC docs](https://docs.nuon.co/guides/byoc).
- **Nuon Self-Hosted** — Deploy and operate Nuon entirely within your own infrastructure, with no dependency on Nuon
  Cloud. Supported on AWS. See the [self-hosted docs](https://docs.nuon.co/guides/self-hosted).

See the full [deployment options guide](https://docs.nuon.co/guides/deployment-options) for a detailed comparison.

## Documentation

Checkout our [Knowledge Base](https://support.nuon.co) for commonly asked questions.

Browse our docs [here](https://docs.nuon.co) or visit a specific section below:

- [Concepts](https://docs.nuon.co/concepts/overview): apps, components, actions, inputs, installs, workflows
- [Architecture](https://docs.nuon.co/runner-architecture): control plane, runner, [CLI](https://docs.nuon.co/cli),
  [API](https://docs.nuon.co/nuon-api)
- [Deployment options](https://docs.nuon.co/deployment-options): Nuon Cloud, Nuon BYOC, and Nuon Self-Hosted
- [Guides](https://docs.nuon.co/guides/app-install-life-cycle): technical info on component types, actions, variables,
  and more

Also browse the [Nuon blog](https://nuon.co/blog) for technical posts on specific features and example app guides.

## Support

Have a technical question? Please [open a GitHub issue](https://github.com/nuonco/nuon/issues), or join
[our Community Slack](https://join.slack.com/t/nuon-byoc/shared_invite/zt-3kzp3zpn4-0pHH4kGZ3OJul2p_y1Mzag) — the
`help` channel is the best place to chat with the team.

## Enterprise

Dedicated support and additional security and governance features are available for an annual license fee. Check out our
enterprise features [here](https://nuon.co/pricing).

## Changelog

When Nuon does a promotion to Nuon Cloud, we [publish a changelog](https://docs.nuon.co/updates/updates) of new
features.

## Contributors

Review our contributing guidelines [here](CONTRIBUTING.md).

<a href="https://github.com/nuonco/nuon/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=nuonco/nuon" />
</a>

Made with [contrib.rocks](https://contrib.rocks).
