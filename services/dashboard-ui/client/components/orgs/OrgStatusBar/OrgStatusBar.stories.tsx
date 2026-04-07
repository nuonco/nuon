export default {
  title: 'Orgs/OrgStatusBar',
}

import { OrgStatusBar } from './OrgStatusBar'

export const Default = () => (
  <OrgStatusBar
    org={{ id: 'org-1', name: 'My Org' } as any}
    runnerConnected={true}
    runnerStatus="connected"
    runnerId="runner-1"
    approvals={[]}
    activeWorkflows={[]}
    approvalItems={[]}
    workflowItems={[]}
  />
)
