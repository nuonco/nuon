import { Code } from '@/components/common/Code'
import { Expand } from '@/components/common/Expand'
import { JSONViewer } from '@/components/common/JSONViewer'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TWorkflow } from '@/types'

const WorkflowHistoryStatus = ({
  status,
  id,
}: {
  status: NonNullable<NonNullable<TWorkflow['status']>['history']>[number]
  id: string
}) => {
  const description = status.status_human_description

  if (!description) {
    return (
      <span className="flex items-center gap-4 py-2">
        <Status status={status.status} variant="badge" />
        <Time seconds={status.created_at_ts} variant="subtext" theme="neutral" />
      </span>
    )
  }

  return (
    <Expand
      id={id}
      hasNoHoverStyle
      headerClassName="!p-0"
      heading={
        <span className="flex items-center gap-4 py-2">
          <Status status={status.status} variant="badge" />
          <Time seconds={status.created_at_ts} variant="subtext" theme="neutral" />
        </span>
      }
    >
      <Code className="mb-2 !text-xs">{description}</Code>
    </Expand>
  )
}

interface IWorkflowMetadata {
  workflow: TWorkflow
}

export const WorkflowMetadata = ({ workflow }: IWorkflowMetadata) => {
  return (
    <div className="flex flex-col gap-2 p-4 border-t">
      <Expand
        className="border rounded-md"
        id="workflow-status-history"
        heading={
          <Text family="mono" variant="subtext">
            View status history
          </Text>
        }
      >
        <div className="border-t flex flex-col p-4 divide-y">
          {workflow?.status?.history?.map((status, idx) => (
            <WorkflowHistoryStatus
              key={`${status.created_at_ts}-${idx}`}
              status={status}
              id={`workflow-history-${idx}`}
            />
          ))}
          {workflow?.status ? (
            <WorkflowHistoryStatus status={workflow.status} id="workflow-history-current" />
          ) : null}
        </div>
      </Expand>

      <Expand
        className="border rounded-md"
        id="workflow-json"
        heading={
          <Text family="mono" variant="subtext">
            View workflow JSON
          </Text>
        }
      >
        <div className="border-t p-4">
          <JSONViewer data={workflow} expanded={1} />
        </div>
      </Expand>
    </div>
  )
}
