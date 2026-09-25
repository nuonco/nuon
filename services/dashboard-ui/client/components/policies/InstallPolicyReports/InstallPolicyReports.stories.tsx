export default {
  title: 'Policies/InstallPolicyReports',
}

import { Button } from '@/components/common/Button'
import { groupPolicyReports } from '@/components/policies/PolicyReportsTable'
import type { TPolicyReport, TPolicyResult, TPolicyViolation } from '@/types'
import { InstallPolicyReports } from './InstallPolicyReports'

const policyNameMap = new Map([
  ['pol-1', 'no-privileged-containers'],
  ['pol-2', 'resource-limits'],
  ['pol-3', 'warn-public-eks-endpoint'],
])

const result = (
  id: string,
  status: 'pass' | 'warn' | 'deny',
  count = 1
): TPolicyResult => ({
  policy_id: id,
  policy_name: policyNameMap.get(id),
  status,
  ...(status === 'pass' ? { pass_count: count } : {}),
  ...(status === 'warn' ? { warn_count: count } : {}),
  ...(status === 'deny' ? { deny_count: count } : {}),
})

const violation = (
  policyId: string,
  severity: 'deny' | 'warn',
  message: string
): TPolicyViolation => ({
  policy_id: policyId,
  severity,
  message,
})

const minutesAgo = (n: number) =>
  new Date(Date.now() - n * 60_000).toISOString()

const denied: TPolicyReport = {
  id: 'rpt-q7fplr1up5atx5zpxotbabm',
  app_id: 'app-1',
  component_id: 'cmp-api',
  component_name: 'api-server',
  owner_type: 'install_deploys',
  evaluated_at: minutesAgo(5),
  deny_count: 1,
  status: { status: 'error', status_human_description: 'Policy denied.' },
  policies: [result('pol-1', 'deny'), result('pol-2', 'pass')],
  violations: [
    violation('pol-1', 'deny', 'Container is running as privileged'),
  ],
} as TPolicyReport

const deniedOlder: TPolicyReport = {
  ...denied,
  id: 'rpt-older1up5atx5zpxotbabm',
  evaluated_at: minutesAgo(180),
  deny_count: 0,
  pass_count: 2,
  status: { status: 'success', status_human_description: 'All checks passed.' },
  policies: [result('pol-1', 'pass'), result('pol-2', 'pass')],
  violations: [],
} as TPolicyReport

const sandboxWarn: TPolicyReport = {
  id: 'rpt-pvr1uctmt8z6oieip8feb1g53w',
  app_id: 'app-1',
  owner_type: 'install_sandbox_runs',
  evaluated_at: minutesAgo(60),
  warn_count: 2,
  status: { status: 'warning', status_human_description: 'Policy warnings.' },
  policies: [result('pol-3', 'warn', 2)],
  violations: [
    violation('pol-3', 'warn', 'EKS API server endpoint is publicly accessible'),
    violation('pol-3', 'warn', 'Cluster logging is disabled'),
  ],
} as TPolicyReport

const passed: TPolicyReport = {
  id: 'rpt-allpass2x6oieip8feb1g53w',
  app_id: 'app-1',
  component_id: 'cmp-worker',
  component_name: 'worker',
  owner_type: 'install_deploys',
  evaluated_at: minutesAgo(120),
  pass_count: 3,
  status: { status: 'success', status_human_description: 'All checks passed.' },
  policies: [result('pol-1', 'pass'), result('pol-2', 'pass')],
} as TPolicyReport

const build: TPolicyReport = {
  id: 'rpt-buildhrr1up5atx5zpxotbabm',
  app_id: 'app-1',
  component_id: 'cmp-api',
  component_name: 'api-server',
  owner_type: 'component_builds',
  evaluated_at: minutesAgo(360),
  warn_count: 1,
  status: { status: 'warning', status_human_description: 'Policy warnings.' },
  policies: [result('pol-2', 'warn')],
  violations: [violation('pol-2', 'warn', 'CPU limit not set')],
} as TPolicyReport

const rows = groupPolicyReports([
  denied,
  deniedOlder,
  sandboxWarn,
  passed,
  build,
])

const filter = <Button variant="secondary">Filter</Button>

export const Default = () => (
  <InstallPolicyReports
    rows={rows}
    orgId="org-1"
    policyNameMap={policyNameMap}
    filterActions={filter}
  />
)

export const AllPassed = () => (
  <InstallPolicyReports
    rows={groupPolicyReports([passed])}
    orgId="org-1"
    policyNameMap={policyNameMap}
  />
)

export const SingleEvaluation = () => (
  <InstallPolicyReports
    rows={groupPolicyReports([sandboxWarn])}
    orgId="org-1"
    policyNameMap={policyNameMap}
  />
)

export const NoResults = () => (
  <InstallPolicyReports
    rows={[]}
    filtered
    orgId="org-1"
    policyNameMap={policyNameMap}
    filterActions={<Button variant="secondary">Filter (1)</Button>}
  />
)

export const Empty = () => (
  <InstallPolicyReports
    rows={[]}
    orgId="org-1"
    policyNameMap={policyNameMap}
  />
)

export const Loading = () => (
  <InstallPolicyReports
    rows={[]}
    loading
    orgId="org-1"
    policyNameMap={policyNameMap}
    filterActions={filter}
  />
)
