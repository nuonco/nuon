export default {
  title: 'Policies/PolicyReportPanel',
}

import { PolicyReportPanelButton } from './PolicyReportPanel'

const mockReport = {
  id: 'report-1',
  app_id: 'app-1',
  component_name: 'my-component',
  owner_type: 'install_deploys',
  evaluated_at: new Date().toISOString(),
  deny_count: 1,
  warn_count: 0,
  policies: [
    {
      policy_id: 'policy-1',
      policy_name: 'No public buckets',
      status: 'deny',
    },
  ],
  violations: [
    {
      policy_id: 'policy-1',
      severity: 'deny',
      message: 'S3 bucket is publicly accessible',
    },
  ],
} as any

const policyNameMap = new Map([['policy-1', 'No public buckets']])

export const Default = () => (
  <div className="p-4">
    <PolicyReportPanelButton
      report={mockReport}
      orgId="org-1"
      policyNameMap={policyNameMap}
      onOpen={() => {}}
    />
  </div>
)
