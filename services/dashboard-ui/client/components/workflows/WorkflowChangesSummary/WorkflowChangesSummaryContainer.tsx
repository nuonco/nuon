import { PlanContainer } from '@/components/approvals/Plan/PlanContainer'
import { WorkflowInputChanges } from '@/components/workflows/WorkflowInputChanges'
import { useWorkflow } from '@/hooks/use-workflow'
import { useWorkflowChangeSummaries } from '@/hooks/use-workflow-change-summaries'
import type { TStepChangeSummary, TWorkflowStep } from '@/types'
import { WorkflowChangesSummary } from './WorkflowChangesSummary'

export const WorkflowChangesSummaryContainer = () => {
  const { workflow } = useWorkflow()
  const summaries = useWorkflowChangeSummaries(workflow)

  const stepsById = new Map<string, TWorkflowStep>(
    (workflow?.steps ?? []).map((step) => [step.id!, step])
  )

  const renderDetail = (summary: TStepChangeSummary) => {
    if (summary.planType === 'install_input') {
      return <WorkflowInputChanges workflow={workflow} />
    }
    const step = stepsById.get(summary.stepId)
    if (!step) return null
    return <PlanContainer step={step} />
  }

  return (
    <WorkflowChangesSummary summaries={summaries} renderDetail={renderDetail} />
  )
}
