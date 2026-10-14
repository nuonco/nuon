import { AgentPromptBanner } from './AgentPromptBanner'

export default { title: 'UI / Agent prompt banner' }

const heading = 'Update the install version'
const description =
  'Copy this prompt and paste it into an agent. It names the install and the version to set.'
const prompt = `Update the acme-payments install (inst-1) to version 2.4.0.

Use the Nuon CLI against org org-1. Confirm the deploy starts and report the workflow id when it is running.`

export const Collapsed = () => (
  <div className="max-w-xl p-6">
    <AgentPromptBanner
      heading={heading}
      description={description}
      prompt={prompt}
    />
  </div>
)

export const Expanded = () => (
  <div className="max-w-xl p-6">
    <AgentPromptBanner
      heading={heading}
      description={description}
      prompt={prompt}
      defaultExpanded
    />
  </div>
)

export const LongHeading = () => (
  <div className="max-w-xl p-6">
    <AgentPromptBanner
      heading="Move acme-payments to the new vpc and update every component that still points at the old network"
      description="Copy this prompt and paste it into an agent. It names the install and the network to move it to."
      prompt={`Move the acme-payments install (inst-1) in org org-1 from vpc-old to vpc-new in us-east-1.

Leave the install version unchanged. Confirm the deploy starts and report the workflow id.`}
      defaultExpanded
    />
  </div>
)

export const LongDescription = () => (
  <div className="max-w-xl p-6">
    <AgentPromptBanner
      heading="Retry the failed deploy"
      description="Copy this prompt and paste it into an agent. It names the failed workflow, the step that stopped, and what to check before retrying. The agent should read the step log first, then retry only that step, and stop if the same error comes back."
      prompt={`Retry the failed deploy for the acme-payments install (inst-1).

The workflow id is wfl-1. The step that failed is "sync and plan certificate". Read that step's log before retrying. If the error is unchanged, stop and quote it.`}
      defaultExpanded
    />
  </div>
)

export const LongPrompt = () => (
  <div className="max-w-xl p-6">
    <AgentPromptBanner
      heading="Repair the install after a partial deploy"
      description="Copy this prompt and paste it into an agent. It walks through the failed deploy and the components to repair."
      prompt={`Inspect the acme-payments install (inst-1) in org org-1. A deploy stopped halfway through, so some components are on the new release and some are still on the previous one.

Read the latest workflow before changing anything. The workflow id is wfl-1. The steps that failed are "deploy worker" and "deploy ingress". Quote the error from each step, then say whether it is safe to retry.

These components still need to land on version 2.4.0:

- payments-api
- worker
- scheduler
- ingress
- redis

For each one, read the current config and the last successful deploy. Show the diff before writing. Do not change the runner version. Do not rotate secrets. Do not touch components that are already on 2.4.0.

After the configs are updated, sync the app and start a deploy. Report the new workflow id, then watch it until the install is healthy or a step fails. If a step fails, quote the error, name the component, and stop. Do not retry on your own.

When the install is healthy, list the version now running for each component above and call out anything that is still behind.`}
      defaultExpanded
    />
  </div>
)

export const CustomButtonText = () => (
  <div className="max-w-xl p-6">
    <AgentPromptBanner
      buttonText="Ask Claude"
      heading="Pin the runner version"
      description="Copy this prompt and paste it into Claude. It pins the runner on this install to a specific release."
      prompt={`Pin the runner on the acme-payments install (inst-1) to version 1.8.2.

Leave the install version unchanged. Confirm the runner reports the new version.`}
      defaultExpanded
    />
  </div>
)
