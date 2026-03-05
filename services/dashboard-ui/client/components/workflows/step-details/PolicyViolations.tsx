import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Expand } from '@/components/common/Expand'
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
    warnCount,
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
            Policy Evaluation
          </Text>
          <div className="flex items-center gap-2">
            {denyCount > 0 ? (
              <Badge theme="error" size="sm">
                <Icon variant="XCircle" size={10} />
                {denyCount} {denyCount === 1 ? 'Deny' : 'Denies'}
              </Badge>
            ) : null}
            {warnCount > 0 ? (
              <Badge theme="warn" size="sm">
                <Icon variant="Warning" size={10} />
                {warnCount} {warnCount === 1 ? 'Warning' : 'Warnings'}
              </Badge>
            ) : null}
            {!hasViolations ? (
              <Badge theme="success" size="sm">
                <Icon variant="CheckCircle" size={10} />
                Passed
              </Badge>
            ) : null}
          </div>
        </div>

        {hasViolations ? (
          <div className="flex flex-col">
            {denyCount > 0 ? (
              <Expand
                id={`policy-denies-${step.id}`}
                heading={
                  <div className="flex items-center gap-2">
                    <Icon
                      variant="XCircle"
                      size={14}
                      className="text-red-600 dark:text-red-500"
                    />
                    <Text variant="subtext" weight="strong">
                      {denyCount} Policy{' '}
                      {denyCount === 1 ? 'Violation' : 'Violations'}
                    </Text>
                  </div>
                }
                className="border-b border-cool-grey-200 dark:border-dark-grey-600 last:border-b-0"
                headerClassName="!p-3"
              >
                <ul className="px-4 pb-3 space-y-2">
                  {denyViolations.map((violation, index) => (
                    <li
                      key={`deny-${violation.policy_id}-${index}`}
                      className="flex items-start gap-2 text-red-700 dark:text-red-400"
                    >
                      <Icon
                        variant="CaretRight"
                        size={12}
                        className="mt-1 shrink-0"
                      />
                      <Text variant="subtext">
                        {violation.message || 'Policy check failed'}
                      </Text>
                    </li>
                  ))}
                </ul>
              </Expand>
            ) : null}

            {warnCount > 0 ? (
              <Expand
                id={`policy-warnings-${step.id}`}
                heading={
                  <div className="flex items-center gap-2">
                    <Icon
                      variant="Warning"
                      size={14}
                      className="text-orange-600 dark:text-orange-500"
                    />
                    <Text variant="subtext" weight="strong">
                      {warnCount} Policy{' '}
                      {warnCount === 1 ? 'Warning' : 'Warnings'}
                    </Text>
                  </div>
                }
                className="border-b border-cool-grey-200 dark:border-dark-grey-600 last:border-b-0"
                headerClassName="!p-3"
              >
                <ul className="px-4 pb-3 space-y-2">
                  {warnViolations.map((violation, index) => (
                    <li
                      key={`warn-${violation.policy_id}-${index}`}
                      className="flex items-start gap-2 text-orange-700 dark:text-orange-400"
                    >
                      <Icon
                        variant="CaretRight"
                        size={12}
                        className="mt-1 shrink-0"
                      />
                      <Text variant="subtext">
                        {violation.message || 'Policy warning'}
                      </Text>
                    </li>
                  ))}
                </ul>
              </Expand>
            ) : null}
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
