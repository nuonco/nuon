import { useConfig } from '@/hooks/use-config'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import { useOnboardingWizard } from '@/hooks/use-onboarding-wizard'
import {
  OnboardingWizardProvider,
  type IOnboardingWizardProps,
} from '@/providers/onboarding-wizard-provider'
import type { TOnboarding } from '@/types/ctl-api.types'
import { OnboardingWizardLayout } from './OnboardingWizard'

function ConnectedWizardLayout() {
  const { onboardingV2 } = useConfig()

  if (onboardingV2) {
    return <V2WizardLayout />
  }

  return <V1WizardLayout />
}

function V1WizardLayout() {
  const { orgId } = useOnboardingJourney()
  const skipHref = orgId ? `/${orgId}/apps` : null

  return <OnboardingWizardLayout skipHref={skipHref} />
}

function V2WizardLayout() {
  const { sharedData } = useOnboardingWizard()
  const orgId = (sharedData.onboarding as TOnboarding | undefined)?.org_id
  const skipHref = orgId ? `/${orgId}` : null

  return <OnboardingWizardLayout skipHref={skipHref} />
}

export function OnboardingWizardContainer(props: IOnboardingWizardProps) {
  return (
    <OnboardingWizardProvider {...props}>
      <ConnectedWizardLayout />
    </OnboardingWizardProvider>
  )
}
