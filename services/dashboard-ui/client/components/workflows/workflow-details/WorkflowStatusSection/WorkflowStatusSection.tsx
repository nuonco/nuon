import { Badge } from '@/components/common/Badge'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TWorkflowQueueItem, TWorkflowQueuePosition } from '@/lib/ctl-api/workflows/get-workflow-queue-position'
import { getStatusTheme } from '@/utils/status-utils'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import type { TWorkflow } from '@/types'
import { Link } from '@/components/common/Link'

interface IWorkflowStatusSection {
  workflow: TWorkflow
  queuePosition?: TWorkflowQueuePosition
  installId?: string
  orgId?: string
  onCancelWorkflow?: (workflowId: string) => void
}

export const WorkflowStatusSection = ({
  workflow,
  queuePosition,
  installId,
  orgId,
  onCancelWorkflow,
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
        <Expand
          id="queue-position"
          className="border rounded-lg"
          heading={
            <div className="flex items-center gap-2">
              <Icon variant="StackIcon" size="16" />
              <Text weight="strong">
                Position {queuePosition?.position ?? '?'} of{' '}
                {queuePosition?.queue_depth ?? '?'} in queue
              </Text>
              <Badge theme="neutral" size="sm">
                {signalsAhead.length} workflow{signalsAhead.length !== 1 ? 's' : ''} ahead
              </Badge>
            </div>
          }
        >
          <div className="flex flex-col divide-y">
            {signalsAhead.map((item, index) => (
              <QueueItem
                key={item.workflow_id}
                item={item}
                position={index + 1}
                installId={installId}
                orgId={orgId}
                onCancel={onCancelWorkflow}
              />
            ))}
          </div>
        </Expand>
      )}
    </div>
  )
}

const QueueItem = ({
  item,
  position,
  installId,
  orgId,
  onCancel,
}: {
  item: TWorkflowQueueItem
  position: number
  installId?: string
  orgId?: string
  onCancel?: (workflowId: string) => void
}) => {
  const workflowLink =
    orgId && installId
      ? `/${orgId}/installs/${installId}/workflows/${item.workflow_id}`
      : undefined

  return (
    <div className="flex items-center justify-between gap-3 px-3 py-2">
      <div className="flex items-center gap-3 min-w-0">
        <Text variant="label" theme="neutral" className="w-6 text-right shrink-0">
          {position}
        </Text>
        <Status status={item.status} isWithoutText />
        <Badge variant="code" size="sm">
          {snakeToWords(item.workflow_type)}
        </Badge>
        {item.created_at && (
          <Time variant="subtext" time={item.created_at} format="relative" />
        )}
      </div>

      <div className="flex items-center gap-2 shrink-0">
        {workflowLink && (
          <Link
            href={workflowLink}
            className="text-xs hover:underline flex items-center gap-1 opacity-60 hover:opacity-100 transition-opacity"
          >
            <Icon variant="ArrowSquareOutIcon" size="14" />
            <span>View</span>
          </Link>
        )}
        {onCancel && (
          <button
            type="button"
            onClick={() => onCancel(item.workflow_id)}
            className="text-xs text-red-500 hover:text-red-400 flex items-center gap-1 opacity-60 hover:opacity-100 transition-opacity cursor-pointer"
          >
            <Icon variant="StopCircleIcon" size="14" />
            <span>Cancel</span>
          </button>
        )}
      </div>
    </div>
  )
}
