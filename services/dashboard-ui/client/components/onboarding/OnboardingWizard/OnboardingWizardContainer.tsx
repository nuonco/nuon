import { useConfig } from '@/hooks/use-config'
import { useOnboardingJourney } from '@/hooks/use-onboarding-journey'
import {
  OnboardingWizardProvider,
  type IOnboardingWizardProps,
} from '@/providers/onboarding-wizard-provider'
import { OnboardingWizardLayout } from './OnboardingWizard'

function ConnectedWizardLayout() {
  const { onboardingV2 } = useConfig()

  if (onboardingV2) {
    return <OnboardingWizardLayout onboardingV2 skipHref={null} />
  }

  return <V1WizardLayout />
}

function V1WizardLayout() {
  const { isStepComplete, getStepMetadata } = useOnboardingJourney()
  const orgCreated = isStepComplete('org_created')
  const orgId = getStepMetadata('org_created', 'org_id') as string | undefined
  const skipHref = orgCreated && orgId ? `/${orgId}/apps` : null

  return (
    <OnboardingWizardLayout
      onboardingV2={false}
      skipHref={skipHref}
    />
  )
}

export function OnboardingWizardContainer(props: IOnboardingWizardProps) {
  return (
    <OnboardingWizardProvider {...props}>
      <ConnectedWizardLayout />
    </OnboardingWizardProvider>
  )
}
