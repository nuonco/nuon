# Engineer Onboarding

Onboarding guide for new engineers.

## First Day

Your first day at Nuon looks something like this:

1. Initial call with Jon in the early afternoon to start onboarding.
1. You will collaboratively create an onboarding checklist Epic together.
1. Make sure you can access every part of the access list today.
1. Get introduced to customers in each slack channel.
1. Update your calendar and add you to daily standups.

**NOTE** we start the onboarding after our morning standup, as many times the team is focused on addressing customer feedback and other urgen issues.

## Access List

The following systems are important to get access to on day 1:

1. Slack / Email (before first day)
1. Github
1. Aws Console
1. Admin Panel
1. Wiki
1. Datadog
1. Community Slack

You can find links to these in our [links](./links.md) document.

**NOTE** we manage access to the AWS console by adding you to the `engineers-root` google group.

[infra/github](../infra/github) is used to manage access to github, as documented [here](./github.md).

## Development / Product List

You will want to make sure you can run the system locally, and setup some type of sample application that you can play 
around with Nuon using. This involves:

1. Walk through the [day to day development guide](./day_to_day_development.md)
1. Make sure you get nuonctl installed using our install script first, and then setup from the mono repo. [Guide](./nuonctl.md).
1. Running the API locally from the mono repo. [Guide](./running_services_locally.md)
1. Seeding your local environment and using the CLI. [Guide](./seed_data.md)
1. Run the dashboard locally.
1. Run an installer locally.

## Support

## Important Repos

To get started, you should only need [mono](./working-in-mono.md), however you may want to clone other repos as you go:

1. `nuonco/terraform-provider-nuon`
1. `nuonco/sandboxes` (or individual sandbox repos)j
1. `nuonco/installer`

## Wiki Updates

We aim to have the `wiki` be our source of truth, as you are onboarding, please push updates to the wiki as you go. Things that are unclear, not documented or otherwise should be addressed as we go, to make sure the next engineer onboarding is even easier.

## Customers

It is important that every engineer is connected to customers to help with support, gather feedback and roll out new features.

In your first few weeks, we will add you to customer channels in both our community slack and internal slack, and start inviting you to calls.

As you address feedback and start shipping new functionality, you will start updating and working with customers directly.

**NOTE** we are still improving our SOPS around customer communication and will be documenting this soon!

## First Month

Your first month at Nuon will be focused on small wins, bugs, and product refinements. This is designed so you can quickly gather the lay of the land across our entire product, and we encourage you to run each part of the system locally, push documentation and other small changes etc.

After your first month, you will naturally get to a place where you can own your first project, and that will be designed with the team based on customer feedback and priorities.

## Project Management

We have a light board and roadmap in [Github Projects](https://github.com/orgs/powertoolsdev/projects/16/views/2).

We have a legacy document about [how we operate](https://www.notion.so/nuon/How-we-operate-2d0176db71c54701a70686460de846ba) that might be useful.

## Notion

We previously used notion, but are going all in on our wiki that is managed in the `mono` repo. Some things may still be updated there, which we will port over.
