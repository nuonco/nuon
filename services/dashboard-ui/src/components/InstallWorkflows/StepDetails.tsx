import type { TInstallWorkflowStep } from '@/types'

export interface IPollStepDetails {
  pollDuration?: number
  shouldPoll?: boolean
  step: TInstallWorkflowStep
}
