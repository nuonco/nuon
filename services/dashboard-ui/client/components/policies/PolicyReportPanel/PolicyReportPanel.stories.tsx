import { PanelStory } from '@/components/__stories__/helpers'
import { PolicyReportPanel } from './PolicyReportPanel'
import type { TPolicyReport } from '@/types'

export default {
  title: 'Features / Policies / Report panel',
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

const skipReport: TPolicyReport = {
  ...mockReport,
  id: 'report-skip',
  component_name: 'database',
  deny_count: 0,
  policies: [
    {
      policy_id: 'policy-skip',
      policy_name: 'Defer database replacements',
      status: 'skip_apply',
    },
  ],
  violations: [],
  status: {
    status: 'success',
    status_human_description: 'Policy decision: skip component apply',
    metadata: {
      resolved_action: 'skip_apply',
      skip_count: 1,
      skip_findings: [
        {
          policy_id: 'policy-skip',
          message: 'Database replacement requires a maintenance window.',
        },
      ],
    },
  },
}

export const SkipRequested = () => (
  <PanelStory>
    <PolicyReportPanel
      report={skipReport}
      orgId="org-1"
      policyNameMap={policyNameMap}
      history={[skipReport]}
    />
  </PanelStory>
)

export const SandboxSkip = () => (
  <PanelStory>
    <PolicyReportPanel
      report={{
        ...skipReport,
        owner_type: 'install_sandbox_runs',
        component_name: undefined,
        policies: [
          {
            policy_id: 'policy-skip',
            policy_name: 'Defer sandbox EKS changes',
            status: 'skip_apply',
          },
        ],
        status: {
          ...skipReport.status,
          status_human_description: 'Policy decision: skip sandbox apply',
          metadata: {
            ...skipReport.status?.metadata,
            skip_findings: [
              {
                policy_id: 'policy-skip',
                message: 'Sandbox apply deferred for EKS version 1.34.',
              },
            ],
          },
        },
      }}
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </PanelStory>
)

export const DenyOverridesSkip = () => (
  <PanelStory>
    <PolicyReportPanel
      report={{
        ...skipReport,
        deny_count: 1,
        policies: [
          ...(skipReport.policies || []),
          ...(mockReport.policies || []),
        ],
        violations: mockReport.violations,
        status: {
          ...skipReport.status,
          status: 'error',
          status_human_description: 'Change denied by policy',
          metadata: { ...skipReport.status?.metadata, resolved_action: 'deny' },
        },
      }}
      orgId="org-1"
      policyNameMap={policyNameMap}
      history={[
        {
          ...skipReport,
          id: 'report-deny-skip',
          deny_count: 1,
          status: {
            status: 'error',
            metadata: { skip_count: 1, resolved_action: 'deny' },
          },
        },
        skipReport,
      ]}
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
