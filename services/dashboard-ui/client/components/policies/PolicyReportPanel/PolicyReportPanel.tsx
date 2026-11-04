import { Link } from '@/components/common/Link'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { reportSubject } from '@/components/policies/report-subject'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { panelTriggerClass } from '@/components/surfaces/panel-trigger'
import { useInstallLink } from '@/hooks/use-install-path'
import { cn } from '@/utils/classnames'
import type { TPolicyReport, TPolicyResult, TPolicyViolation } from '@/types'

function getPolicyStatusBadge(policy: TPolicyResult) {
  if (policy.status === 'skip_apply') {
    return (
      <Badge theme="neutral" size="sm">
        Skip apply
      </Badge>
    )
  }
  if (policy.status === 'deny') {
    return (
      <Badge theme="error" size="sm">
        <Icon variant="XCircleIcon" size={10} />
        Denied
      </Badge>
    )
  }
  if (policy.status === 'warn') {
    return (
      <Badge theme="warn" size="sm">
        <Icon variant="WarningIcon" size={10} />
        Warning
      </Badge>
    )
  }
  return (
    <Badge theme="success" size="sm">
      <Icon variant="CheckCircleIcon" size={10} />
      Passed
    </Badge>
  )
}

export interface IPolicyReportDetails {
  report: TPolicyReport
  orgId: string
  policyNameMap: Map<string, string>
}

