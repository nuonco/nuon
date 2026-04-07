export default {
  title: 'Onboarding/WizardNav',
}

import { WizardNav } from './WizardNav'

const steps = [
  { id: 'welcome', title: 'Welcome' },
  { id: 'app-profile', title: 'App profile' },
  { id: 'cloud-setup', title: 'Cloud setup' },
  { id: 'next-steps', title: 'Next steps' },
]

export const Default = () => (
  <div className="max-w-2xl p-4">
    <WizardNav
      steps={steps}
      currentStepIndex={0}
      completedSteps={new Set()}
      onboardingV2
      onGoToStep={() => {}}
    />
  </div>
)

export const MidProgress = () => (
  <div className="max-w-2xl p-4">
    <WizardNav
      steps={steps}
      currentStepIndex={2}
      completedSteps={new Set(['welcome', 'app-profile'])}
      onboardingV2
      onGoToStep={() => {}}
    />
  </div>
)
