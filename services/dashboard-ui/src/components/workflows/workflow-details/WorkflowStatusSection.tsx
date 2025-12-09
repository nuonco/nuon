'use client'

import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { useWorkflow } from '@/hooks/use-workflow'
import { getStatusTheme } from '@/utils/status-utils'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'

export const WorkflowStatusSection = () => {
  const { workflow } = useWorkflow()
  return (
    <div className="flex flex-wrap items-center gap-2 md:gap-8 md:mt-6">
      <Text
        variant="h3"
        weight="stronger"
        className="inline-flex gap-2"
        theme={getStatusTheme(workflow.status.status) as any}
      >
        <Status status={workflow.status.status} variant="timeline" />
        {toSentenceCase(
          workflow.status.status_human_description || workflow.status.status
        )}
      </Text>

      <Text variant="h3" weight="stronger">
        Triggered via {snakeToWords(workflow.type)}
      </Text>
    </div>
  )
}