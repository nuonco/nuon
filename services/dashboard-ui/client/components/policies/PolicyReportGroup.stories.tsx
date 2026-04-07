export default {
  title: 'Policies/PolicyReportGroup',
}

import { PolicyReportGroup } from './PolicyReportGroup'
import type { TPolicyReport } from '@/types'

const policyNameMap = new Map([
  ['pol-1', 'no-privileged-containers'],
  ['pol-2', 'resource-limits'],
])

const mockReport: TPolicyReport = {
  id: 'report-1',
  app_id: 'app-1',
  component_name: 'api',
  owner_type: 'install_deploys',
  evaluated_at: new Date(Date.now() - 600000).toISOString(),
  deny_count: 1,
  warn_count: 0,
  policies: [
    {
      policy_id: 'pol-1',
      policy_name: 'no-privileged-containers',
      status: 'deny',
      deny_count: 1,
      warn_count: 0,
    },
    {
      policy_id: 'pol-2',
      policy_name: 'resource-limits',
      status: 'pass',
      deny_count: 0,
      warn_count: 0,
    },
  ],
} as TPolicyReport

export const WithDeny = () => (
  <PolicyReportGroup report={mockReport} orgId="org-1" policyNameMap={policyNameMap} />
)

export const AllPassed = () => (
  <PolicyReportGroup
    report={{ ...mockReport, deny_count: 0, policies: mockReport.policies?.map((p) => ({ ...p, status: 'pass', deny_count: 0 })) }}
    orgId="org-1"
    policyNameMap={policyNameMap}
  />
)

export const WithWarning = () => (
  <PolicyReportGroup
    report={{
      ...mockReport,
      deny_count: 0,
      warn_count: 1,
      policies: [{ policy_id: 'pol-2', policy_name: 'resource-limits', status: 'warn', deny_count: 0, warn_count: 1 }],
    }}
    orgId="org-1"
    policyNameMap={policyNameMap}
  />
)
