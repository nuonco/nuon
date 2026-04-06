export default {
  title: 'Onboarding/OnboardingWizard',
}

import { OnboardingWizardLayout } from './OnboardingWizard'

export const Default = () => (
  <OnboardingWizardLayout
    onboardingV2={false}
    skipHref="/org-123/apps"
  />
)

export const V2 = () => (
  <OnboardingWizardLayout
    onboardingV2={true}
    skipHref={null}
  />
)

export const NoSkip = () => (
  <OnboardingWizardLayout
    onboardingV2={false}
    skipHref={null}
  />
)
