# Roadmap

Here's a list of features we're currently working on.

## Logs for Builds, Deploys, and Instances

You can view logs for builds, deploys, and instances in the UI. This enables you to learn if something worked and, when it doesn't, what happened.

## Dedicated Install DNS Zones and URLs

Your customers can access an app running in their own cloud at a custom domain. This domain will be secure, isolated, and won't require additional configuration during onboarding, unless they want to customize it.

## One-click Installs

Provide a one-click solution to create an IAM role with the required policies, in a customer's account, and return the Amazon Resource Name (ARN). Currently, your customer has to create an IAM role and share its ARN manually, before you can provision the install.

## Manual Hooks

You can manually run a hook to do one-off jobs. Here are a few examples.

- run `rails db:migrate` in an install using a linked container image
- run `nslookup` in a public image, such as a standard docker debugger
- run `kubectl` commands

Hooks will only be run when deployed. They will be able to use connected component configurations, such as secrets and outputs.

## Lifecycle Event Hooks

Provide the ability to create hooks that automatically get executed in response to lifecycle events, e.g., when an install is created or a component is deployed.