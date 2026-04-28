import type { TComponent, TWorkflow, TWorkflowStep } from '@/types'

export type TDeployOutletContext = {
  component: TComponent
  workflow?: TWorkflow
  step: TWorkflowStep | null
}
