export default {
  title: 'Features / Workflows / Install workflow panel',
}

import { PanelStory } from '@/components/__stories__/helpers'
import { InstallContext } from '@/providers/install-provider'
import { OrgContext } from '@/providers/org-provider'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TWorkflow } from '@/types'
import {
  InstallWorkflowPanel,
  InstallWorkflowPanelContent,
} from './InstallWorkflowPanel'

const workflow = {
  id: 'wf-1',
  name: 'Restart workers',
  type: 'action_workflow_run',
  created_at: '2026-09-30T13:00:00Z',
  updated_at: '2026-09-30T13:02:00Z',
  status: {
    status: 'success',
    status_human_description: 'Action completed successfully',
    metadata: { all_steps_loaded: true },
  },
  steps: [],
} as TWorkflow

export const Default = () => (
  <OrgContext.Provider
    value={{ org: { id: 'org-1', name: 'Acme' } as any, refresh: () => {} }}
  >
    <InstallContext.Provider
      value={{
        install: {
          id: 'inst-1',
          name: 'payments',
          app_id: 'app-1',
        } as any,
        labelColors: {},
        refresh: () => {},
      }}
    >
      <WorkflowContext.Provider
        value={{
          workflow,
          stopPolling: () => {},
          workflowSteps: [],
          hasApprovals: false,
          failedSteps: [],
          pendingApprovals: [],
          discardedSteps: [],
          completedSteps: [],
          stepsWithPolicyViolations: [],
          totalSteps: 0,
          pendingApprovalsCount: 0,
          discardedStepsCount: 0,
          completedStepsCount: 0,
          failedStepsCount: 0,
          policyViolationsCount: 0,
        }}
      >
        <PanelStory label="Open workflow details">
          <InstallWorkflowPanel>
            <InstallWorkflowPanelContent />
          </InstallWorkflowPanel>
        </PanelStory>
      </WorkflowContext.Provider>
    </InstallContext.Provider>
  </OrgContext.Provider>
)
