import { Link } from '@/components/common/Link'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useWorkflow } from '@/hooks/use-workflow'
import { useWorkflowChangeSummaries } from '@/hooks/use-workflow-change-summaries'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { ChangesAggregate } from './ChangesAggregate'
import { WorkflowChangesSummaryContainer } from './WorkflowChangesSummaryContainer'

const PANEL_KEY = 'workflow-changes'

export const WorkflowChangesLink = () => {
  const { workflow } = useWorkflow()
  const { addPanel } = useSurfaces()
  const summaries = useWorkflowChangeSummaries(workflow)

  if (summaries.length === 0) return null

  const panel = (
    <Panel panelKey={PANEL_KEY} heading="Workflow changes" size="3/4">
      <WorkflowProvider workflowId={workflow!.id!}>
        <WorkflowChangesSummaryContainer />
      </WorkflowProvider>
    </Panel>
  )

  return (
    <div className="flex items-center gap-2">
      <ChangesAggregate summaries={summaries} />
      <Link isATag onClick={() => addPanel(panel, PANEL_KEY)} className="cursor-pointer">
        View all changes
      </Link>
    </div>
  )
}
