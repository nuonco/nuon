import type { TWorkflow, TWorkflowStep } from '@/types'

export type TSandboxRunOutletContext = {
  workflow?: TWorkflow
  step: TWorkflowStep | null
}
