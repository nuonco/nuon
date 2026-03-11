import { createContext, useState, useCallback, useEffect } from 'react'
import type { ComponentType } from 'react'

export interface IWizardStepComponentProps {
  isComplete: boolean
  sharedData: Record<string, unknown>
  setSharedData: (key: string, val: unknown) => void
  onAdvance: () => void
  nextStepTitle?: string
}

export interface IWizardStepDef<TData = unknown> {
  id: string
  title: string
  navLabel?: string
  description?: string
  component: ComponentType<IWizardStepComponentProps>
  data?: TData
}

export interface IOnboardingWizardProps {
  steps: IWizardStepDef[]
  onComplete: () => void
  canClose?: boolean
  onClose?: () => void
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

const STORAGE_KEY = 'onboarding-wizard-step'

export function OnboardingWizardProvider({
  steps,
  onComplete,
  canClose = false,
  onClose,
  children,
}: IOnboardingWizardProps & { children: React.ReactNode }) {
  const [currentStepIndex, setCurrentStepIndex] = useState(() => {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved === null) return 0
    const parsed = parseInt(saved, 10)
    return parsed >= 0 && parsed < steps.length ? parsed : 0
  })
  const [completedSteps, setCompletedSteps] = useState<Set<string>>(() => {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved === null) return new Set()
    const parsed = parseInt(saved, 10)
    return new Set(steps.slice(0, parsed).map((s) => s.id))
  })
  const [sharedData, setSharedDataState] = useState<Record<string, unknown>>({})

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, String(currentStepIndex))
  }, [currentStepIndex])

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
      if (prev < steps.length - 1) return prev + 1
      localStorage.removeItem(STORAGE_KEY)
      onComplete()
      return prev
    })
  }, [steps.length, onComplete])

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
