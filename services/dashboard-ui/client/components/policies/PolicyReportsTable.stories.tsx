export default {
  title: 'Policies/PolicyReportsTable',
}

import { PolicyReportsTable } from './PolicyReportsTable'
import type { TPolicyReport } from '@/types'

const policyNameMap = new Map([['pol-1', 'no-privileged-containers']])

const mockReports: TPolicyReport[] = [
  {
    id: 'report-1',
    app_id: 'app-1',
    component_name: 'api',
    owner_type: 'install_deploys',
    evaluated_at: new Date(Date.now() - 300000).toISOString(),
    deny_count: 0,
    warn_count: 0,
    policies: [{ policy_id: 'pol-1', status: 'pass', deny_count: 0, warn_count: 0 }],
  } as TPolicyReport,
  {
    id: 'report-2',
    app_id: 'app-1',
    component_name: 'worker',
    owner_type: 'install_deploys',
    evaluated_at: new Date(Date.now() - 900000).toISOString(),
    deny_count: 1,
    warn_count: 0,
    policies: [{ policy_id: 'pol-1', status: 'deny', deny_count: 1, warn_count: 0 }],
  } as TPolicyReport,
]

export const Default = () => (
  <PolicyReportsTable
    reports={mockReports}
    orgId="org-1"
    installId="install-1"
    policyNameMap={policyNameMap}
  />
)

export const Empty = () => (
  <PolicyReportsTable
    reports={[]}
    orgId="org-1"
    installId="install-1"
    policyNameMap={policyNameMap}
  />
)
