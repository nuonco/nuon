import { useEffect } from 'react'
import { useAuth } from '@/hooks/use-auth'
import { FIRST_RUN_STEPS, type TFirstRunStep } from '@/hooks/use-first-run-journey'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import { isCloud, type TCloud, type TPath } from './constants'
import { writeFirstRunSession } from './session'

const STEP_NAMES = new Set<string>(FIRST_RUN_STEPS.map((step) => step.name))

const isPath = (value: unknown): value is TPath => value === 'own' || value === 'example'

export function FirstRunProgress() {
  const { user, isLoading } = useAuth()
  const { steps, currentStepIndex, sharedData } = useOnboardingWizard()
  const step = steps[currentStepIndex]?.id

  useEffect(() => {
    if (isLoading || !step || !STEP_NAMES.has(step)) return
    const path: TPath = isPath(sharedData.path) ? sharedData.path : 'example'
    const cloud: TCloud = isCloud(sharedData.cloud) ? sharedData.cloud : 'aws'
    writeFirstRunSession(user?.sub, {
      started: true,
      step: step as TFirstRunStep,
      path,
      cloud,
      sharedData,
    })
  }, [isLoading, user?.sub, step, sharedData])

  return null
}
