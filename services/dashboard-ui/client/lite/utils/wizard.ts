import type { ReactNode } from 'react'

export type TWizardStepStatus = 'done' | 'current' | 'upcoming'

export interface IWizardStepContext<TState> {
  state: TState
  readOnly: boolean
  editing: boolean
}

export interface IWizardStep<TState> {
  id: string
  label: string
  complete: (state: TState) => boolean
  editable?: (state: TState) => boolean
  render: (context: IWizardStepContext<TState>) => ReactNode
}

export interface IWizardDescriptor<TState> {
  steps: IWizardStep<TState>[]
}

export interface IResolvedWizard<TState> {
  current?: IWizardStep<TState>
  currentIndex: number
  finished: boolean
  statuses: TWizardStepStatus[]
}

export const resolveWizardStep = <TState>(
  descriptor: IWizardDescriptor<TState>,
  state: TState
): IResolvedWizard<TState> => {
  const steps = descriptor.steps ?? []
  const currentIndex = steps.findIndex((step) => !step.complete(state))
  const finished = currentIndex === -1
  const statuses: TWizardStepStatus[] = steps.map((_, index) => {
    if (finished || index < currentIndex) return 'done'
    if (index === currentIndex) return 'current'
    return 'upcoming'
  })

  return {
    current: finished ? undefined : steps[currentIndex],
    currentIndex: finished ? steps.length : currentIndex,
    finished,
    statuses,
  }
}

export const isWizardComplete = <TState>(
  descriptor: IWizardDescriptor<TState>,
  state: TState
) => resolveWizardStep(descriptor, state).finished
