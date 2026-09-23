import { useWorkflowChangeSummaries } from '@/hooks/use-workflow-change-summaries'
import type { TWorkflow } from '@/types'
import { ChangesAggregate } from './ChangesAggregate'
import { WorkflowChangesTrigger } from './WorkflowChangesTrigger'

interface IWorkflowChangesLink {
  workflow?: TWorkflow
  className?: string
}

export const WorkflowChangesLink = ({
  workflow,
  className,
}: IWorkflowChangesLink) => {
  const summaries = useWorkflowChangeSummaries(workflow)

  if (!workflow?.id || summaries.length === 0) return null

  return (
    <div className={className}>
      <div className="flex items-center gap-2">
        <ChangesAggregate summaries={summaries} />
        <WorkflowChangesTrigger workflow={workflow} label="View all changes" />
      </div>
    </div>
  )
}
