import { Card } from '@/components/common/Card'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { getPolicyViolationCounts } from '@/utils/workflow-utils'

interface IPolicyViolations {
  step: TWorkflowStep
}

export const PolicyViolations = ({ step }: IPolicyViolations) => {
  const {
    denyViolations,
    warnViolations,
    denyCount,
    hasPolicyData,
    hasViolations,
  } = getPolicyViolationCounts(step)

  if (!hasPolicyData) {
    return null
  }

  return (
    <Card className="!p-0 overflow-hidden">
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
            {!hasViolations ? (
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

        {hasViolations ? (
          <div className="flex flex-col gap-2 p-4">
            {denyViolations.map((violation, index) => (
              <div
                key={`deny-${violation.policy_id}-${index}`}
                className="flex items-start gap-2 rounded-md p-3 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400"
              >
                <Icon
                  variant="XCircle"
                  size={14}
                  className="mt-0.5 shrink-0"
                />
                <Text variant="subtext">
                  {violation.message || 'Policy check failed'}
                </Text>
              </div>
            ))}
            {warnViolations.map((violation, index) => (
              <div
                key={`warn-${violation.policy_id}-${index}`}
                className="flex items-start gap-2 rounded-md p-3 bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-400"
              >
                <Icon
                  variant="Warning"
                  size={14}
                  className="mt-0.5 shrink-0"
                />
                <Text variant="subtext">
                  {violation.message || 'Policy warning'}
                </Text>
              </div>
            ))}
          </div>
        ) : (
          <div className="p-4">
            <Text variant="subtext" theme="neutral">
              All policy checks passed successfully.
            </Text>
          </div>
        )}
      </div>
    </Card>
  )
}
