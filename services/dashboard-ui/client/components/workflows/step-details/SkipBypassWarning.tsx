'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TWorkflowStep } from '@/types'

export const SkipBypassWarning = ({
  ranAnyway,
}: {
  ranAnyway: TWorkflowStep
}) => (
  <Tooltip
    tipContentClassName="!whitespace-normal !w-96"
    tipContent={
      <div className="flex flex-col gap-2 text-left pb-2">
        <Text weight="strong" flex>
          <Icon variant="WarningIcon" theme="warn" size={18} />
          Skip didn&apos;t stop the group
        </Text>
        <Text variant="subtext">
          {`This step was skipped, but "${ranAnyway.name}" in the same group still ran. Skipping a plan doesn't always discard the steps after it, so any changes that step made were applied. Review its output before trusting this result.`}
        </Text>
      </div>
    }
  >
    <Text variant="subtext" weight="strong" theme="warn" className="cursor-help">
      Later step still ran
    </Text>
  </Tooltip>
)
