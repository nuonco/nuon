import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { cn } from '@/utils/classnames'
import {
  getPolicyViolationCounts,
  type PolicyViolation,
} from '@/utils/workflow-utils'

interface IPolicyViolations {
  step: TWorkflowStep
}

const VISIBLE_WARNINGS = 3

const plural = (count: number, noun: string) =>
  `${count} ${noun}${count === 1 ? '' : 's'}`

export const policyReportTitle = (step: TWorkflowStep) => {
  const component = step?.metadata?.component_name
  if (component) return `${component} policy report`
  if (step?.name?.includes('sandbox')) return 'Sandbox policy report'
  return 'Policy report'
}

const ROW_THEMES = {
  deny: {
    className:
      'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400',
    icon: 'XCircleIcon',
    fallback: 'Policy check failed',
  },
  warn: {
    className:
      'bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400',
    icon: 'WarningIcon',
    fallback: 'Policy warning',
  },
  skip: {
    className:
      'bg-cool-grey-100 dark:bg-dark-grey-700 text-cool-grey-800 dark:text-cool-grey-400',
    icon: 'InfoIcon',
    fallback: 'Skipped by policy',
  },
} as const

const ViolationRow = ({
  violation,
  theme,
}: {
  violation: Pick<PolicyViolation, 'message'>
  theme: keyof typeof ROW_THEMES
}) => (
  <div
    className={cn(
      'flex items-start gap-2 rounded-md p-3',
      ROW_THEMES[theme].className
    )}
  >
    <Icon
      variant={ROW_THEMES[theme].icon}
      size={14}
      className="mt-0.5 shrink-0"
    />
    <Text variant="subtext">
      {violation.message || ROW_THEMES[theme].fallback}
    </Text>
  </div>
)

export const PolicyViolations = ({ step }: IPolicyViolations) => {
  const {
    denyViolations,
    warnViolations,
    skipFindings,
    denyCount,
    warnCount,
    hasPolicyData,
    hasViolations,
  } = getPolicyViolationCounts(step)

  if (!hasPolicyData || (!hasViolations && !skipFindings.length)) {
    return null
  }

  return (
    <Card className="!p-0 shrink-0 overflow-hidden">
      <Expand
        id={`policy-report-${step.id}`}
        headerClassName="!p-4 gap-4"
        heading={
          <div className="flex min-w-0 flex-1 items-center justify-between gap-4">
            <Text weight="strong" variant="body" className="truncate">
              {policyReportTitle(step)}
            </Text>
            <div className="flex shrink-0 items-center gap-3">
              {skipFindings.length > 0 ? (
                <Text variant="subtext" theme="neutral" weight="strong">
                  {denyCount > 0
                    ? 'Change denied by policy'
                    : step.status?.metadata?.check === 'policy-skip-apply'
                      ? 'Apply skipped by policy'
                      : 'Skip apply'}
                </Text>
              ) : null}
              {denyCount > 0 ? (
                <Text
                  variant="subtext"
                  weight="strong"
                  className="text-red-600 dark:text-red-500"
                >
                  {plural(denyCount, 'violation')}
                </Text>
              ) : null}
              {warnCount > 0 ? (
                <Text
                  variant="subtext"
                  weight="strong"
                  className="text-orange-600 dark:text-orange-500"
                >
                  {plural(warnCount, 'warning')}
                </Text>
              ) : null}
            </div>
          </div>
        }
      >
        <div className="flex flex-col gap-2 border-t p-4">
            {skipFindings.map((finding, index) => (
              <ViolationRow
                key={`skip-${finding.policy_id}-${index}`}
                violation={finding}
                theme="skip"
              />
            ))}
            {denyViolations.map((violation, index) => (
              <ViolationRow
                key={`deny-${violation.policy_id}-${index}`}
                violation={violation}
                theme="deny"
              />
            ))}
            {warnViolations
              .slice(0, VISIBLE_WARNINGS)
              .map((violation, index) => (
                <ViolationRow
                  key={`warn-${violation.policy_id}-${index}`}
                  violation={violation}
                  theme="warn"
                />
              ))}
            {warnViolations.length > VISIBLE_WARNINGS ? (
              <Expand
                id={`policy-warnings-${step.id}`}
                heading={
                  <Text variant="subtext" weight="strong">
                    {warnViolations.length - VISIBLE_WARNINGS} more{' '}
                    {warnViolations.length - VISIBLE_WARNINGS === 1
                      ? 'warning'
                      : 'warnings'}
                  </Text>
                }
                headerClassName="!px-1 !py-1"
              >
                <div className="flex flex-col gap-2 pb-1">
                  {warnViolations
                    .slice(VISIBLE_WARNINGS)
                    .map((violation, index) => (
                      <ViolationRow
                        key={`warn-more-${violation.policy_id}-${index}`}
                        violation={violation}
                        theme="warn"
                      />
                    ))}
                </div>
              </Expand>
            ) : null}
          </div>
      </Expand>
    </Card>
  )
}
