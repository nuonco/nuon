export default {
  title: 'Runbooks/RunbookReadmePanel',
}

import { PanelStory } from '@/components/__stories__/helpers'
import { RunbookReadmePanel } from './RunbookReadmePanel'

const readme = `# Rotate secrets

Rotates the API keys and database credentials for this install.

## When to run this

- A credential has leaked or is suspected to have leaked
- Quarterly rotation is due
- A teammate with access has left the team

## Before you start

1. Confirm the install is healthy — a rotation during a failed deploy leaves
   half the components on the old credentials.
2. Tell the on-call channel. Workers reconnect during step 3 and requests can
   fail for up to a minute.

## Steps

| Step | What it does | Typical duration |
| --- | --- | --- |
| \`generate\` | Issues new credentials in the secret store | ~10s |
| \`deploy\` | Rolls the API and worker components | 2–4 min |
| \`verify\` | Runs a smoke check against the public endpoint | ~30s |

\`\`\`bash
nuon runbooks run rotate-secrets --install inst-1
\`\`\`

## If it fails

Re-run the runbook. Steps are idempotent and the old credentials stay valid
until \`verify\` passes.
`

export const Default = () => (
  <PanelStory>
    <RunbookReadmePanel readme={readme} runbookName="rotate-secrets" />
  </PanelStory>
)

export const ShortReadme = () => (
  <PanelStory>
    <RunbookReadmePanel
      readme="Drains nodes one at a time so pods reschedule without downtime."
      runbookName="drain-nodes"
    />
  </PanelStory>
)

export const NoReadme = () => (
  <PanelStory>
    <RunbookReadmePanel runbookName="failover-database" />
  </PanelStory>
)

export const WithTrigger = () => (
  <RunbookReadmePanel
    readme={readme}
    runbookName="rotate-secrets"
    panelKey="runbook-readme-rbk-1"
    triggerButton={{ variant: 'secondary', size: 'sm', children: 'Readme' }}
  />
)
