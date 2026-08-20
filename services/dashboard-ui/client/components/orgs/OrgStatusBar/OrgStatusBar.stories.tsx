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

export const WithBYOCBadge = () => (
  <OrgStatusBar
    org={{ id: 'org-1', name: 'My Org' } as any}
    approvals={[]}
    activeWorkflows={[]}
    approvalItems={[]}
    workflowItems={[]}
    byocName="acme payments"
    byocColor="#1A6B4A"
    byocTextColor="#FFFFFF"
  />
)

export const WithBYOCBadgeLightColor = () => (
  <OrgStatusBar
    org={{ id: 'org-1', name: 'My Org' } as any}
    approvals={[]}
    activeWorkflows={[]}
    approvalItems={[]}
    workflowItems={[]}
    byocName="acme"
    byocColor="#F5D90A"
    byocTextColor="#000000"
  />
)
