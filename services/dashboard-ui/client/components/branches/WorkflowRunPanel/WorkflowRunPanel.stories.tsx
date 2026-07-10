export default {
  title: 'Branches/WorkflowRunPanel',
}

import { SurfacesProvider } from '@/providers/surfaces-provider'
import { PanelStory } from '@/components/__stories__/helpers'
import type { TInstallWorkflowStep } from '@/types'
import { WorkflowRunPanel } from './WorkflowRunPanel'

const steps: TInstallWorkflowStep[] = [
  {
    id: 'step-1',
    name: 'commit changes',
    status: { status: 'success' },
    execution_time: 4200,
  } as TInstallWorkflowStep,
  {
    id: 'step-2',
    name: 'sync app config',
    status: { status: 'in-progress' },
    execution_time: 12000,
  } as TInstallWorkflowStep,
  {
    id: 'step-3',
    name: 'deploy install group production',
    status: { status: 'pending' },
  } as TInstallWorkflowStep,
]

export const Default = () => (
  <SurfacesProvider>
    <PanelStory label="Open workflow panel">
      <WorkflowRunPanel
        runTitle="Manual update"
        status="in-progress"
        steps={steps}
        selectedStep={null}
        activeStep={steps[1]}
        onSelectStep={() => {}}
        onJumpToActive={() => {}}
        onClose={() => {}}
      />
    </PanelStory>
  </SurfacesProvider>
)

export const Loading = () => (
  <SurfacesProvider>
    <PanelStory label="Open loading panel">
      <WorkflowRunPanel
        runTitle="Workflow run"
        status="unknown"
        steps={[]}
        selectedStep={null}
        onSelectStep={() => {}}
        onJumpToActive={() => {}}
        onClose={() => {}}
        isLoading
      />
    </PanelStory>
  </SurfacesProvider>
)
