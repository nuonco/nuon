import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PolicyReportPanelButton } from '@/components/policies/PolicyReportPanel'
import type { TPolicyReport, TPolicyResult } from '@/types'

function formatOwnerType(ownerType: string): {
  label: string
  theme: 'info' | 'brand' | 'neutral'
} {
  switch (ownerType) {
    case 'install_deploys':
      return { label: 'Deploy', theme: 'info' }
    case 'install_sandbox_runs':
      return { label: 'Sandbox', theme: 'brand' }
    case 'component_builds':
      return { label: 'Build', theme: 'neutral' }
    default:
      return { label: ownerType, theme: 'neutral' }
  }
}

function getOverallStatusBadge(report: TPolicyReport) {
  if ((report.deny_count || 0) > 0) {
    return (
      <Badge theme="error" size="md">
        <Icon variant="XCircle" size={12} />
        Denied
      </Badge>
    )
  }
  if ((report.warn_count || 0) > 0) {
    return (
      <Badge theme="warn" size="md">
        <Icon variant="Warning" size={12} />
        Warning
      </Badge>
    )
  }
  return (
    <Badge theme="success" size="md">
      <Icon variant="CheckCircle" size={12} />
      Passed
    </Badge>
  )
}

function getPolicyStatusBadge(policy: TPolicyResult) {
  if (policy.status === 'deny') {
    return (
      <Badge theme="error" size="sm">
        <Icon variant="XCircle" size={10} />
        Denied
      </Badge>
    )
  }
  if (policy.status === 'warn') {
    return (
      <Badge theme="warn" size="sm">
        <Icon variant="Warning" size={10} />
        Warning
      </Badge>
    )
  }
  return (
    <Badge theme="success" size="sm">
      <Icon variant="CheckCircle" size={10} />
      Passed
    </Badge>
  )
}

interface IPolicyReportGroup {
  report: TPolicyReport
  orgId: string
  policyNameMap: Map<string, string>
}

export function PolicyReportGroup({ report, orgId, policyNameMap }: IPolicyReportGroup) {
  const { label: ownerTypeLabel, theme: ownerTypeTheme } = formatOwnerType(
    report.owner_type || ''
  )
  const policies = report.policies || []

  return (
    <Card className="!p-0 !gap-0 overflow-hidden">
      <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
        <div className="flex items-center gap-6 min-w-0">
          <div className="flex flex-col gap-0.5 min-w-0">
            <Text weight="strong" variant="body">
              {report.component_name ? `Component - ${report.component_name}` : 'Sandbox'}
            </Text>
            <ID>{report.id || ''}</ID>
          </div>
          <div className="flex items-center gap-1.5">
            <Text variant="subtext" theme="neutral">Type:</Text>
            <Badge theme={ownerTypeTheme} size="md">
              {ownerTypeLabel}
            </Badge>
          </div>
          <div className="flex items-center gap-1.5">
            <Text variant="subtext" theme="neutral">Status:</Text>
            {getOverallStatusBadge(report)}
          </div>
        </div>
        <div className="flex items-center gap-3 shrink-0">
          <Time
            time={report.evaluated_at || ''}
            format="relative"
          />
          <PolicyReportPanelButton report={report} orgId={orgId} policyNameMap={policyNameMap} />
        </div>
      </div>

      {policies.length > 0 ? (
        <div className="divide-y divide-cool-grey-200 dark:divide-dark-grey-600">
          {policies.map((policy: TPolicyResult) => {
            const policyName =
              policy.policy_name ||
              (policy.policy_id && policyNameMap.get(policy.policy_id)) ||
              policy.policy_id

            return (
            <div
              key={policy.policy_id}
              className="flex items-center justify-between px-4 py-3"
            >
              <div className="flex flex-col gap-0.5 min-w-0">
                {policy.policy_id && report.app_id ? (
                  <Link
                    href={`/${orgId}/apps/${report.app_id}/policies/${policy.policy_id}`}
                    className="hover:underline"
                  >
                    <Text variant="body">
                      {policyName}
                    </Text>
                  </Link>
                ) : (
                  <Text variant="body">
                    {policyName}
                  </Text>
                )}
              </div>
              <div className="flex items-center gap-4 shrink-0">
                {(policy.deny_count || 0) > 0 && (
                  <Text
                    variant="subtext"
                    className="text-red-600 dark:text-red-400"
                  >
                    {policy.deny_count} {policy.deny_count === 1 ? 'deny' : 'denies'}
                  </Text>
                )}
                {(policy.warn_count || 0) > 0 && (
                  <Text
                    variant="subtext"
                    className="text-orange-600 dark:text-orange-400"
                  >
                    {policy.warn_count} {policy.warn_count === 1 ? 'warning' : 'warnings'}
                  </Text>
                )}
                {getPolicyStatusBadge(policy)}
              </div>
            </div>
            )
          })}
        </div>
      ) : (
        <div className="px-4 py-3">
          <Text variant="subtext" theme="neutral">
            No individual policy results available.
          </Text>
        </div>
      )}
    </Card>
  )
}
