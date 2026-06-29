'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TWorkflowStep } from '@/types'
import {
  getPolicyViolationCounts,
  type PolicyViolation,
} from '@/utils/workflow-utils'

interface IPolicyCountsBadge {
  step: TWorkflowStep
}

const ViolationList = ({
  heading,
  iconTheme,
  iconVariant,
  violations,
  fallback,
}: {
  heading: string
  iconTheme: 'error' | 'warn'
  iconVariant: 'XCircleIcon' | 'WarningIcon'
  violations: PolicyViolation[]
  fallback: string
}) => (
  <div className="flex flex-col gap-2 text-left pb-2">
    <Text weight="strong" flex>
      <Icon variant={iconVariant} theme={iconTheme} size={18} />
      {heading}
    </Text>
    <ul className="flex flex-col gap-2 list-disc pl-5">
      {violations.map((violation, index) => (
        <li key={`${violation.policy_id}-${index}`} className="leading-none">
          <Text variant="subtext">{violation.message || fallback}</Text>
        </li>
      ))}
    </ul>
  </div>
)

export const PolicyCountsBadge = ({ step }: IPolicyCountsBadge) => {
  const {
    denyCount,
    warnCount,
    denyViolations,
    warnViolations,
    hasPolicyData,
  } = getPolicyViolationCounts(step)

  if (!hasPolicyData) {
    return null
  }

  return (
    <div className="flex items-center gap-3">
      {denyCount > 0 ? (
        <Tooltip
          tipContentClassName="!whitespace-normal !w-96"
          tipContent={
            <ViolationList
              heading={
                denyCount === 1 ? 'Policy violation' : 'Policy violations'
              }
              iconTheme="error"
              iconVariant="XCircleIcon"
              violations={denyViolations}
              fallback="Policy check failed"
            />
          }
        >
          <Text
            variant="subtext"
            weight="strong"
            theme="error"
            className="cursor-help"
          >
            {denyCount} {denyCount === 1 ? 'violation' : 'violations'}
          </Text>
        </Tooltip>
      ) : null}
      {warnCount > 0 ? (
        <Tooltip
          tipContentClassName="!whitespace-normal !w-96"
          tipContent={
            <ViolationList
              heading={warnCount === 1 ? 'Policy warning' : 'Policy warnings'}
              iconTheme="warn"
              iconVariant="WarningIcon"
              violations={warnViolations}
              fallback="Policy warning"
            />
          }
        >
          <Text
            variant="subtext"
            weight="strong"
            theme="warn"
            className="cursor-help"
          >
            {warnCount} {warnCount === 1 ? 'warning' : 'warnings'}
          </Text>
        </Tooltip>
      ) : null}
    </div>
  )
}
