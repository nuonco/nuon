import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import type { TWorkflowQueuePosition } from '@/lib/ctl-api/workflows/get-workflow-queue-position'
import { getStatusTheme } from '@/utils/status-utils'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import type { TWorkflow } from '@/types'

interface IWorkflowStatusSection {
  workflow: TWorkflow
  queuePosition?: TWorkflowQueuePosition
}

export const WorkflowStatusSection = ({
  workflow,
  queuePosition,
}: IWorkflowStatusSection) => {
  const signalsAhead = queuePosition?.signals_ahead ?? []

  return (
    <div className="flex flex-col gap-3 md:mt-6">
      <div className="flex flex-wrap items-center gap-2 md:gap-8">
        <Text
          variant="h3"
          weight="stronger"
          className="inline-flex gap-2 max-w-[600px]"
          theme={getStatusTheme(workflow.status.status) as any}
          title={toSentenceCase(
            workflow.status.status_human_description || workflow.status.status
          )}
        >
          <Status status={workflow.status.status} variant="timeline" />
          <span className="truncate">
            {toSentenceCase(
              workflow.status.status_human_description || workflow.status.status
            )}
          </span>
        </Text>

        <Text variant="h3" weight="stronger">
          Triggered via {snakeToWords(workflow.type)}
        </Text>
      </div>

      {signalsAhead.length > 0 && (
        <div className="flex flex-col gap-2">
          <Text variant="subtext">
            Position {queuePosition?.position ?? '?'} of{' '}
            {queuePosition?.queue_depth ?? '?'} in queue
            {' \u2014 '}{signalsAhead.length} workflow
            {signalsAhead.length !== 1 ? 's' : ''} ahead
          </Text>
          <div className="flex flex-wrap gap-1.5">
            {signalsAhead.map((item) => (
              <Badge key={item.workflow_id} variant="code" size="sm">
                {snakeToWords(item.workflow_type)}
              </Badge>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
