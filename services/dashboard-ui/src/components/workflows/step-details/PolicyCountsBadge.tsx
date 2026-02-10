'use client'

import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import type { TWorkflowStep } from '@/types'
import { getPolicyViolationCounts } from '@/utils/workflow-utils'

interface IPolicyCountsBadge {
  step: TWorkflowStep
}

export const PolicyCountsBadge = ({ step }: IPolicyCountsBadge) => {
  const { denyCount, warnCount, hasPolicyData, hasViolations } =
    getPolicyViolationCounts(step)

  if (!hasPolicyData) {
    return null
  }

  return (
    <div className="flex items-center gap-1">
      {denyCount > 0 ? (
        <Badge theme="error" size="sm">
          <Icon variant="XCircle" size={10} />
          {denyCount}
        </Badge>
      ) : null}
      {warnCount > 0 ? (
        <Badge theme="warn" size="sm">
          <Icon variant="Warning" size={10} />
          {warnCount}
        </Badge>
      ) : null}
      {!hasViolations ? (
        <Badge theme="success" size="sm">
          <Icon variant="CheckCircle" size={10} />
          Passed
        </Badge>
      ) : null}
    </div>
  )
}
