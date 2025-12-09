'use client'

import { BackLink } from '@/components/common/BackLink'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { useWorkflow } from '@/hooks/use-workflow'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { WorkflowActionButtons } from './WorkflowActionButtons'

export const WorkflowHeader = () => {
  const { workflow } = useWorkflow()
  return (
    <div className="flex flex-wrap items-center gap-3 justify-between w-full">
      <div className="flex flex-col gap-4">
        <BackLink />
        <HeadingGroup>
          <Text
            className="inline-flex gap-2 items-center"
            variant="h3"
            weight="strong"
          >
            {workflow.name || toSentenceCase(snakeToWords(workflow.type))}
          </Text>
          <Text theme="neutral">
            Watch your app get updated here and provide needed approvals.
          </Text>
        </HeadingGroup>
      </div>

      <div className="flex flex-col gap-4">
        <WorkflowActionButtons />
      </div>
    </div>
  )
}
