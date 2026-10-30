import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import {
  getPolicyViolationCounts,
  type PolicyViolation,
} from '@/utils/workflow-utils'

interface IPolicyViolations {
  step: TWorkflowStep
}

const VISIBLE_WARNINGS = 3

const ViolationRow = ({
  violation,
  theme,
}: {
  violation: PolicyViolation
  theme: 'deny' | 'warn'
}) => (
  <div
    className={
      theme === 'deny'
        ? 'flex items-start gap-2 rounded-md p-3 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400'
        : 'flex items-start gap-2 rounded-md p-3 bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400'
    }
  >
    <Icon
      variant={theme === 'deny' ? 'XCircleIcon' : 'WarningIcon'}
      size={14}
      className="mt-0.5 shrink-0"
    />
    <Text variant="subtext">
      {violation.message ||
        (theme === 'deny' ? 'Policy check failed' : 'Policy warning')}
    </Text>
  </div>
)

export const PolicyViolations = ({ step }: IPolicyViolations) => {
  const {
    denyViolations,
    warnViolations,
    skipFindings,
    denyCount,
    hasPolicyData,
    hasViolations,
  } = getPolicyViolationCounts(step)

  if (!hasPolicyData) {
    return null
  }

  return (
    <Card className="!p-0 shrink-0 overflow-hidden">
      <div className="flex flex-col">
        <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
          <Text weight="strong" variant="body">
            Policy report
          </Text>
          <div className="flex items-center gap-3">
            {denyCount > 0 ? (
              <Text
                variant="subtext"
                weight="strong"
                className="text-red-600 dark:text-red-500"
              >
                {denyCount} {denyCount === 1 ? 'violation' : 'violations'}
              </Text>
            ) : null}
            {skipFindings.length > 0 ? (
              <Text variant="subtext" theme="neutral" weight="strong">
                {denyCount > 0
                  ? 'Change denied by policy'
                  : step.status?.metadata?.check === 'policy-skip-apply'
                    ? 'Apply skipped by policy'
                    : 'Skip apply'}
              </Text>
            ) : null}
            {!hasViolations && !skipFindings.length ? (
              <Text
                variant="subtext"
                weight="strong"
                className="text-green-600 dark:text-green-500"
              >
                All passed
              </Text>
            ) : null}
          </div>
        </div>

        {hasViolations || skipFindings.length > 0 ? (
          <div className="flex flex-col gap-2 p-4">
            {skipFindings.map((finding, index) => (
              <Banner
                key={`skip-${finding.policy_id}-${index}`}
                theme="neutral"
              >
                <Text variant="subtext">{finding.message}</Text>
              </Banner>
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
        ) : (
          <div className="p-4">
            <Text variant="subtext" theme="neutral">
              All policy checks passed.
            </Text>
          </div>
        )}
      </div>
    </Card>
  )
}
