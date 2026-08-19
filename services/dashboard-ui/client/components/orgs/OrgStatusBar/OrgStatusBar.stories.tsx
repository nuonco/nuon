export default {
  title: 'Orgs/OrgStatusBar',
}

import { OrgStatusBar } from './OrgStatusBar'

export const Default = () => (
  <OrgStatusBar
    org={{ id: 'org-1', name: 'My Org' } as any}
    approvals={[]}
    activeWorkflows={[]}
    approvalItems={[]}
    workflowItems={[]}
  />
)
