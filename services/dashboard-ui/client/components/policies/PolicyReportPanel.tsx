import { Link } from '@/components/common/Link'
import { Badge } from '@/components/common/Badge'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TPolicyReport, TPolicyResult, TPolicyViolation } from '@/types'

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
      <Badge theme="error" size="sm">
        <Icon variant="XCircle" size={10} />
        Denied
      </Badge>
    )
  }
  if ((report.warn_count || 0) > 0) {
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

interface IPolicyReportPanel extends IPanel {
  report: TPolicyReport
  orgId: string
}

export const PolicyReportPanel = ({
  report,
  orgId,
  ...props
}: IPolicyReportPanel) => {
  const { label: ownerTypeLabel, theme: ownerTypeTheme } = formatOwnerType(
    report.owner_type || ''
  )

  const policies = report.policies || []
  const deniedPolicies = policies.filter(
    (p: TPolicyResult) => p.status === 'deny'
  )
  const warnedPolicies = policies.filter(
    (p: TPolicyResult) => p.status === 'warn'
  )
  const passedPolicies = policies.filter(
    (p: TPolicyResult) => p.status === 'pass'
  )

  const hasViolations = deniedPolicies.length > 0 || warnedPolicies.length > 0

  const denyViolations =
    report.violations?.filter((v: TPolicyViolation) => v.severity === 'deny') ||
    []
  const warnViolations =
    report.violations?.filter((v: TPolicyViolation) => v.severity === 'warn') ||
    []

  return (
    <Panel heading={report.component_name || 'Sandbox'} size="half" {...props}>
      <Card className="!p-0 overflow-hidden">
        <div className="flex flex-col">
          <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
            <Text weight="strong" variant="body">
              Report Details
            </Text>
            {getOverallStatusBadge(report)}
          </div>

          <div className="p-4 space-y-3">
            <div className="flex items-center justify-between">
              <Text variant="subtext" theme="neutral">
                Report ID
              </Text>
              <ID>{report.id || ''}</ID>
            </div>
            <div className="flex items-center justify-between">
              <Text variant="subtext" theme="neutral">
                Type
              </Text>
              <Badge theme={ownerTypeTheme} size="sm">
                {ownerTypeLabel}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <Text variant="subtext" theme="neutral">
                Evaluated
              </Text>
              <Time time={report.evaluated_at || ''} format="relative" />
            </div>
          </div>
        </div>
      </Card>

      <Card className="!p-0 overflow-hidden">
        <div className="flex flex-col">
          <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
            <Text weight="strong" variant="body">
              Summary
            </Text>
          </div>

          <div className="p-4">
            <div className="grid grid-cols-3 gap-4 text-center">
              <div className="flex flex-col gap-1 p-3 rounded-lg bg-red-50 dark:bg-red-900/20">
                <Text
                  variant="base"
                  weight="strong"
                  className="text-red-600 dark:text-red-400"
                >
                  {deniedPolicies.length}
                </Text>
                <Text variant="subtext" theme="neutral">
                  {deniedPolicies.length === 1 ? 'Policy' : 'Policies'} Denied
                </Text>
              </div>
              <div className="flex flex-col gap-1 p-3 rounded-lg bg-orange-50 dark:bg-orange-900/20">
                <Text
                  variant="base"
                  weight="strong"
                  className="text-orange-600 dark:text-orange-400"
                >
                  {warnedPolicies.length}
                </Text>
                <Text variant="subtext" theme="neutral">
                  {warnedPolicies.length === 1 ? 'Policy' : 'Policies'} Warning
                </Text>
              </div>
              <div className="flex flex-col gap-1 p-3 rounded-lg bg-green-50 dark:bg-green-900/20">
                <Text
                  variant="base"
                  weight="strong"
                  className="text-green-600 dark:text-green-400"
                >
                  {passedPolicies.length}
                </Text>
                <Text variant="subtext" theme="neutral">
                  {passedPolicies.length === 1 ? 'Policy' : 'Policies'} Passed
                </Text>
              </div>
            </div>
          </div>
        </div>
      </Card>

      {report.policies && report.policies.length > 0 ? (
        <Card className="!p-0 overflow-hidden">
          <div className="flex flex-col">
            <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
              <Text weight="strong" variant="body">
                Policies Evaluated
              </Text>
              <Text variant="subtext" theme="neutral">
                {report.policies.length}{' '}
                {report.policies.length === 1 ? 'policy' : 'policies'}
              </Text>
            </div>

            <div className="divide-y divide-cool-grey-200 dark:divide-dark-grey-600">
              {policies.map((policy: TPolicyResult) => {
                const policyDenyViolations =
                  report.violations?.filter(
                    (v: TPolicyViolation) =>
                      v.policy_id === policy.policy_id && v.severity === 'deny'
                  ) || []
                const policyWarnViolations =
                  report.violations?.filter(
                    (v: TPolicyViolation) =>
                      v.policy_id === policy.policy_id && v.severity === 'warn'
                  ) || []
                const hasPolicyViolations =
                  policyDenyViolations.length > 0 ||
                  policyWarnViolations.length > 0

                return (
                  <div key={policy.policy_id} className="flex flex-col">
                    <div className="flex items-center justify-between p-4">
                      <div className="flex flex-col gap-1">
                        {policy.policy_id && report.app_id ? (
                          <Link
                            href={`/${orgId}/apps/${report.app_id}/policies/${policy.policy_id}`}
                            className="hover:underline"
                          >
                            <Text variant="body">
                              {policy.policy_name || policy.policy_id}
                            </Text>
                          </Link>
                        ) : (
                          <Text variant="body">
                            {policy.policy_name || policy.policy_id}
                          </Text>
                        )}
                        {policy.policy_id && <ID>{policy.policy_id}</ID>}
                      </div>
                      {getPolicyStatusBadge(policy)}
                    </div>

                    {hasPolicyViolations && (
                      <div className="px-4 pb-4 space-y-2">
                        {policyDenyViolations.length > 0 && (
                          <Expand
                            id={`policy-denies-${policy.policy_id}`}
                            heading={
                              <div className="flex items-center gap-2">
                                <Icon
                                  variant="XCircle"
                                  size={14}
                                  className="text-red-600 dark:text-red-500"
                                />
                                <Text variant="subtext" weight="strong">
                                  {policyDenyViolations.length}{' '}
                                  {policyDenyViolations.length === 1
                                    ? 'Deny'
                                    : 'Denies'}
                                </Text>
                              </div>
                            }
                            className="border border-cool-grey-200 dark:border-dark-grey-600 rounded-lg"
                            headerClassName="!p-3"
                          >
                            <div className="px-4 pb-3 space-y-2">
                              {policyDenyViolations.map(
                                (
                                  violation: TPolicyViolation,
                                  index: number
                                ) => (
                                  <div
                                    key={`deny-${policy.policy_id}-${index}`}
                                    className="flex items-start gap-2 text-red-700 dark:text-red-400"
                                  >
                                    <Icon
                                      variant="CaretRight"
                                      size={12}
                                      className="mt-1 shrink-0"
                                    />
                                    <div className="flex flex-col gap-1">
                                      <Text variant="subtext">
                                        {violation.message ||
                                          'Policy check failed'}
                                      </Text>
                                      {violation.input_identity && (
                                        <Text
                                          variant="subtext"
                                          theme="neutral"
                                          className="text-xs"
                                        >
                                          Input: {violation.input_identity}
                                        </Text>
                                      )}
                                    </div>
                                  </div>
                                )
                              )}
                            </div>
                          </Expand>
                        )}
                        {policyWarnViolations.length > 0 && (
                          <Expand
                            id={`policy-warnings-${policy.policy_id}`}
                            heading={
                              <div className="flex items-center gap-2">
                                <Icon
                                  variant="Warning"
                                  size={14}
                                  className="text-orange-600 dark:text-orange-500"
                                />
                                <Text variant="subtext" weight="strong">
                                  {policyWarnViolations.length}{' '}
                                  {policyWarnViolations.length === 1
                                    ? 'Warning'
                                    : 'Warnings'}
                                </Text>
                              </div>
                            }
                            className="border border-cool-grey-200 dark:border-dark-grey-600 rounded-lg"
                            headerClassName="!p-3"
                          >
                            <div className="px-4 pb-3 space-y-2">
                              {policyWarnViolations.map(
                                (
                                  violation: TPolicyViolation,
                                  index: number
                                ) => (
                                  <div
                                    key={`warn-${policy.policy_id}-${index}`}
                                    className="flex items-start gap-2 text-orange-700 dark:text-orange-400"
                                  >
                                    <Icon
                                      variant="CaretRight"
                                      size={12}
                                      className="mt-1 shrink-0"
                                    />
                                    <div className="flex flex-col gap-1">
                                      <Text variant="subtext">
                                        {violation.message || 'Policy warning'}
                                      </Text>
                                      {violation.input_identity && (
                                        <Text
                                          variant="subtext"
                                          theme="neutral"
                                          className="text-xs"
                                        >
                                          Input: {violation.input_identity}
                                        </Text>
                                      )}
                                    </div>
                                  </div>
                                )
                              )}
                            </div>
                          </Expand>
                        )}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        </Card>
      ) : hasViolations ? (
        <Card className="!p-0 overflow-hidden">
          <div className="flex flex-col">
            <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
              <Text weight="strong" variant="body">
                Violations
              </Text>
            </div>
            <div className="p-4 space-y-2">
              {denyViolations.map(
                (violation: TPolicyViolation, index: number) => (
                  <div
                    key={`deny-${violation.policy_id}-${index}`}
                    className="flex items-start gap-2 text-red-700 dark:text-red-400"
                  >
                    <Icon
                      variant="XCircle"
                      size={12}
                      className="mt-1 shrink-0"
                    />
                    <Text variant="subtext">
                      {violation.message || 'Policy check failed'}
                    </Text>
                  </div>
                )
              )}
              {warnViolations.map(
                (violation: TPolicyViolation, index: number) => (
                  <div
                    key={`warn-${violation.policy_id}-${index}`}
                    className="flex items-start gap-2 text-orange-700 dark:text-orange-400"
                  >
                    <Icon
                      variant="Warning"
                      size={12}
                      className="mt-1 shrink-0"
                    />
                    <Text variant="subtext">
                      {violation.message || 'Policy warning'}
                    </Text>
                  </div>
                )
              )}
            </div>
          </div>
        </Card>
      ) : (
        <Card>
          <div className="flex items-center gap-2 text-green-600 dark:text-green-400">
            <Icon variant="CheckCircle" size={16} />
            <Text variant="body">All policy checks passed successfully.</Text>
          </div>
        </Card>
      )}
    </Panel>
  )
}

interface IPolicyReportPanelButton extends IButtonAsButton {
  report: TPolicyReport
  orgId: string
}

export const PolicyReportPanelButton = ({
  report,
  orgId,
  ...props
}: IPolicyReportPanelButton) => {
  const { addPanel } = useSurfaces()
  const panel = <PolicyReportPanel report={report} orgId={orgId} />

  return (
    <Button
      variant="ghost"
      className="!p-1"
      onClick={() => addPanel(panel)}
      title="View details"
      aria-label="View policy report details"
      {...props}
    >
      <Icon variant="ArrowRight" size={16} />
    </Button>
  )
}
