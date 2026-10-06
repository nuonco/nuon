export default {
  title: 'Features / Installs / Runbooks / Row actions',
}

import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'
import { InstallRunbookRowActions } from './InstallRunbookRowActions'

const installRunbook = {
  id: 'irb-1',
  runbook_id: 'rb-1',
  runbook: {
    id: 'rb-1',
    name: 'rotate-secrets',
    configs: [
      {
        id: 'rbc-1',
        runbook_id: 'rb-1',
        readme: 'Drain one node at a time.',
      },
    ],
  },
} as TInstallRunbook

export const WithReadme = () => (
  <InstallRunbookRowActions installRunbook={installRunbook} />
)

export const WithoutReadme = () => (
  <InstallRunbookRowActions
    installRunbook={{
      ...installRunbook,
      runbook: {
        ...installRunbook.runbook,
        configs: [{ id: 'rbc-1', runbook_id: 'rb-1' }],
      },
    }}
  />
)
