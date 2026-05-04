export default {
  title: 'Policies/PolicyReportGroup',
}

import { PolicyReportGroup } from './PolicyReportGroup'
import type { TPolicyReport, TPolicyResult, TPolicyViolation } from '@/types'

const policyNameMap = new Map([
  ['pol-1', 'no-privileged-containers'],
  ['pol-2', 'resource-limits'],
  ['pol-3', 'warn-public-eks-endpoint'],
  ['pol-4', 'require-encryption-at-rest'],
  ['pol-5', 'allowed-image-registries'],
  ['pol-6', 'no-host-network-access'],
  ['pol-7', 'require-resource-limits'],
])

const baseReport: TPolicyReport = {
  id: 'rpt-q7fplr1up5atx5zpxotbabm',
  app_id: 'app-1',
  component_name: 'api-server',
  owner_type: 'install_deploys',
  evaluated_at: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
} as TPolicyReport

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div className="max-w-4xl">{children}</div>
)

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

export const SinglePassed = () => (
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          policies: [passed('pol-1')],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const SingleWarning = () => (
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          warn_count: 1,
          policies: [warned('pol-3')],
          violations: [
            v(
              'pol-3',
              'warn',
              'EKS API server endpoint is publicly accessible',
              'aws_eks_cluster.main'
            ),
          ],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const SingleDenied = () => (
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          deny_count: 2,
          policies: [denied('pol-1', 2)],
          violations: [
            v(
              'pol-1',
              'deny',
              'Container is running as privileged',
              'Deployment/default/api-server'
            ),
            v(
              'pol-1',
              'deny',
              'SecurityContext.privileged must be false',
              'Deployment/default/worker'
            ),
          ],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const MixedStatuses = () => (
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          deny_count: 1,
          warn_count: 2,
          pass_count: 2,
          policies: [
            denied('pol-1'),
            warned('pol-3'),
            warned('pol-2', 2),
            passed('pol-4'),
            passed('pol-5'),
          ],
          violations: [
            v('pol-1', 'deny', 'Container is running as privileged'),
            v('pol-3', 'warn', 'EKS endpoint is public'),
            v('pol-2', 'warn', 'CPU limit not set', 'Deployment/api-server'),
            v('pol-2', 'warn', 'Memory limit not set', 'Deployment/api-server'),
          ],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const ManyPolicies = () => (
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          deny_count: 2,
          warn_count: 3,
          pass_count: 4,
          policies: [
            denied('pol-1'),
            denied('pol-6'),
            warned('pol-3'),
            warned('pol-2', 2),
            warned('pol-7'),
            passed('pol-4'),
            passed('pol-5'),
            passed('pol-1'),
            passed('pol-6'),
          ],
          violations: [
            v('pol-1', 'deny', 'Privileged container detected'),
            v('pol-6', 'deny', 'Host network access enabled'),
            v('pol-3', 'warn', 'EKS endpoint is public'),
            v('pol-2', 'warn', 'CPU limit missing'),
            v('pol-2', 'warn', 'Memory limit missing'),
            v('pol-7', 'warn', 'No resource quotas configured'),
          ],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const SandboxReport = () => (
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          component_name: undefined,
          owner_type: 'install_sandbox_runs',
          warn_count: 1,
          policies: [warned('pol-3')],
          violations: [
            v(
              'pol-3',
              'warn',
              'EKS endpoint is public',
              'aws_eks_cluster.sandbox'
            ),
          ],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const SandboxMultipleWarningsOnePolicy = () => (
  // Sandbox report where one policy has multiple warning violations.
  // Single row, click "Show details" expands to a list of all warnings.
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          component_name: undefined,
          owner_type: 'install_sandbox_runs',
          warn_count: 3,
          policies: [warned('pol-3', 3)],
          violations: [
            v(
              'pol-3',
              'warn',
              'EKS API server endpoint is publicly accessible',
              'aws_eks_cluster.sandbox'
            ),
            v(
              'pol-3',
              'warn',
              'EKS cluster has no private subnets configured',
              'aws_eks_cluster.sandbox'
            ),
            v(
              'pol-3',
              'warn',
              'EKS cluster logging is incomplete',
              'aws_eks_cluster.sandbox'
            ),
          ],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const SandboxMultipleWarningPolicies = () => (
  // Sandbox report with several policies, each emitting a warning.
  // Each row gets its own "Show details" toggle.
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          component_name: undefined,
          owner_type: 'install_sandbox_runs',
          warn_count: 4,
          policies: [
            warned('pol-3'),
            warned('pol-2', 2),
            warned('pol-4'),
            warned('pol-7'),
            passed('pol-1'),
          ],
          violations: [
            v(
              'pol-3',
              'warn',
              'EKS API server endpoint is publicly accessible',
              'aws_eks_cluster.sandbox'
            ),
            v('pol-2', 'warn', 'CPU limit not set', 'kubernetes_deployment.api'),
            v(
              'pol-2',
              'warn',
              'Memory limit not set',
              'kubernetes_deployment.api'
            ),
            v(
              'pol-4',
              'warn',
              'EBS volume encryption not enabled',
              'aws_ebs_volume.data'
            ),
            v(
              'pol-7',
              'warn',
              'Namespace has no ResourceQuota',
              'kubernetes_namespace.app'
            ),
          ],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const LongNamesAndMessages = () => (
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          component_name:
            'api-server-with-an-extremely-long-component-name-for-overflow-testing',
          deny_count: 1,
          policies: [
            {
              policy_id: 'pol-long',
              policy_name:
                'enforce-very-long-policy-name-that-should-truncate-or-wrap-gracefully-no-matter-what',
              status: 'deny',
              deny_count: 1,
            } as TPolicyResult,
          ],
          violations: [
            v(
              'pol-long',
              'deny',
              'This violation message is intentionally extremely long to verify that the rendered card wraps the message text properly without breaking the layout, even when the input identity is also very long.',
              'apps/v1/Deployment/default/api-server-with-a-very-long-name-too'
            ),
          ],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const StatusWithoutViolations = () => (
  // Edge case: policy has warn/deny status but report.violations is empty.
  // Expected: status badge renders, but no "Show details" button.
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          warn_count: 1,
          policies: [warned('pol-3')],
          violations: [],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const NoAppId = () => (
  // Edge case: report has no app_id, so policy names render as plain text (no link).
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          app_id: undefined,
          deny_count: 1,
          policies: [denied('pol-1')],
          violations: [v('pol-1', 'deny', 'Container is privileged')],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const NoPolicies = () => (
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          policies: [],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)

export const PolicyWithoutNameOnlyId = () => (
  // Edge case: policy_name missing, policyNameMap has no entry. Falls back to id.
  <Wrap>
    <PolicyReportGroup
      report={
        {
          ...baseReport,
          warn_count: 1,
          policies: [
            {
              policy_id: 'pol-unknown',
              status: 'warn',
              warn_count: 1,
            } as TPolicyResult,
          ],
          violations: [v('pol-unknown', 'warn', 'Policy violation detected')],
        } as TPolicyReport
      }
      orgId="org-1"
      policyNameMap={policyNameMap}
    />
  </Wrap>
)
