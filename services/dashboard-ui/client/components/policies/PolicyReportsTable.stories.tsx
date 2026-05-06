export default {
  title: 'Policies/PolicyReportsTable',
}

import { PolicyReportsTable } from './PolicyReportsTable'
import type { TPolicyReport, TPolicyResult, TPolicyViolation } from '@/types'

const policyNameMap = new Map([
  ['pol-1', 'no-privileged-containers'],
  ['pol-2', 'resource-limits'],
  ['pol-3', 'warn-public-eks-endpoint'],
  ['pol-4', 'require-encryption-at-rest'],
])

const passed = (id: string): TPolicyResult => ({
  policy_id: id,
  policy_name: policyNameMap.get(id),
  status: 'pass',
  pass_count: 1,
})

const warned = (id: string, count = 1): TPolicyResult => ({
  policy_id: id,
  policy_name: policyNameMap.get(id),
  status: 'warn',
  warn_count: count,
})

const denied = (id: string, count = 1): TPolicyResult => ({
  policy_id: id,
  policy_name: policyNameMap.get(id),
  status: 'deny',
  deny_count: count,
})

const v = (
  policyId: string,
  severity: 'deny' | 'warn',
  message: string,
  inputIdentity?: string
): TPolicyViolation => ({
  policy_id: policyId,
  severity,
  message,
  input_identity: inputIdentity,
})

const minutesAgo = (n: number) => new Date(Date.now() - n * 60_000).toISOString()

const reportWithDeny: TPolicyReport = {
  id: 'rpt-q7fplr1up5atx5zpxotbabm',
  app_id: 'app-1',
  component_id: 'cmp-api',
  component_name: 'api-server',
  owner_type: 'install_deploys',
  evaluated_at: minutesAgo(5),
  deny_count: 1,
  policies: [denied('pol-1'), passed('pol-4')],
  violations: [
    v(
      'pol-1',
      'deny',
      'Container is running as privileged',
      'Deployment/default/api-server'
    ),
  ],
} as TPolicyReport

const reportSandboxWarn: TPolicyReport = {
  id: 'rpt-pvr1uctmt8z6oieip8feb1g53w',
  app_id: 'app-1',
  owner_type: 'install_sandbox_runs',
  evaluated_at: minutesAgo(60),
  warn_count: 1,
  policies: [warned('pol-3')],
  violations: [
    v(
      'pol-3',
      'warn',
      'EKS API server endpoint is publicly accessible',
      'aws_eks_cluster.sandbox'
    ),
  ],
} as TPolicyReport

const reportAllPassed: TPolicyReport = {
  id: 'rpt-allpass2x6oieip8feb1g53w',
  app_id: 'app-1',
  component_id: 'cmp-worker',
  component_name: 'worker',
  owner_type: 'install_deploys',
  evaluated_at: minutesAgo(120),
  pass_count: 3,
  policies: [passed('pol-1'), passed('pol-2'), passed('pol-4')],
} as TPolicyReport

const reportBuild: TPolicyReport = {
  id: 'rpt-buildhrr1up5atx5zpxotbabm',
  app_id: 'app-1',
  component_id: 'cmp-api',
  component_name: 'api-server',
  owner_type: 'component_builds',
  evaluated_at: minutesAgo(360),
  warn_count: 2,
  policies: [warned('pol-2', 2)],
  violations: [
    v('pol-2', 'warn', 'CPU limit not set'),
    v('pol-2', 'warn', 'Memory limit not set'),
  ],
} as TPolicyReport

const reportMixed: TPolicyReport = {
  id: 'rpt-mixedc933tcyzji01s7us3aeo3',
  app_id: 'app-1',
  component_id: 'cmp-gateway',
  component_name: 'gateway',
  owner_type: 'install_deploys',
  evaluated_at: minutesAgo(15),
  deny_count: 1,
  warn_count: 2,
  pass_count: 1,
  policies: [
    denied('pol-1'),
    warned('pol-3'),
    warned('pol-2', 2),
    passed('pol-4'),
  ],
  violations: [
    v('pol-1', 'deny', 'Privileged container'),
    v('pol-3', 'warn', 'Public EKS endpoint'),
    v('pol-2', 'warn', 'CPU limit not set'),
    v('pol-2', 'warn', 'Memory limit not set'),
  ],
} as TPolicyReport

export const SingleReport = () => (
  <PolicyReportsTable
    reports={[reportWithDeny]}
    orgId="org-1"
    installId="install-1"
    policyNameMap={policyNameMap}
  />
)

export const MultipleReports = () => (
  <PolicyReportsTable
    reports={[reportWithDeny, reportSandboxWarn, reportAllPassed, reportBuild]}
    orgId="org-1"
    installId="install-1"
    policyNameMap={policyNameMap}
  />
)

export const ManyReports = () => (
  <PolicyReportsTable
    reports={[
      reportWithDeny,
      reportMixed,
      reportSandboxWarn,
      reportBuild,
      reportAllPassed,
      { ...reportWithDeny, id: 'rpt-2', evaluated_at: minutesAgo(720) },
      { ...reportSandboxWarn, id: 'rpt-3', evaluated_at: minutesAgo(1440) },
      { ...reportAllPassed, id: 'rpt-4', evaluated_at: minutesAgo(2880) },
    ]}
    orgId="org-1"
    installId="install-1"
    policyNameMap={policyNameMap}
  />
)

export const ReprovisionHistory = () => (
  // 5 sandbox reprovisions + 3 deploys for one component + 1 deploy for another.
  // Should render only 3 group cards (sandbox, api-server deploys, worker deploys),
  // each with a 'Show N earlier evaluations' toggle.
  <PolicyReportsTable
    reports={[
      // Sandbox reprovisions
      {
        ...reportSandboxWarn,
        id: 'rpt-sb-1',
        evaluated_at: minutesAgo(0.4),
      } as TPolicyReport,
      {
        ...reportSandboxWarn,
        id: 'rpt-sb-2',
        evaluated_at: minutesAgo(60),
      } as TPolicyReport,
      {
        ...reportSandboxWarn,
        id: 'rpt-sb-3',
        evaluated_at: minutesAgo(180),
        warn_count: 0,
        deny_count: 1,
        policies: [denied('pol-3')],
      } as TPolicyReport,
      {
        ...reportSandboxWarn,
        id: 'rpt-sb-4',
        evaluated_at: minutesAgo(720),
      } as TPolicyReport,
      {
        ...reportSandboxWarn,
        id: 'rpt-sb-5',
        evaluated_at: minutesAgo(1440),
        warn_count: 0,
        pass_count: 1,
        policies: [passed('pol-3')],
      } as TPolicyReport,
      // api-server deploys
      reportWithDeny,
      {
        ...reportWithDeny,
        id: 'rpt-api-2',
        evaluated_at: minutesAgo(45),
      } as TPolicyReport,
      {
        ...reportWithDeny,
        id: 'rpt-api-3',
        evaluated_at: minutesAgo(180),
        deny_count: 0,
        pass_count: 2,
        policies: [passed('pol-1'), passed('pol-4')],
      } as TPolicyReport,
      // worker deploy (single)
      reportAllPassed,
    ]}
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

export const EmptyWithFilters = () => (
  <PolicyReportsTable
    reports={[]}
    orgId="org-1"
    installId="install-1"
    policyNameMap={policyNameMap}
    currentStatus="warning"
  />
)
