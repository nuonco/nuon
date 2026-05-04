import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PolicyReportPanelButton } from '@/components/policies/PolicyReportPanel'
import { cn } from '@/utils/classnames'
import type { TPolicyReport, TPolicyResult, TPolicyViolation } from '@/types'

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

interface IPolicyRow {
  policy: TPolicyResult
  policyName: string
  orgId: string
  appId?: string
  violations: TPolicyViolation[]
}

function PolicyRow({
  policy,
  policyName,
  orgId,
  appId,
  violations,
}: IPolicyRow) {
  const [isExpanded, setIsExpanded] = useState(false)
  const hasViolations = violations.length > 0

  const toggle = () => {
    if (hasViolations) setIsExpanded((p) => !p)
  }

  return (
    <div className="flex flex-col">
      <div
        role={hasViolations ? 'button' : undefined}
        tabIndex={hasViolations ? 0 : undefined}
        aria-expanded={hasViolations ? isExpanded : undefined}
        onClick={toggle}
        onKeyDown={(e) => {
          if (!hasViolations) return
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            toggle()
          }
        }}
        className={cn(
          'flex items-center justify-between gap-4 px-4 py-3 outline-none',
          hasViolations && 'cursor-pointer'
        )}
      >
        <div className="flex items-center gap-3 min-w-0">
          <div className="min-w-0 truncate">
            {policy.policy_id && appId ? (
              <Link
                href={`/${orgId}/apps/${appId}/policies/${policy.policy_id}`}
                className="hover:underline"
                onClick={(e) => e.stopPropagation()}
              >
                <Text variant="body" className="truncate">
                  {policyName}
                </Text>
              </Link>
            ) : (
              <Text variant="body" className="truncate">
                {policyName}
              </Text>
            )}
          </div>
          <div className="shrink-0">{getPolicyStatusBadge(policy)}</div>
        </div>

        {hasViolations ? (
          <Icon
            variant={isExpanded ? 'CaretUp' : 'CaretDown'}
            size={16}
            className="shrink-0"
          />
        ) : null}
      </div>

      {isExpanded && hasViolations ? (
        <div className="px-4 pb-3 space-y-2">
          {violations.map((violation, index) => {
            const isDeny = violation.severity === 'deny'
            return (
              <div
                key={`${policy.policy_id}-${index}`}
                className={cn(
                  'flex items-start gap-2 rounded-md p-3',
                  isDeny
                    ? 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400'
                    : 'bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400'
                )}
              >
                <Icon
                  variant={isDeny ? 'XCircle' : 'Warning'}
                  size={14}
                  className="mt-0.5 shrink-0"
                />
                <div className="flex flex-col gap-1 min-w-0">
                  <Text variant="subtext">
                    {violation.message ||
                      (isDeny ? 'Policy check failed' : 'Policy warning')}
                  </Text>
                  {violation.input_identity ? (
                    <Text
                      variant="subtext"
                      theme="neutral"
                      className="text-xs"
                    >
                      Input: {violation.input_identity}
                    </Text>
                  ) : null}
                </div>
              </div>
            )
          })}
        </div>
      ) : null}
    </div>
  )
}

interface IPolicyReportGroup {
  report: TPolicyReport
  orgId: string
  policyNameMap: Map<string, string>
  variant?: 'card' | 'embedded'
}

export function PolicyReportGroup({
  report,
  orgId,
  policyNameMap,
  variant = 'card',
}: IPolicyReportGroup) {
  const policies = report.policies || []
  const allViolations = report.violations || []
  const isEmbedded = variant === 'embedded'

  const content = (
    <>
      <div
        className={cn(
          'flex items-center justify-between p-4',
          policies.length > 0 &&
            'border-b border-cool-grey-200 dark:border-dark-grey-600'
        )}
      >
        <div className="flex items-center gap-6 min-w-0">
          <div className="flex flex-col gap-0.5 min-w-0">
            <Text weight="strong" variant="body" className="truncate">
              {report.component_name
                ? `Component - ${report.component_name}`
                : 'Sandbox'}
            </Text>
            {!isEmbedded ? <ID>{report.id || ''}</ID> : null}
          </div>
        </div>
        <div className="flex items-center gap-3 shrink-0">
          <Time
            variant="subtext"
            theme="neutral"
            time={report.evaluated_at || ''}
            format="relative"
          />
          <PolicyReportPanelButton
            report={report}
            orgId={orgId}
            policyNameMap={policyNameMap}
          />
        </div>
      </div>

      {policies.length > 0 ? (
        <div className="divide-y divide-cool-grey-200 dark:divide-dark-grey-600">
          {policies.map((policy: TPolicyResult) => {
            const policyName =
              policy.policy_name ||
              (policy.policy_id && policyNameMap.get(policy.policy_id)) ||
              policy.policy_id ||
              'Unknown policy'
            const policyViolations = allViolations.filter(
              (v) => v.policy_id === policy.policy_id
            )
            return (
              <PolicyRow
                key={policy.policy_id}
                policy={policy}
                policyName={policyName}
                orgId={orgId}
                appId={report.app_id}
                violations={policyViolations}
              />
            )
          })}
        </div>
      ) : null}
    </>
  )

  if (isEmbedded) {
    return <div className="flex flex-col">{content}</div>
  }

  return (
    <Card className="!p-0 !gap-0 overflow-hidden">{content}</Card>
  )
}
