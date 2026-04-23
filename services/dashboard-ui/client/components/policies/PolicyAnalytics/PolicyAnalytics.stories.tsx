import { DateTime } from 'luxon'
import { PolicyAnalytics } from './PolicyAnalytics'
import type {
  TPolicyAnalyticsBreakdown,
  TPolicyAnalyticsSummary,
  TPolicyAnalyticsTimeseries,
} from '@/types'

export default { title: 'Policies/PolicyAnalytics' }

const now = DateTime.now().toUTC()

const mockSummary: TPolicyAnalyticsSummary = {
  total_evaluations: 1247,
  total_denies: 23,
  total_warns: 89,
  total_passes: 1135,
  unique_reports: 312,
  unique_policies: 8,
  start: now.minus({ days: 30 }).toISO()!,
  end: now.toISO()!,
}

const mockTimeseries: TPolicyAnalyticsTimeseries = {
  interval: 'day',
  group_by: [],
  start: now.minus({ days: 30 }).toISO()!,
  end: now.toISO()!,
  buckets: Array.from({ length: 15 }, (_, i) => ({
    time: now.minus({ days: 30 - i * 2 }).toISO()!,
    evaluations: 30 + Math.floor(Math.random() * 60),
    passes: 25 + Math.floor(Math.random() * 50),
    warns: Math.floor(Math.random() * 8),
    denies: Math.floor(Math.random() * 3),
  })),
}

const mockByPolicy: TPolicyAnalyticsBreakdown = {
  dimension: 'policy_id',
  entries: [
    { key: 'pol_restrict_ns', evaluations: 400, denies: 12, warns: 20, passes: 368 },
    { key: 'pol_ecr_only', evaluations: 350, denies: 8, warns: 30, passes: 312 },
    { key: 'pol_resource_limits', evaluations: 300, denies: 3, warns: 25, passes: 272 },
    { key: 'pol_no_latest_tag', evaluations: 197, denies: 0, warns: 14, passes: 183 },
  ],
}

const mockByInstall: TPolicyAnalyticsBreakdown = {
  dimension: 'install_id',
  entries: [
    { key: 'ins_customer_acme', evaluations: 500, denies: 15, warns: 40, passes: 445 },
    { key: 'ins_customer_globex', evaluations: 400, denies: 5, warns: 30, passes: 365 },
    { key: 'ins_customer_initech', evaluations: 347, denies: 3, warns: 19, passes: 325 },
  ],
}

const mockByOwnerType: TPolicyAnalyticsBreakdown = {
  dimension: 'owner_type',
  entries: [
    { key: 'install_deploys', evaluations: 850, denies: 18, warns: 60, passes: 772 },
    { key: 'component_builds', evaluations: 310, denies: 5, warns: 22, passes: 283 },
    { key: 'install_sandbox_runs', evaluations: 87, denies: 0, warns: 7, passes: 80 },
  ],
}

const mockPolicyNames: Record<string, string> = {
  pol_restrict_ns: 'Restricted namespaces',
  pol_ecr_only: 'ECR images only',
  pol_resource_limits: 'Resource limits',
  pol_no_latest_tag: 'No latest tag',
}

const mockInstallNames: Record<string, string> = {
  ins_customer_acme: 'Acme Corp',
  ins_customer_globex: 'Globex Inc',
  ins_customer_initech: 'Initech Ltd',
}

const defaultProps = {
  policyNames: mockPolicyNames,
  installNames: mockInstallNames,
  selectedRange: '30d',
  onRangeChange: () => {},
}

export const Default = () => (
  <PolicyAnalytics
    summary={mockSummary}
    timeseries={mockTimeseries}
    byPolicy={mockByPolicy}
    byInstall={mockByInstall}
    byOwnerType={mockByOwnerType}
    isLoading={false}
    {...defaultProps}
  />
)

export const Loading = () => (
  <PolicyAnalytics
    summary={undefined}
    timeseries={undefined}
    byPolicy={undefined}
    byInstall={undefined}
    byOwnerType={undefined}
    isLoading={true}
    {...defaultProps}
  />
)

export const Empty = () => (
  <PolicyAnalytics
    summary={{
      total_evaluations: 0,
      total_denies: 0,
      total_warns: 0,
      total_passes: 0,
      unique_reports: 0,
      unique_policies: 0,
      start: '',
      end: '',
    }}
    timeseries={{ interval: 'day', group_by: [], start: '', end: '', buckets: [] }}
    byPolicy={{ dimension: 'policy_id', entries: [] }}
    byInstall={{ dimension: 'install_id', entries: [] }}
    byOwnerType={{ dimension: 'owner_type', entries: [] }}
    isLoading={false}
    {...defaultProps}
  />
)

export const HighViolations = () => (
  <PolicyAnalytics
    summary={{
      ...mockSummary,
      total_denies: 156,
      total_warns: 234,
      total_passes: 857,
      total_evaluations: 1247,
    }}
    timeseries={{
      ...mockTimeseries,
      buckets: mockTimeseries.buckets.map((b) => ({
        ...b,
        denies: 5 + Math.floor(Math.random() * 10),
        warns: 8 + Math.floor(Math.random() * 15),
      })),
    }}
    byPolicy={mockByPolicy}
    byInstall={mockByInstall}
    byOwnerType={mockByOwnerType}
    isLoading={false}
    selectedRange="7d"
    onRangeChange={() => {}}
    policyNames={mockPolicyNames}
    installNames={mockInstallNames}
  />
)
