import { Button } from '@/components/common/Button'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useWorkflowChangeSummaries } from '@/hooks/use-workflow-change-summaries'
import { WorkflowProvider } from '@/providers/workflow-provider'
import type { TWorkflow } from '@/types'
import { WorkflowChangesSummaryContainer } from './WorkflowChangesSummaryContainer'

export const CHANGES_PANEL_KEY = 'workflow-changes'

interface IWorkflowChangesTrigger {
  workflow?: TWorkflow
  label?: string
  className?: string
}

export const WorkflowChangesTrigger = ({
  workflow,
  label = 'Change summary',
  className,
}: IWorkflowChangesTrigger) => {
  const { addPanel } = useSurfaces()
  const summaries = useWorkflowChangeSummaries(workflow)

  if (!workflow?.id || summaries.length === 0) return null

  const panel = (
    <Panel
      panelKey={CHANGES_PANEL_KEY}
      heading="Workflow changes"
      size="3/4"
    >
      <WorkflowProvider workflowId={workflow.id}>
        <WorkflowChangesSummaryContainer />
      </WorkflowProvider>
    </Panel>
  )

  return (
    <Button onClick={() => addPanel(panel, CHANGES_PANEL_KEY)} className={className}>
      {label}
    </Button>
  )
}
