'use client'

import { Notice } from '@/components/Notice'
import { InstallWorkflowSteps } from '@/components/InstallWorkflows/InstallWorkflowSteps'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TWorkflow } from '@/types'

interface IWorkflowSteps extends IPollingProps {
  initWorkflow: TWorkflow
  workflowId: string
}
export const WorkflowSteps = ({
  initWorkflow,
  pollInterval = 3000,
  shouldPoll = false,
  workflowId,
}: IWorkflowSteps) => {
  const { org } = useOrg()
  const { data: workflow, error } = usePolling<TWorkflow>({
    initData: initWorkflow,
    path: `/api/orgs/${org.id}/workflows/${workflowId}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <div>
      {error ? (
        <Notice className="!rounded-none !border-none">
          {error?.error || 'Unabled to load workflow steps'}
        </Notice>
      ) : null}
      <InstallWorkflowSteps installWorkflow={workflow} />
    </div>
  )
}