export const PolicyReportDetails = ({
  report,
  orgId,
  policyNameMap,
}: IPolicyReportDetails) => {
  const installLink = useInstallLink()
  const isDeploy = report.owner_type === 'install_deploys'
  const isSandbox = report.owner_type === 'install_sandbox_runs'
  const reportLink = (suffix: string) =>
    report.install_id
      ? installLink({ orgId, installId: report.install_id, suffix })
      : undefined
  const subjectLink =
    isDeploy && report.component_id
      ? {
          href: reportLink(`/components/${report.component_id}`),
          label: 'View component',
        }
      : isSandbox
        ? { href: reportLink('/sandbox'), label: 'View sandbox' }
        : undefined
  const ownerLink =
    isDeploy && report.component_id && report.owner_id
      ? {
          href: reportLink(
            `/components/${report.component_id}/deploys/${report.owner_id}`
          ),
          label: 'View deploy',
        }
      : isSandbox && report.owner_id
        ? {
            href: reportLink(`/sandbox/runs/${report.owner_id}`),
            label: 'View sandbox run',
          }
        : undefined

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
  const skipFindings =
    (report.status?.metadata?.skip_findings as TPolicyViolation[]) || []

  const hasViolations = deniedPolicies.length > 0 || warnedPolicies.length > 0

  const denyViolations =
    report.violations?.filter((v: TPolicyViolation) => v.severity === 'deny') ||
    []
  const warnViolations =
    report.violations?.filter((v: TPolicyViolation) => v.severity === 'warn') ||
    []

  return (
    <>
      {skipFindings.length > 0 ? (
        <Banner theme={deniedPolicies.length ? 'error' : 'neutral'}>
          <Text variant="subtext">
            {report.status?.status_human_description ||
              (deniedPolicies.length
                ? 'Change denied by policy'
                : 'Policy decision: skip apply')}
          </Text>
        </Banner>
      ) : null}
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
            {deniedPolicies.length === 1 ? 'Policy' : 'Policies'} denied
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
            {warnedPolicies.length === 1 ? 'Policy' : 'Policies'} warning
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
            {passedPolicies.length === 1 ? 'Policy' : 'Policies'} passed
          </Text>
        </div>
      </div>

      <div className="flex flex-wrap items-start gap-x-8 gap-y-3">
        <LabeledValue label="Report ID">
          <ID>{report.id || ''}</ID>
        </LabeledValue>
        <LabeledValue label="Evaluated">
          <Time
            variant="subtext"
            time={report.evaluated_at || ''}
            format="relative"
          />
        </LabeledValue>
        <div className="ml-auto flex items-center gap-4">
          {subjectLink?.href ? (
            <Link href={subjectLink.href}>{subjectLink.label}</Link>
          ) : null}
          {ownerLink?.href ? (
            <Link href={ownerLink.href}>{ownerLink.label}</Link>
          ) : null}
        </div>
      </div>

      {report.policies && report.policies.length > 0 ? (
        <Card className="!p-0 overflow-hidden">
          <div className="flex flex-col">
            <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
              <Text weight="strong" variant="body">
                Policies evaluated
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
                const policySkipFindings = skipFindings.filter(
                  (finding) => finding.policy_id === policy.policy_id
                )

                const policyName =
                  policy.policy_name ||
                  (policy.policy_id && policyNameMap.get(policy.policy_id)) ||
                  policy.policy_id

                return (
                  <div key={policy.policy_id} className="flex flex-col">
                    <div className="flex items-center justify-between gap-4 p-4">
                      <div className="flex flex-col gap-1 min-w-0">
                        {policy.policy_id && report.app_id ? (
                          <Link
                            href={`/${orgId}/apps/${report.app_id}/policies/${policy.policy_id}`}
                            className="hover:underline truncate"
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
                        {policy.policy_id && <ID>{policy.policy_id}</ID>}
                      </div>
                      <div className="shrink-0">
                        {getPolicyStatusBadge(policy)}
                      </div>
                    </div>

                    {policySkipFindings.length > 0 ? (
                      <div className="flex flex-col gap-2 px-4 pb-4">
                        {policySkipFindings.map((finding, index) => (
                          <Text
                            key={`skip-${index}`}
                            variant="subtext"
                            theme="neutral"
                          >
                            {finding.message}
                          </Text>
                        ))}
                      </div>
                    ) : null}
                    {hasPolicyViolations && (
                      <div className="px-4 pb-4 space-y-2">
                        {[...policyDenyViolations, ...policyWarnViolations].map(
                          (violation: TPolicyViolation, index: number) => {
                            const isDeny = violation.severity === 'deny'
                            return (
                              <div
                                key={`${policy.policy_id}-${violation.severity}-${index}`}
                                className={cn(
                                  'flex items-start gap-2 rounded-md p-3',
                                  isDeny
                                    ? 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400'
                                    : 'bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400'
                                )}
                              >
                                <Icon
                                  variant={
                                    isDeny ? 'XCircleIcon' : 'WarningIcon'
                                  }
                                  size={14}
                                  className="mt-0.5 shrink-0"
                                />
                                <div className="flex flex-col gap-1 min-w-0">
                                  <Text variant="subtext">
                                    {violation.message ||
                                      (isDeny
                                        ? 'Policy check failed'
                                        : 'Policy warning')}
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
                          }
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
                      variant="XCircleIcon"
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
                      variant="WarningIcon"
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
            <Icon variant="CheckCircleIcon" size={16} />
            <Text variant="body">All policy checks passed.</Text>
          </div>
        </Card>
      )}
    </>
  )
}

export interface IPolicyReportPanel extends IPanel {
  report: TPolicyReport
  orgId: string
  policyNameMap: Map<string, string>
  history?: TPolicyReport[]
}

export const PolicyReportPanel = ({
  report,
  orgId,
  policyNameMap,
  history,
  ...props
}: IPolicyReportPanel) => (
  <Panel heading={reportSubject(report)} size="half" {...props}>
    <PolicyReportDetails
      report={report}
      orgId={orgId}
      policyNameMap={policyNameMap}
    />

    {history?.length ? (
      <>
        <Divider dividerWord="Evaluation history" />
        <div className="flex flex-col divide-y">
          {history.map((olderReport) => (
            <PolicyReportHistoryRow
              key={olderReport.id}
              report={olderReport}
              orgId={orgId}
              policyNameMap={policyNameMap}
            />
          ))}
        </div>
      </>
    ) : null}
  </Panel>
)

const PolicyReportHistoryRow = ({
  report,
  orgId,
  policyNameMap,
}: {
  report: TPolicyReport
  orgId: string
  policyNameMap: Map<string, string>
}) => {
  const denyCount = report?.deny_count ?? 0
  const warnCount = report?.warn_count ?? 0
  const skipCount = Number(report?.status?.metadata?.skip_count ?? 0)
  const skipsApply = report?.status?.metadata?.resolved_action === 'skip_apply'

  return (
    <div className="flex items-center justify-between gap-4 py-2">
      <span className="flex items-center gap-3">
        <Status variant="badge" status={report?.status?.status ?? 'unknown'} />
        <PolicyReportPanel
          report={report}
          orgId={orgId}
          policyNameMap={policyNameMap}
          panelKey={`policy-report-${report?.id}`}
          triggerButton={{
            variant: 'ghost',
            className: panelTriggerClass,
            children: (
              <Time
                variant="subtext"
                time={report?.evaluated_at || ''}
                format="relative"
              />
            ),
          }}
        />
      </span>
      <span className="flex items-center gap-2">
        {denyCount ? (
          <Badge size="sm" theme="error">
            {denyCount} denied
          </Badge>
        ) : null}
        {warnCount ? (
          <Badge size="sm" theme="warn">
            {warnCount} warning{warnCount === 1 ? '' : 's'}
          </Badge>
        ) : null}
        {skipCount > 0 ? (
          <Badge size="sm" theme="neutral">
            {skipsApply
              ? 'Skip apply'
              : `${skipCount} skip finding${skipCount === 1 ? '' : 's'}`}
          </Badge>
        ) : null}
        {!denyCount && !warnCount && !skipCount ? (
          <Text variant="subtext" theme="neutral">
            Passed
          </Text>
        ) : null}
      </span>
    </div>
  )
}
