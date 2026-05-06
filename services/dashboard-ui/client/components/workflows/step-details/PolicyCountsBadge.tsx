'use client'

import { Text } from '@/components/common/Text'
import type { TWorkflowStep } from '@/types'
import { getPolicyViolationCounts } from '@/utils/workflow-utils'

interface IPolicyCountsBadge {
  step: TWorkflowStep
}

export const PolicyCountsBadge = ({ step }: IPolicyCountsBadge) => {
  const { denyCount, warnCount, hasPolicyData } =
    getPolicyViolationCounts(step)

  if (!hasPolicyData) {
    return null
  }

  return (
    <div className="flex items-center gap-3">
      {denyCount > 0 ? (
        <Text variant="subtext" className="text-red-600 dark:text-red-500">
          {denyCount} {denyCount === 1 ? 'violation' : 'violations'}
        </Text>
      ) : null}
      {warnCount > 0 ? (
        <Text
          variant="subtext"
          className="text-orange-600 dark:text-orange-500"
        >
          {warnCount} {warnCount === 1 ? 'warning' : 'warnings'}
        </Text>
      ) : null}
    </div>
  )
}
