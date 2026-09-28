import { createContext, useState, useCallback } from 'react'
import type { ComponentType, ReactNode } from 'react'

export interface IWizardStepComponentProps {
  isComplete: boolean
  sharedData: Record<string, unknown>
  setSharedData: (key: string, val: unknown) => void
  onAdvance: () => void
  onGoBack?: () => void
  nextStepTitle?: string
}

export interface IWizardStepDef<TData = unknown> {
  id: string
  title: string
  navLabel?: string
  description?: ReactNode
  hideTitle?: boolean
  component: ComponentType<IWizardStepComponentProps>
  data?: TData
}

export interface IOnboardingWizardProps {
  steps: IWizardStepDef[]
  onComplete: () => void
  canClose?: boolean
  onClose?: () => void
  initialSharedData: Record<string, unknown>
  initialStepIndex: number
}

export interface IWizardContext {
  steps: IWizardStepDef[]
  currentStepIndex: number
  completedSteps: Set<string>
  sharedData: Record<string, unknown>
  canClose: boolean
  markComplete: (id: string) => void
  setSharedData: (key: string, val: unknown) => void
  goToStep: (index: number) => void
  goNext: () => void
  goPrev: () => void
  onComplete: () => void
  onClose?: () => void
}

export const WizardContext = createContext<IWizardContext | undefined>(undefined)

export function OnboardingWizardProvider({
  steps,
  onComplete,
  canClose = false,
  onClose,
  initialSharedData,
  initialStepIndex,
  children,
}: IOnboardingWizardProps & { children: React.ReactNode }) {
  const [currentStepIndex, setCurrentStepIndex] = useState(initialStepIndex)

  const [completedSteps, setCompletedSteps] = useState<Set<string>>(
    () => new Set(steps.slice(0, initialStepIndex).map((s) => s.id))
  )

  const [sharedData, setSharedDataState] = useState<Record<string, unknown>>(initialSharedData)

  const markComplete = useCallback((id: string) => {
    setCompletedSteps((prev) => new Set([...prev, id]))
  }, [])

  const setSharedData = useCallback((key: string, val: unknown) => {
    setSharedDataState((prev) => ({ ...prev, [key]: val }))
  }, [])

  const goToStep = useCallback(
    (index: number) => {
      if (index >= 0 && index < steps.length) {
        setCurrentStepIndex(index)
      }
    },
    [steps.length]
  )

  const goNext = useCallback(() => {
    setCurrentStepIndex((prev) => {
      const currentStepId = steps[prev]?.id
      if (currentStepId) {
        setCompletedSteps((s) => new Set([...s, currentStepId]))
      }
      if (prev < steps.length - 1) return prev + 1
      onComplete()
      return prev
    })
  }, [steps, onComplete])

  const goPrev = useCallback(() => {
    setCurrentStepIndex((prev) => Math.max(0, prev - 1))
  }, [])

  return (
    <WizardContext.Provider
      value={{
        steps,
        currentStepIndex,
        completedSteps,
        sharedData,
        canClose,
        markComplete,
        setSharedData,
        goToStep,
        goNext,
        goPrev,
        onComplete,
        onClose,
      }}
    >
      {children}
    </WizardContext.Provider>
  )
}
