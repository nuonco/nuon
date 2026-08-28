import { PanelStory } from '@/components/__stories__/helpers'
import { PolicyReportPanel } from './PolicyReportPanel'
import type { TPolicyReport } from '@/types'

export default {
  title: 'Policies/PolicyReportPanel',
}

const hour = 3600000

const mockReport = {
  id: 'report-1',
  app_id: 'app-1',
  component_name: 'payments-api',
  owner_type: 'install_deploys',
  evaluated_at: new Date().toISOString(),
  deny_count: 1,
  warn_count: 0,
  status: { status: 'error' },
  policies: [
    {
      policy_id: 'policy-1',
      policy_name: 'No public buckets',
      status: 'deny',
    },
    {
      policy_id: 'policy-2',
      policy_name: 'Resource limits set',
      status: 'pass',
    },
  ],
  violations: [
    {
      policy_id: 'policy-1',
      severity: 'deny',
      message: 'S3 bucket is publicly accessible',
      input_identity: 'Bucket/default/assets',
    },
  ],
} as unknown as TPolicyReport

const mockHistory = [
  {
    id: 'report-0',
    app_id: 'app-1',
    component_name: 'payments-api',
    owner_type: 'install_deploys',
    evaluated_at: new Date(Date.now() - hour * 6).toISOString(),
    deny_count: 0,
    warn_count: 2,
    status: { status: 'warning' },
    policies: [],
    violations: [],
  },
  {
    id: 'report-neg-1',
    app_id: 'app-1',
    component_name: 'payments-api',
    owner_type: 'install_deploys',
    evaluated_at: new Date(Date.now() - hour * 30).toISOString(),
    deny_count: 0,
    warn_count: 0,
    status: { status: 'success' },
    policies: [],
    violations: [],
  },
] as unknown as TPolicyReport[]

const policyNameMap = new Map([['policy-1', 'No public buckets']])

export const Default = () => (
  <PanelStory>
    <PolicyReportPanel
      report={mockReport}
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </PanelStory>
)

export const WithHistory = () => (
  <PanelStory>
    <PolicyReportPanel
      report={mockReport}
      history={mockHistory}
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </PanelStory>
)

export const AllPassed = () => (
  <PanelStory>
    <PolicyReportPanel
      report={
        {
          ...mockReport,
          deny_count: 0,
          status: { status: 'success' },
          policies: [],
          violations: [],
        } as unknown as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </PanelStory>
)
